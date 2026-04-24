/*
Copyright 2026 The Dbmeta Team Group, website: https://www.dbmeta.com
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
You may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.gnu.org/licenses/gpl-3.0.html

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package task

import (
	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const taskKeyAiTableCommentAccuracy = "ai_table_comment_accuracy"
const tableCommentAccuracyBatchSize = 20

type tableCommentAccuracyJSON struct {
	Reason string  `json:"reason"`
	Score  float64 `json:"score"`
}

type tableCommentAccuracyBatchJSON struct {
	ID     int     `json:"id"`
	Reason string  `json:"reason"`
	Score  float64 `json:"score"`
}

func init() {
	go aiTableCommentAccuracyCrontabTask()
}

func aiTableCommentAccuracyCrontabTask() {
	time.Sleep(30 * time.Second)
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", taskKeyAiTableCommentAccuracy).Take(&record)
	if strings.TrimSpace(record.Crontab) == "" {
		record.Crontab = "*/30 * * * *"
	}
	c := cron.New()
	_, _ = c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", taskKeyAiTableCommentAccuracy).Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key=?", taskKeyAiTableCommentAccuracy).Updates(map[string]interface{}{
				"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999"),
			})
			doAiTableCommentAccuracyTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key=?", taskKeyAiTableCommentAccuracy).Updates(map[string]interface{}{
				"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999"),
			})
		}
	})
	c.Start()
}

func doAiTableCommentAccuracyTask() {
	logger := log.Logger
	logger.Info("开始执行 AI 表注释准确度评估任务")

	taskLogger := NewTaskLogger(taskKeyAiTableCommentAccuracy)
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

	var tables []model.MetaTable
	if err := database.DB.Where("is_deleted = 0").Find(&tables).Error; err != nil {
		msg := fmt.Sprintf("查询表数据失败: %v", err)
		taskLogger.Failed(msg)
		logger.Error(msg)
		return
	}
	if len(tables) == 0 {
		taskLogger.Success("没有需要评估的数据表")
		return
	}

	successCount, failedCount := 0, 0
	var errorDetails []string

	// 预处理：表名或表备注为空的记录，直接置 0.0，不走大模型
	resetResult := database.DB.Model(&model.MetaTable{}).
		Where("is_deleted = 0 AND (table_name IS NULL OR table_name = '' OR table_comment IS NULL OR table_comment = '')").
		Update("table_comment_accuracy", 0.0)
	if resetResult.Error != nil {
		failedCount++
		errorDetails = append(errorDetails, fmt.Sprintf("预处理空表注释失败: %v", resetResult.Error))
	} else {
		successCount += int(resetResult.RowsAffected)
	}

	// 仅对有表名且有表注释的数据调用大模型评估
	batchCandidates := make([]model.MetaTable, 0, len(tables))
	for _, table := range tables {
		if strings.TrimSpace(table.TableNameX) == "" || strings.TrimSpace(table.TableComment) == "" {
			continue
		}
		batchCandidates = append(batchCandidates, table)
	}

	for start := 0; start < len(batchCandidates); start += tableCommentAccuracyBatchSize {
		end := start + tableCommentAccuracyBatchSize
		if end > len(batchCandidates) {
			end = len(batchCandidates)
		}
		chunk := batchCandidates[start:end]
		results, e := evaluateTableCommentAccuracyBatch(chunk, invoker)
		if e != nil {
			failedCount += len(chunk)
			errorDetails = append(errorDetails, fmt.Sprintf("批次 %d-%d 评估失败: %v", start+1, end, e))
			taskLogger.UpdateResult(fmt.Sprintf("已处理 %d/%d，成功 %d，失败 %d", successCount+failedCount, len(tables), successCount, failedCount))
			continue
		}

		for _, table := range chunk {
			result, ok := results[table.Id]
			if !ok {
				failedCount++
				errorDetails = append(errorDetails, fmt.Sprintf("表 %s 缺少评估结果", table.TableNameX))
				continue
			}
			score := normalizeAccuracyScore(result.Score)
			if err := database.DB.Model(&model.MetaTable{}).Where("id = ?", table.Id).Update("table_comment_accuracy", score).Error; err != nil {
				failedCount++
				errorDetails = append(errorDetails, fmt.Sprintf("表 %s 更新失败: %v", table.TableNameX, err))
			} else {
				successCount++
				logger.Info("表注释准确度评估完成",
					zap.String("table_name", table.TableNameX),
					zap.Float64("score", score),
					zap.String("reason", result.Reason))
			}
		}
		taskLogger.UpdateResult(fmt.Sprintf("已处理 %d/%d，成功 %d，失败 %d", successCount+failedCount, len(tables), successCount, failedCount))
	}

	summary := fmt.Sprintf("任务完成 - 总计: %d, 成功: %d, 失败: %d", len(tables), successCount, failedCount)
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

func evaluateTableCommentAccuracyBatch(tables []model.MetaTable, invoker *gradingLLMInvoker) (map[int]tableCommentAccuracyJSON, error) {
	if len(tables) == 0 {
		return map[int]tableCommentAccuracyJSON{}, nil
	}

	lines := make([]string, 0, len(tables))
	for _, table := range tables {
		lines = append(lines, fmt.Sprintf(`{"id":%d,"table_name":"%s","table_comment":"%s"}`,
			table.Id, escapeJSONString(table.TableNameX), escapeJSONString(table.TableComment)))
	}

	prompt := fmt.Sprintf(`你是数据库元数据质量评估助手。请根据每条记录的 table_name 和 table_comment 语义匹配程度，评估注释准确度。
评分规则：
1) score 取值范围 0-1，保留 1 位小数。
2) 0 表示无注释或与表名明显不一致。
3) 1 表示与表名高度匹配、语义准确。
4) 中间值表示部分匹配，根据匹配程度打分，越匹配越接近 1，越不匹配越接近 0。

只输出 JSON 数组，不要 Markdown，不要额外说明。每条结果必须包含 id、score、reason。格式：
[{"id":1,"score":0.8,"reason":"一句话说明"}]

待评估数据：
[%s]`, strings.Join(lines, ","))

	answer, err := invoker.complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("调用模型失败: %v", err)
	}
	parsed, err := parseTableCommentAccuracyBatch(answer)
	if err != nil {
		return nil, fmt.Errorf("解析模型结果失败: %v，原始响应: %s", err, answer)
	}

	out := make(map[int]tableCommentAccuracyJSON, len(parsed))
	for _, item := range parsed {
		if item.ID <= 0 {
			continue
		}
		out[item.ID] = tableCommentAccuracyJSON{
			Score:  normalizeAccuracyScore(item.Score),
			Reason: strings.TrimSpace(item.Reason),
		}
	}
	return out, nil
}

func parseTableCommentAccuracyBatch(answer string) ([]tableCommentAccuracyBatchJSON, error) {
	s := strings.TrimSpace(answer)
	if i := strings.Index(s, "["); i >= 0 {
		if j := strings.LastIndex(s, "]"); j > i {
			s = s[i : j+1]
		}
	}
	var payload []tableCommentAccuracyBatchJSON
	if err := json.Unmarshal([]byte(s), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func escapeJSONString(s string) string {
	s = strings.TrimSpace(s)
	b, _ := json.Marshal(s)
	return strings.Trim(string(b), `"`)
}

func normalizeAccuracyScore(v float64) float64 {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	return math.Round(v*10) / 10
}

// ExecuteAiTableCommentAccuracyTask 手动触发（与定时任务相同逻辑）
func ExecuteAiTableCommentAccuracyTask() {
	doAiTableCommentAccuracyTask()
}
