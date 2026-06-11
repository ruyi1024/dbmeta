/*
Copyright 2026 The Dbmeta Team Group, website: https://www.dbmeta.com
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
You may not use this file except in compliance with the License.
*/

package task

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"github.com/ruyi1024/dbmeta/src/utils"

	"go.uber.org/zap"
)

// pumpkinRemoteRowToPumpkinRecord 将远程查询的一行转为 PumpkinTableSize 及生命周期元数据；ok=false 表示跳过该行。
func pumpkinRemoteRowToPumpkinRecord(datasourceType, host, port string, item map[string]interface{}) (model.PumpkinTableSize, *pumpkinTableGatherMeta, bool) {
	if item["database_name"] == nil || item["table_name"] == nil {
		log.Logger.Warn("跳过无效记录：缺少必要字段", zap.Any("item", item))
		return model.PumpkinTableSize{}, nil, false
	}
	databaseName := formatPumpkinInterface(item["database_name"])
	tableName := formatPumpkinInterface(item["table_name"])
	if databaseName == "" || tableName == "" {
		log.Logger.Warn("跳过无效记录：字段为空", zap.String("database_name", databaseName), zap.String("table_name", tableName))
		return model.PumpkinTableSize{}, nil, false
	}
	var dataSize int64
	if item["data_size"] != nil {
		dataSize = utils.StrToInt64(formatPumpkinInterface(item["data_size"]))
	}
	var indexSize int64
	if item["index_size"] != nil {
		indexSize = utils.StrToInt64(formatPumpkinInterface(item["index_size"]))
	}
	var freeSize int64
	if item["free_size"] != nil {
		freeSize = utils.StrToInt64(formatPumpkinInterface(item["free_size"]))
	}
	var tableRows int64
	if item["table_rows"] != nil {
		tableRows = utils.StrToInt64(formatPumpkinInterface(item["table_rows"]))
	}
	var avgRowLength int64
	if item["avg_row_length"] != nil {
		avgRowLength = utils.StrToInt64(formatPumpkinInterface(item["avg_row_length"]))
	}
	record := model.PumpkinTableSize{
		DatasourceType: datasourceType,
		Host:           host,
		Port:           port,
		DatabaseName:   databaseName,
		TableNameField: tableName,
		DataSize:       dataSize,
		IndexSize:      indexSize,
		FreeSize:       freeSize,
		TableRows:      tableRows,
		AvgRowLength:   avgRowLength,
	}
	return record, metaFromPumpkinRow(item), true
}

// pumpkinCapacityQueryKind 统一数据源 type 写法（大小写、别名），用于选择南瓜表级 SQL 变体。
func pumpkinCapacityQueryKind(datasourceType string) string {
	t := strings.TrimSpace(datasourceType)
	l := strings.ToLower(t)
	switch l {
	case "mysql", "mariadb", "greatsql", "tidb":
		return "mysqlWithCreateTime"
	case "doris", "oceanbase":
		return "mysqlNoCreateTime"
	case "postgresql", "postgres":
		return "postgresql"
	case "clickhouse":
		return "clickhouse"
	case "sqlserver", "mssql":
		return "sqlserver"
	}
	switch t {
	case "MySQL", "MariaDB", "GreatSQL", "TiDB":
		return "mysqlWithCreateTime"
	case "Doris", "OceanBase":
		return "mysqlNoCreateTime"
	case "PostgreSQL":
		return "postgresql"
	case "ClickHouse":
		return "clickhouse"
	case "SQLServer":
		return "sqlserver"
	case "达梦数据库":
		return "dameng"
	}
	return ""
}

