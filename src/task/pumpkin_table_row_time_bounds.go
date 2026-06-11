/*
Copyright 2026 The Dbmeta Team Group, website: https://www.dbmeta.com
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
*/

package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ruyi1024/dbmeta/log"

	"go.uber.org/zap"
)

var pumpkinSafeSQLIdent = regexp.MustCompile(`^[a-zA-Z0-9_$]+$`)

func pumpkinQuoteMySQLIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func pumpkinQuotePGIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func pumpkinSanitizeSQLIdent(name string) (string, bool) {
	s := strings.TrimSpace(name)
	if s == "" || !pumpkinSafeSQLIdent.MatchString(s) {
		return "", false
	}
	return s, true
}

// queryPumpkinTableRowTimeBoundsFromConn 根据业务表中时间类型列的 MIN/MAX 推断「首条有业务时间的记录」与「最近一条有业务时间的记录」对应的时间边界。
// 优先 MySQL 系（含 TiDB/Doris/OceanBase 协议兼容）；PostgreSQL 支持 schema.table 形式表名。
// ctx 用于取消/超时，避免单表全表扫描导致整任务长时间挂起；ctx 可为 nil（等价于 context.Background）。
func queryPumpkinTableRowTimeBoundsFromConn(ctx context.Context, db *sql.DB, datasourceType, databaseName, tableName string) (minAt, maxAt *time.Time, err error) {
	if db == nil {
		return nil, nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	kind := pumpkinCapacityQueryKind(datasourceType)
	switch kind {
	case "mysqlWithCreateTime", "mysqlNoCreateTime":
		return mysqlLikePumpkinTableRowTimeBounds(ctx, db, databaseName, tableName)
	case "postgresql":
		return postgresPumpkinTableRowTimeBounds(ctx, db, tableName)
	default:
		return nil, nil, nil
	}
}

func mysqlLikePumpkinTableRowTimeBounds(ctx context.Context, db *sql.DB, databaseName, tableName string) (minAt, maxAt *time.Time, err error) {
	dbn, ok1 := pumpkinSanitizeSQLIdent(databaseName)
	tbn, ok2 := pumpkinSanitizeSQLIdent(tableName)
	if !ok1 || !ok2 {
		return nil, nil, nil
	}
	rows, err := db.QueryContext(ctx,
		`SELECT COLUMN_NAME, DATA_TYPE FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		   AND DATA_TYPE IN ('datetime','timestamp','date')
		 ORDER BY ORDINAL_POSITION`,
		dbn, tbn,
	)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, context.DeadlineExceeded
		}
		return nil, nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var colName, dataType string
		if scanErr := rows.Scan(&colName, &dataType); scanErr != nil {
			continue
		}
		if c, ok := pumpkinSanitizeSQLIdent(colName); ok {
			cols = append(cols, c)
		}
	}
	if err := rows.Err(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, context.DeadlineExceeded
		}
		return nil, nil, err
	}
	if len(cols) == 0 {
		return nil, nil, nil
	}
	minCol, maxCol := pickPumpkinMinMaxTimeColumns(cols)
	if minCol == "" {
		return nil, nil, nil
	}
	if maxCol == "" {
		maxCol = minCol
	}
	q := fmt.Sprintf(
		"SELECT MIN(%s), MAX(%s) FROM %s.%s",
		pumpkinQuoteMySQLIdent(minCol),
		pumpkinQuoteMySQLIdent(maxCol),
		pumpkinQuoteMySQLIdent(dbn),
		pumpkinQuoteMySQLIdent(tbn),
	)
	var vmin, vmax interface{}
	if err := db.QueryRowContext(ctx, q).Scan(&vmin, &vmax); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, context.DeadlineExceeded
		}
		log.Logger.Debug("表行时间 MIN/MAX 查询失败",
			zap.String("db", dbn), zap.String("table", tbn), zap.Error(err))
		return nil, nil, nil
	}
	return parsePumpkinOptionalTime(vmin), parsePumpkinOptionalTime(vmax), nil
}

