/*
Copyright 2014-2022 The Lepus Team Group, website: https://www.lepus.cc
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
*/

package task

import (
	"dbmcloud/log"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const taskKeyAiColumnCommentAccuracy = "ai_column_comment_accuracy"
const columnCommentAccuracyBatchSize = 20

type columnCommentAccuracyJSON struct {
	Reason string  `json:"reason"`
	Score  float64 `json:"score"`
}

type columnCommentAccuracyBatchJSON struct {
	ID     int     `json:"id"`
	Reason string  `json:"reason"`
	Score  float64 `json:"score"`
}

func init() {
	go aiColumnCommentAccuracyCrontabTask()
}

func aiColumnCommentAccuracyCrontabTask() {
	time.Sleep(30 * time.Second)
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", taskKeyAiColumnCommentAccuracy).Take(&record)
	if strings.TrimSpace(record.Crontab) == "" {
		record.Crontab = "*/30 * * * *"
	}
	c := cron.New()
	_, _ = c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", taskKeyAiColumnCommentAccuracy).Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key=?", taskKeyAiColumnCommentAccuracy).Updates(map[string]interface{}{
				"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999"),
			})
			doAiColumnCommentAccuracyTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key=?", taskKeyAiColumnCommentAccuracy).Updates(map[string]interface{}{
				"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999"),
			})
		}
	})
	c.Start()
}

func doAiColumnCommentAccuracyTask() {
	logger := log.Logger
	logger.Info("开始执行 AI 字段注释准确度评估任务")

	taskLogger := NewTaskLogger(taskKeyAiColumnCommentAccuracy)
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	invoker, err := newTableColumnAccuracyLLMInvoker(logger)
	if err != nil {
		msg := fmt.Sprintf("初始化模型调用失败: %v", err)
		taskLogger.Failed(msg)
		logger.Error(msg)
		return
	}

	var columns []model.MetaColumn
	if err := database.DB.Where("is_deleted = 0").Find(&columns).Error; err != nil {
		msg := fmt.Sprintf("查询字段数据失败: %v", err)
		taskLogger.Failed(msg)
		logger.Error(msg)
		return
	}
	if len(columns) == 0 {
		taskLogger.Success("没有需要评估的数据字段")
		return
	}

	successCount, failedCount := 0, 0
	var errorDetails []string

	// 预处理：字段名或字段备注为空的记录，直接置 0.0，不走大模型
	resetResult := database.DB.Model(&model.MetaColumn{}).
		Where("is_deleted = 0 AND (column_name IS NULL OR column_name = '' OR column_comment IS NULL OR column_comment = '')").
		Update("column_comment_accuracy", 0.0)
	if resetResult.Error != nil {
		failedCount++
		errorDetails = append(errorDetails, fmt.Sprintf("预处理空字段备注失败: %v", resetResult.Error))
	} else {
		successCount += int(resetResult.RowsAffected)
	}

	// 仅对有字段名且有字段备注的数据调用大模型评估
	batchCandidates := make([]model.MetaColumn, 0, len(columns))
	for _, col := range columns {
		if strings.TrimSpace(col.ColumnName) == "" || strings.TrimSpace(col.ColumnComment) == "" {
			continue
		}
		batchCandidates = append(batchCandidates, col)
	}

	for start := 0; start < len(batchCandidates); start += columnCommentAccuracyBatchSize {
		end := start + columnCommentAccuracyBatchSize
		if end > len(batchCandidates) {
			end = len(batchCandidates)
		}
		chunk := batchCandidates[start:end]

		results, e := evaluateColumnCommentAccuracyBatch(chunk, invoker)
		if e != nil {
			failedCount += len(chunk)
			errorDetails = append(errorDetails, fmt.Sprintf("批次 %d-%d 评估失败: %v", start+1, end, e))
			taskLogger.UpdateResult(fmt.Sprintf("已处理 %d/%d，成功 %d，失败 %d", successCount+failedCount, len(columns), successCount, failedCount))
			continue
		}

		for _, col := range chunk {
			result, ok := results[col.Id]
			if !ok {
				failedCount++
				errorDetails = append(errorDetails, fmt.Sprintf("字段 %s.%s 缺少评估结果", col.TableNameX, col.ColumnName))
				continue
			}
			score := normalizeColumnAccuracyScore(result.Score)
			if err := database.DB.Model(&model.MetaColumn{}).Where("id = ?", col.Id).Update("column_comment_accuracy", score).Error; err != nil {
				failedCount++
				errorDetails = append(errorDetails, fmt.Sprintf("字段 %s.%s 更新失败: %v", col.TableNameX, col.ColumnName, err))
			} else {
				successCount++
				logger.Info("字段注释准确度评估完成",
					zap.String("table_name", col.TableNameX),
					zap.String("column_name", col.ColumnName),
					zap.Float64("score", score),
					zap.String("reason", result.Reason))
			}
		}
		taskLogger.UpdateResult(fmt.Sprintf("已处理 %d/%d，成功 %d，失败 %d", successCount+failedCount, len(columns), successCount, failedCount))
	}

	summary := fmt.Sprintf("任务完成 - 总计: %d, 成功: %d, 失败: %d", len(columns), successCount, failedCount)
	if len(errorDetails) > 0 {
		summary += fmt.Sprintf("。失败详情: %s", errorDetails[0])
		if len(errorDetails) > 1 {
			summary += fmt.Sprintf(" 等 %d 个错误", len(errorDetails))
		}
		taskLogger.Failed(summary)
	} else {
		taskLogger.Success(summary)
	}
	logger.Info(summary)
}