// pumpkinBuildRemoteTableSizeSQL 构建南瓜表级容量查询 SQL（单条或达梦多条备选）。
func pumpkinBuildRemoteTableSizeSQL(datasourceType string) (queryTableSizeSql string, queryTableSizeSqlList []string, err error) {
	switch pumpkinCapacityQueryKind(datasourceType) {
	case "mysqlWithCreateTime":
		queryTableSizeSql = `
			SELECT 
				table_schema as database_name,
				table_name as table_name,
				data_length as data_size,
				index_length as index_size,
				data_free as free_size,
				table_rows as table_rows,
				avg_row_length as avg_row_length,
				create_time as table_create_time,
				update_time as table_update_time
			FROM information_schema.tables 
			WHERE table_schema NOT IN ('information_schema', 'performance_schema', 'sys', 'mysql', 'metrics_schema', '__internal_schema', 'sys_audit', 'lbacsys', 'oceanbase', 'ocs', 'oraauditor')
			AND table_type='BASE TABLE'
		`
	case "mysqlNoCreateTime":
		// Doris / OceanBase 等 information_schema 可能无 create_time/update_time 或与 MySQL 语义不一致，避免整段 SQL 失败导致零行
		queryTableSizeSql = `
			SELECT 
				table_schema as database_name,
				table_name as table_name,
				data_length as data_size,
				index_length as index_size,
				data_free as free_size,
				table_rows as table_rows,
				avg_row_length as avg_row_length,
				CAST(NULL AS DATETIME) as table_create_time,
				CAST(NULL AS DATETIME) as table_update_time
			FROM information_schema.tables 
			WHERE table_schema NOT IN ('information_schema', 'performance_schema', 'sys', 'mysql', 'metrics_schema', '__internal_schema', 'sys_audit', 'lbacsys', 'oceanbase', 'ocs', 'oraauditor')
			AND table_type='BASE TABLE'
		`
	case "postgresql":
		queryTableSizeSql = `
			SELECT
				lower(current_database()) as database_name,
				lower(concat(n.nspname, '.', c.relname)) as table_name,
				(pg_total_relation_size(c.oid) - pg_indexes_size(c.oid))::bigint as data_size,
				pg_indexes_size(c.oid)::bigint as index_size,
				0::bigint as free_size,
				coalesce(s.n_live_tup, 0)::bigint as table_rows,
				CASE
					WHEN coalesce(s.n_live_tup, 0) > 0 THEN (pg_total_relation_size(c.oid) / s.n_live_tup)::bigint
					ELSE 0::bigint
				END as avg_row_length,
				NULL::timestamp as table_create_time,
				NULL::timestamp as table_update_time
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			LEFT JOIN pg_stat_user_tables s ON s.relid = c.oid
			WHERE c.relkind IN ('r', 'm')
			AND n.nspname NOT IN ('pg_catalog', 'information_schema')
			AND n.nspname NOT LIKE 'pg_toast%'
			AND n.nspname NOT LIKE 'pg_temp_%'
		`
	case "clickhouse":
		queryTableSizeSql = `
			SELECT
				lower(database) as database_name,
				lower(table) as table_name,
				sum(data_compressed_bytes) as data_size,
				0 as index_size,
				0 as free_size,
				sum(rows) as table_rows,
				CASE
					WHEN sum(rows) > 0 THEN toInt64(sum(data_uncompressed_bytes) / sum(rows))
					ELSE 0
				END as avg_row_length,
				CAST(NULL AS Nullable(DateTime)) as table_create_time,
				CAST(NULL AS Nullable(DateTime)) as table_update_time
			FROM system.parts
			WHERE active = 1
			AND lower(database) NOT IN ('information_schema', 'system')
			GROUP BY database, table
		`
	case "sqlserver":
		queryTableSizeSql = `
			SELECT
				lower(DB_NAME()) as database_name,
				lower(t.name) as table_name,
				SUM(CASE WHEN i.index_id < 2 THEN a.data_pages ELSE 0 END) * 8 * 1024 as data_size,
				(SUM(a.used_pages) - SUM(CASE WHEN i.index_id < 2 THEN a.data_pages ELSE 0 END)) * 8 * 1024 as index_size,
				(SUM(a.total_pages) - SUM(a.used_pages)) * 8 * 1024 as free_size,
				SUM(CASE WHEN i.index_id < 2 THEN p.rows ELSE 0 END) as table_rows,
				CASE
					WHEN SUM(CASE WHEN i.index_id < 2 THEN p.rows ELSE 0 END) > 0
					THEN (SUM(CASE WHEN i.index_id < 2 THEN a.data_pages ELSE 0 END) * 8 * 1024)
						/ SUM(CASE WHEN i.index_id < 2 THEN p.rows ELSE 0 END)
					ELSE 0
				END as avg_row_length,
				MIN(t.create_date) as table_create_time,
				MAX(t.modify_date) as table_update_time
			FROM sys.tables t
			JOIN sys.indexes i ON t.object_id = i.object_id
			JOIN sys.partitions p ON i.object_id = p.object_id AND i.index_id = p.index_id
			JOIN sys.allocation_units a ON p.partition_id = a.container_id
			WHERE t.is_ms_shipped = 0
			GROUP BY t.name
		`
	case "dameng":
		queryTableSizeSqlList = []string{
			`
			SELECT
				t.owner as database_name,
				t.table_name as table_name,
				COALESCE(ds.data_size, 0) as data_size,
				COALESCE(idx.index_size, 0) as index_size,
				0 as free_size,
				COALESCE(t.num_rows, 0) as table_rows,
				CASE
					WHEN COALESCE(t.num_rows, 0) > 0 THEN (COALESCE(ds.data_size, 0) + COALESCE(idx.index_size, 0)) / t.num_rows
					ELSE 0
				END as avg_row_length,
				NULL as table_create_time,
				NULL as table_update_time
			FROM all_tables t
			LEFT JOIN (
				SELECT owner, segment_name as table_name, SUM(bytes) as data_size
				FROM dba_segments
				WHERE segment_type IN ('TABLE', 'TABLE PARTITION', 'TABLE SUBPARTITION')
				GROUP BY owner, segment_name
			) ds
				ON ds.owner = t.owner AND ds.table_name = t.table_name
			LEFT JOIN (
				SELECT i.table_owner as owner, i.table_name, SUM(s.bytes) as index_size
				FROM all_indexes i
				JOIN dba_segments s ON s.owner = i.owner AND s.segment_name = i.index_name
				GROUP BY i.table_owner, i.table_name
			) idx
				ON idx.owner = t.owner AND idx.table_name = t.table_name
			WHERE t.owner NOT IN ('SYS', 'SYSTEM', 'SYSAUDITOR')
			`,
			`
			SELECT
				t.owner as database_name,
				t.table_name as table_name,
				COALESCE(ds.data_size, 0) as data_size,
				0 as index_size,
				0 as free_size,
				COALESCE(t.num_rows, 0) as table_rows,
				CASE
					WHEN COALESCE(t.num_rows, 0) > 0 THEN COALESCE(ds.data_size, 0) / t.num_rows
					ELSE 0
				END as avg_row_length,
				NULL as table_create_time,
				NULL as table_update_time
			FROM all_tables t
			LEFT JOIN (
				SELECT owner, segment_name as table_name, SUM(bytes) as data_size
				FROM all_segments
				WHERE segment_type IN ('TABLE', 'TABLE PARTITION', 'TABLE SUBPARTITION')
				GROUP BY owner, segment_name
			) ds
				ON ds.owner = t.owner AND ds.table_name = t.table_name
			WHERE t.owner NOT IN ('SYS', 'SYSTEM', 'SYSAUDITOR')
			`,
			`
			SELECT
				USER as database_name,
				t.table_name as table_name,
				COALESCE(ds.data_size, 0) as data_size,
				COALESCE(idx.index_size, 0) as index_size,
				0 as free_size,
				COALESCE(t.num_rows, 0) as table_rows,
				CASE
					WHEN COALESCE(t.num_rows, 0) > 0 THEN (COALESCE(ds.data_size, 0) + COALESCE(idx.index_size, 0)) / t.num_rows
					ELSE 0
				END as avg_row_length,
				NULL as table_create_time,
				NULL as table_update_time
			FROM user_tables t
			LEFT JOIN (
				SELECT segment_name as table_name, SUM(bytes) as data_size
				FROM user_segments
				WHERE segment_type IN ('TABLE', 'TABLE PARTITION', 'TABLE SUBPARTITION')
				GROUP BY segment_name
			) ds
				ON ds.table_name = t.table_name
			LEFT JOIN (
				SELECT i.table_name, SUM(s.bytes) as index_size
				FROM user_indexes i
				JOIN user_segments s ON s.segment_name = i.index_name
				GROUP BY i.table_name
			) idx
				ON idx.table_name = t.table_name
			WHERE 1=1
			`,
		}
	default:
		return "", nil, fmt.Errorf("不支持的数据库类型: %s", datasourceType)
	}
	return queryTableSizeSql, queryTableSizeSqlList, nil
}