func postgresPumpkinTableRowTimeBounds(ctx context.Context, db *sql.DB, tableName string) (minAt, maxAt *time.Time, err error) {
	schema, rel, ok := splitPumpkinPGTableName(tableName)
	if !ok {
		return nil, nil, nil
	}
	sch, ok1 := pumpkinSanitizeSQLIdent(schema)
	tbl, ok2 := pumpkinSanitizeSQLIdent(rel)
	if !ok1 || !ok2 {
		return nil, nil, nil
	}
	rows, err := db.QueryContext(ctx,
		`SELECT column_name, data_type FROM information_schema.columns
		 WHERE table_schema = $1 AND table_name = $2
		   AND (data_type LIKE '%timestamp%' OR data_type = 'date')
		 ORDER BY ordinal_position`,
		sch, tbl,
	)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, context.DeadlineExceeded
		}
		return nil, nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var colName, dataType string
		if scanErr := rows.Scan(&colName, &dataType); scanErr != nil {
			continue
		}
		if c, ok := pumpkinSanitizeSQLIdent(colName); ok {
			cols = append(cols, c)
		}
	}
	if err := rows.Err(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, context.DeadlineExceeded
		}
		return nil, nil, err
	}
	if len(cols) == 0 {
		return nil, nil, nil
	}
	minCol, maxCol := pickPumpkinMinMaxTimeColumns(cols)
	if minCol == "" {
		return nil, nil, nil
	}
	if maxCol == "" {
		maxCol = minCol
	}
	q := fmt.Sprintf(
		`SELECT MIN(%s), MAX(%s) FROM %s.%s`,
		pumpkinQuotePGIdent(minCol),
		pumpkinQuotePGIdent(maxCol),
		pumpkinQuotePGIdent(sch),
		pumpkinQuotePGIdent(tbl),
	)
	var vmin, vmax interface{}
	if err := db.QueryRowContext(ctx, q).Scan(&vmin, &vmax); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, context.DeadlineExceeded
		}
		log.Logger.Debug("PostgreSQL 表行时间 MIN/MAX 查询失败",
			zap.String("schema", sch), zap.String("table", tbl), zap.Error(err))
		return nil, nil, nil
	}
	return parsePumpkinOptionalTime(vmin), parsePumpkinOptionalTime(vmax), nil
}

// splitPumpkinPGTableName 解析 pumpkin 采集的 PG 表名：一般为 schema.rel；无 schema 时默认 public。
func splitPumpkinPGTableName(tableName string) (schema, rel string, ok bool) {
	tn := strings.TrimSpace(tableName)
	if tn == "" {
		return "", "", false
	}
	if strings.Contains(tn, ".") {
		parts := strings.SplitN(tn, ".", 2)
		s0, s1 := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if s0 != "" && s1 != "" {
			return s0, s1, true
		}
	}
	return "public", tn, true
}

// pickPumpkinMinMaxTimeColumns 在已按序排列的时间列中，挑选用于 MIN（偏创建语义）与 MAX（偏更新语义）的列名。
func pickPumpkinMinMaxTimeColumns(cols []string) (minCol, maxCol string) {
	createCandidates := []string{
		"gmt_create", "create_time", "created_at", "insert_time", "add_time",
		"cre_time", "ctime", "create_date", "rec_create_time", "first_insert_time",
	}
	updateCandidates := []string{
		"gmt_modified", "update_time", "updated_at", "modify_time", "mtime",
		"edit_time", "utime", "last_update", "update_date", "modified_time", "last_modified",
	}
	minCol = findPumpkinTimeColumnByCandidates(cols, createCandidates)
	maxCol = findPumpkinTimeColumnByCandidates(cols, updateCandidates)
	if minCol == "" {
		minCol = cols[0]
	}
	if maxCol == "" {
		maxCol = minCol
	}
	return minCol, maxCol
}

func findPumpkinTimeColumnByCandidates(cols []string, candidates []string) string {
	for _, cand := range candidates {
		for _, c := range cols {
			if strings.EqualFold(c, cand) {
				return c
			}
		}
	}
	suffixes := make([]string, 0, len(candidates))
	for _, cand := range candidates {
		suffixes = append(suffixes, "_"+cand)
	}
	for _, c := range cols {
		lc := strings.ToLower(c)
		for _, suf := range suffixes {
			if strings.HasSuffix(lc, suf) {
				return c
			}
		}
	}
	return ""
}

func pumpkinMaxTimePtr(cands ...*time.Time) *time.Time {
	var best *time.Time
	for _, t := range cands {
		if t == nil || t.IsZero() {
			continue
		}
		if best == nil || t.After(*best) {
			tt := *t
			best = &tt
		}
	}
	return best
}