func evaluateColumnCommentAccuracyBatch(columns []model.MetaColumn, invoker *gradingLLMInvoker) (map[int]columnCommentAccuracyJSON, error) {
	if len(columns) == 0 {
		return map[int]columnCommentAccuracyJSON{}, nil
	}

	lines := make([]string, 0, len(columns))
	for _, col := range columns {
		lines = append(lines, fmt.Sprintf(`{"id":%d,"table_name":"%s","column_name":"%s","data_type":"%s","column_comment":"%s"}`,
			col.Id,
			escapeJSONString(col.TableNameX),
			escapeJSONString(col.ColumnName),
			escapeJSONString(col.DataType),
			escapeJSONString(col.ColumnComment),
		))
	}

	prompt := fmt.Sprintf(`你是数据库元数据质量评估助手。请根据每条记录的 table_name、column_name、data_type 与 column_comment 语义匹配程度，评估字段注释准确度。
评分规则：
1) score 取值范围 0-1，保留 1 位小数。
2) 0 表示无注释或与字段语义明显不一致。
3) 1 表示与字段语义高度匹配、注释准确。
4) 中间值表示部分匹配，根据匹配程度打分，越匹配越接近 1。

只输出 JSON 数组，不要 Markdown，不要额外说明。每条结果必须包含 id、score、reason。格式：
[{"id":1,"score":0.8,"reason":"一句话说明"}]

待评估数据：
[%s]`, strings.Join(lines, ","))

	answer, err := invoker.complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("调用模型失败: %v", err)
	}
	parsed, err := parseColumnCommentAccuracyBatch(answer)
	if err != nil {
		return nil, fmt.Errorf("解析模型结果失败: %v，原始响应: %s", err, answer)
	}

	out := make(map[int]columnCommentAccuracyJSON, len(parsed))
	for _, item := range parsed {
		if item.ID <= 0 {
			continue
		}
		out[item.ID] = columnCommentAccuracyJSON{
			Score:  normalizeColumnAccuracyScore(item.Score),
			Reason: strings.TrimSpace(item.Reason),
		}
	}
	return out, nil
}

func parseColumnCommentAccuracyBatch(answer string) ([]columnCommentAccuracyBatchJSON, error) {
	s := strings.TrimSpace(answer)
	if i := strings.Index(s, "["); i >= 0 {
		if j := strings.LastIndex(s, "]"); j > i {
			s = s[i : j+1]
		}
	}
	var payload []columnCommentAccuracyBatchJSON
	if err := json.Unmarshal([]byte(s), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func normalizeColumnAccuracyScore(v float64) float64 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return math.Round(v*10) / 10
}

// ExecuteAiColumnCommentAccuracyTask 手动触发，与定时任务逻辑一致（计划任务平台「手工运行」）
func ExecuteAiColumnCommentAccuracyTask() {
	doAiColumnCommentAccuracyTask()
}