// execPumpkinRemoteTableSizeQuery 在已打开的远程连接上执行南瓜表级容量查询。
func execPumpkinRemoteTableSizeQuery(db *sql.DB, datasourceType string) ([]map[string]interface{}, error) {
	queryTableSizeSql, queryTableSizeSqlList, err := pumpkinBuildRemoteTableSizeSQL(datasourceType)
	if err != nil {
		return nil, err
	}

	var tableSizeList []map[string]interface{}
	if len(queryTableSizeSqlList) > 0 {
		var lastErr error
		for _, oneSql := range queryTableSizeSqlList {
			tableSizeList, err = database.QueryRemote(db, oneSql)
			if err == nil {
				lastErr = nil
				break
			}
			lastErr = err
		}
		if lastErr != nil {
			return nil, fmt.Errorf("查询表容量数据失败: %v", lastErr)
		}
	} else {
		tableSizeList, err = database.QueryRemote(db, queryTableSizeSql)
		if err != nil {
			return nil, fmt.Errorf("查询表容量数据失败: %v", err)
		}
	}
	return normalizePumpkinRowKeysToLower(tableSizeList), nil
}

// queryPumpkinRemoteTableRowsFromConn 在已有远程连接上查询表级容量（供生命周期任务与行时间 MIN/MAX 共用同一连接）。
func queryPumpkinRemoteTableRowsFromConn(db *sql.DB, datasourceType string) ([]map[string]interface{}, error) {
	return execPumpkinRemoteTableSizeQuery(db, datasourceType)
}

// queryPumpkinRemoteTableRows 在各数据源执行表级容量与引擎时间列查询（gather_pumpkin 与 gather_table_lifecycle 共用）。
func queryPumpkinRemoteTableRows(datasourceType, host, port, user, origPass, dbid string) ([]map[string]interface{}, error) {
	dbCon := getPumpkinDbCon(datasourceType, host, port, user, origPass, dbid)
	if dbCon == nil {
		return nil, fmt.Errorf("无法连接到数据库 %s:%s", host, port)
	}
	defer dbCon.Close()
	return execPumpkinRemoteTableSizeQuery(dbCon, datasourceType)
}
