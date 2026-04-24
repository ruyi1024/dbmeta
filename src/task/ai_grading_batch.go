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
	"github.com/ruyi1024/dbmeta/src/service"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	taskKeyAiGradingBatch       = "ai_grading_batch"
	lowConfidenceThreshold int8 = 60
	maxTablesPerRun             = 30
	maxColumnsPerRun            = 40
	aiAssignSource              = "ai"
	operatorSystem              = "ai_grading_batch"
)

type aiGradeJSON struct {
	GradeCode  string `json:"grade_code"`
	Confidence int    `json:"confidence"`
	Reason     string `json:"reason"`
}

// gradingLLMInvoker 优先使用「数据分级」场景默认模型（OpenAI 兼容）；否则回退 Dify common_chat_agent。
type gradingLLMInvoker struct {
	client  service.AIClient
	apiURL  string
	apiKey  string
	timeout time.Duration
}

func newGradingLLMInvoker(logger *zap.Logger) (*gradingLLMInvoker, error) {
	def, err := service.GetDefaultAIModelByScenario(model.AIModelScenarioGrading)
	if err != nil {
		logger.Warn("读取分级默认模型失败，将尝试 Dify", zap.Error(err))
		def = nil
	}
	if def != nil && def.Enabled != 1 {
		logger.Warn("分级默认模型未启用，将尝试 Dify", zap.Int("model_id", def.Id))
		def = nil
	}
	if def != nil {
		cli, cerr := service.NewAIClient(def)
		if cerr != nil {
			logger.Warn("创建分级默认模型客户端失败，将尝试 Dify", zap.String("model", def.Name), zap.Error(cerr))
		} else {
			logger.Info("数据分级批处理使用配置的默认模型",
				zap.String("provider", def.Provider), zap.String("model", def.ModelName))
			return &gradingLLMInvoker{client: cli}, nil
		}
	}
	apiURL, apiKey, timeout, derr := getDifyConfigForTableComment()
	if derr != nil {
		return nil, fmt.Errorf("未配置可用的分级默认模型且 Dify 不可用: %w", derr)
	}
	logger.Info("数据分级批处理使用 Dify（common_chat_agent）")
	return &gradingLLMInvoker{apiURL: apiURL, apiKey: apiKey, timeout: timeout}, nil
}

// newTableColumnCommentLLMInvoker 表/字段 AI 备注任务：优先「表字段备注生成」默认模型，否则回退 Dify
func newTableColumnCommentLLMInvoker(logger *zap.Logger) (*gradingLLMInvoker, error) {
	def, err := service.GetDefaultAIModelByScenario(model.AIModelScenarioTableColumnComment)
	if err != nil {
		logger.Warn("读取表字段备注默认模型失败，将尝试 Dify", zap.Error(err))
		def = nil
	}
	if def != nil && def.Enabled != 1 {
		logger.Warn("表字段备注默认模型未启用，将尝试 Dify", zap.Int("model_id", def.Id))
		def = nil
	}
	if def != nil {
		cli, cerr := service.NewAIClient(def)
		if cerr != nil {
			logger.Warn("创建表字段备注默认模型客户端失败，将尝试 Dify", zap.String("model", def.Name), zap.Error(cerr))
		} else {
			logger.Info("表/字段备注生成使用配置的默认模型",
				zap.String("provider", def.Provider), zap.String("model", def.ModelName))
			return &gradingLLMInvoker{client: cli}, nil
		}
	}
	apiURL, apiKey, timeout, derr := getDifyConfigForTableComment()
	if derr != nil {
		return nil, fmt.Errorf("未配置可用的表字段备注默认模型且 Dify 不可用: %w", derr)
	}
	logger.Info("表/字段备注生成使用 Dify（common_chat_agent）")
	return &gradingLLMInvoker{apiURL: apiURL, apiKey: apiKey, timeout: timeout}, nil
}

// newTableColumnAccuracyLLMInvoker 表字段注释准确度评估：优先「表字段准确度评估」默认模型，否则回退 Dify
func newTableColumnAccuracyLLMInvoker(logger *zap.Logger) (*gradingLLMInvoker, error) {
	def, err := service.GetDefaultAIModelByScenario(model.AIModelScenarioTableColumnAccuracy)
	if err != nil {
		logger.Warn("读取表字段准确度评估默认模型失败，将尝试 Dify", zap.Error(err))
		def = nil
	}
	if def != nil && def.Enabled != 1 {
		logger.Warn("表字段准确度评估默认模型未启用，将尝试 Dify", zap.Int("model_id", def.Id))
		def = nil
	}
	if def != nil {
		cli, cerr := service.NewAIClient(def)
		if cerr != nil {
			logger.Warn("创建表字段准确度评估模型客户端失败，将尝试 Dify", zap.String("model", def.Name), zap.Error(cerr))
		} else {
			logger.Info("表字段准确度评估使用配置的默认模型",
				zap.String("provider", def.Provider), zap.String("model", def.ModelName))
			return &gradingLLMInvoker{client: cli}, nil
		}
	}
	apiURL, apiKey, timeout, derr := getDifyConfigForTableComment()
	if derr != nil {
		return nil, fmt.Errorf("未配置可用的表字段准确度评估默认模型且 Dify 不可用: %w", derr)
	}
	logger.Info("表字段准确度评估使用 Dify（common_chat_agent）")
	return &gradingLLMInvoker{apiURL: apiURL, apiKey: apiKey, timeout: timeout}, nil
}

func (g *gradingLLMInvoker) complete(prompt string) (string, error) {
	if g.client != nil {
		resp, err := g.client.Chat([]service.Message{{Role: "user", Content: prompt}}, nil)
		if err != nil {
			return "", err
		}
		if resp == nil || strings.TrimSpace(resp.Content) == "" {
			return "", fmt.Errorf("模型返回空内容")
		}
		return resp.Content, nil
	}
	return callDifyAPIForTableComment(prompt, g.apiURL, g.apiKey, g.timeout)
}

func init() {
	go aiGradingBatchCrontabTask()
}

func aiGradingBatchCrontabTask() {
	time.Sleep(time.Second * time.Duration(35))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", taskKeyAiGradingBatch).Take(&record)
	if record.Crontab == "" {
		record.Crontab = "*/30 * * * *"
	}
	c := cron.New()
	_, _ = c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", taskKeyAiGradingBatch).Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key=?", taskKeyAiGradingBatch).Updates(map[string]interface{}{
				"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999"),
			})
			doAiGradingBatchTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key=?", taskKeyAiGradingBatch).Updates(map[string]interface{}{
				"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999"),
			})
		}
	})
	c.Start()
}

func doAiGradingBatchTask() {
	logger := log.Logger
	logger.Info("开始执行 AI 数据分级批处理任务")

	taskLogger := NewTaskLogger(taskKeyAiGradingBatch)
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	llm, err := newGradingLLMInvoker(logger)
	if err != nil {
		msg := fmt.Sprintf("初始化分级推理后端失败: %v", err)
		logger.Error(msg)
		taskLogger.Failed(msg)
		return
	}

	var grades []model.DataSecurityGrade
	database.DB.Where("enable = 1").Order("level_order ASC").Find(&grades)
	codeToID := map[string]int64{}
	for _, g := range grades {
		codeToID[strings.ToUpper(g.GradeCode)] = g.Id
	}
	if len(codeToID) == 0 {
		taskLogger.Failed("未找到启用的分级字典，请先初始化 data_security_grade")
		return
	}

	tableIDs, err := queryTableCandidateIDs(lowConfidenceThreshold, maxTablesPerRun)
	if err != nil {
		taskLogger.Failed(fmt.Sprintf("查询待分级表失败: %v", err))
		return
	}
	colIDs, err := queryColumnCandidateIDs(lowConfidenceThreshold, maxColumnsPerRun)
	if err != nil {
		taskLogger.Failed(fmt.Sprintf("查询待分级列失败: %v", err))
		return
	}

	total := len(tableIDs) + len(colIDs)
	if total == 0 {
		msg := "没有需要 AI 分级的表或列（无分级或 AI 低置信度）"
		logger.Info(msg)
		taskLogger.Success(msg)
		return
	}

	taskLogger.UpdateResult(fmt.Sprintf("待处理: 表 %d、列 %d", len(tableIDs), len(colIDs)))

	ok, fail := 0, 0
	var errs []string

	for _, id := range tableIDs {
		var mt model.MetaTable
		if database.DB.First(&mt, id).Error != nil {
			fail++
			continue
		}
		e := processMetaTableGrading(mt, llm, codeToID)
		if e != nil {
			fail++
			errs = append(errs, fmt.Sprintf("表 %s: %v", mt.TableNameX, e))
			logger.Warn("表分级失败", zap.String("table", mt.TableNameX), zap.Error(e))
		} else {
			ok++
		}
		taskLogger.UpdateResult(fmt.Sprintf("表进度 已处理，成功 %d 失败 %d", ok, fail))
		time.Sleep(2 * time.Second)
	}

	ok2, fail2 := 0, 0
	for _, id := range colIDs {
		var mc model.MetaColumn
		if database.DB.First(&mc, id).Error != nil {
			fail2++
			continue
		}
		e := processMetaColumnGrading(mc, llm, codeToID)
		if e != nil {
			fail2++
			errs = append(errs, fmt.Sprintf("列 %s.%s: %v", mc.TableNameX, mc.ColumnName, e))
			logger.Warn("列分级失败", zap.String("col", mc.ColumnName), zap.Error(e))
		} else {
			ok2++
		}
		taskLogger.UpdateResult(fmt.Sprintf("列进度 成功 %d 失败 %d", ok2, fail2))
		time.Sleep(2 * time.Second)
	}

	summary := fmt.Sprintf("完成：表 成功 %d 失败 %d；列 成功 %d 失败 %d", ok, fail, ok2, fail2)
	if len(errs) > 0 && len(errs) <= 3 {
		summary += "；" + strings.Join(errs, "；")
	} else if len(errs) > 3 {
		summary += fmt.Sprintf("；示例错误: %s 等共 %d 条", errs[0], len(errs))
	}
	if fail+fail2 == 0 {
		taskLogger.Success(summary)
	} else {
		taskLogger.Failed(summary)
	}
	logger.Info(summary)
}

// ExecuteAiGradingBatchTask 手动触发（与定时任务相同逻辑）
func ExecuteAiGradingBatchTask() {
	doAiGradingBatchTask()
}

func queryTableCandidateIDs(threshold int8, limit int) ([]int, error) {
	var rows []struct {
		Id int `gorm:"column:id"`
	}
	sql := `
SELECT mt.id AS id
FROM meta_table mt
INNER JOIN datasource ds ON ds.type = mt.datasource_type AND ds.host = mt.host AND ds.port = mt.port AND ds.enable = 1
LEFT JOIN data_asset_security_grade ag ON ag.datasource_id = ds.id
  AND ag.database_name = mt.database_name
  AND ag.table_name = mt.table_name
  AND ag.column_name = ''
WHERE mt.is_deleted = 0
AND (
  ag.id IS NULL
  OR (ag.assign_source = ? AND ag.confidence IS NOT NULL AND ag.confidence < ?)
)
LIMIT ?
`
	err := database.DB.Raw(sql, aiAssignSource, threshold, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(rows))
	for i, r := range rows {
		ids[i] = r.Id
	}
	return ids, nil
}

func queryColumnCandidateIDs(threshold int8, limit int) ([]int, error) {
	var rows []struct {
		Id int `gorm:"column:id"`
	}
	sql := `
SELECT mc.id AS id
FROM meta_column mc
INNER JOIN datasource ds ON ds.type = mc.datasource_type AND ds.host = mc.host AND ds.port = mc.port AND ds.enable = 1
LEFT JOIN data_asset_security_grade ag ON ag.datasource_id = ds.id
  AND ag.database_name = mc.database_name
  AND ag.table_name = mc.table_name
  AND ag.column_name = mc.column_name
WHERE mc.is_deleted = 0
AND mc.column_name <> ''
AND (
  ag.id IS NULL
  OR (ag.assign_source = ? AND ag.confidence IS NOT NULL AND ag.confidence < ?)
)
LIMIT ?
`
	err := database.DB.Raw(sql, aiAssignSource, threshold, limit).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(rows))
	for i, r := range rows {
		ids[i] = r.Id
	}
	return ids, nil
}

func resolveDatasourceID(mt model.MetaTable) (int, error) {
	var ds model.Datasource
	err := database.DB.Where("type = ? AND host = ? AND port = ? AND enable = 1", mt.DatasourceType, mt.Host, mt.Port).First(&ds).Error
	if err != nil {
		return 0, err
	}
	return ds.Id, nil
}

func resolveDatasourceIDColumn(mc model.MetaColumn) (int, error) {
	var ds model.Datasource
	err := database.DB.Where("type = ? AND host = ? AND port = ? AND enable = 1", mc.DatasourceType, mc.Host, mc.Port).First(&ds).Error
	if err != nil {
		return 0, err
	}
	return ds.Id, nil
}

func processMetaTableGrading(mt model.MetaTable, llm *gradingLLMInvoker, codeToID map[string]int64) error {
	dsID, err := resolveDatasourceID(mt)
	if err != nil {
		return fmt.Errorf("匹配数据源: %w", err)
	}
	prompt := buildTableGradingPrompt(mt.DatabaseName, mt.TableNameX, mt.TableComment, mt.AiComment)
	answer, err := llm.complete(prompt)
	if err != nil {
		return err
	}
	code, conf := parseGradeAnswer(answer)
	gradeID, ok := codeToID[code]
	if !ok {
		gradeID = codeToID["GENERAL"]
	}
	if gradeID == 0 {
		return fmt.Errorf("无效分级代码: %s", code)
	}
	return upsertAssetGrade(dsID, mt.DatabaseName, mt.TableNameX, "", gradeID, conf)
}

func processMetaColumnGrading(mc model.MetaColumn, llm *gradingLLMInvoker, codeToID map[string]int64) error {
	dsID, err := resolveDatasourceIDColumn(mc)
	if err != nil {
		return fmt.Errorf("匹配数据源: %w", err)
	}
	prompt := buildColumnGradingPrompt(mc.DatabaseName, mc.TableNameX, mc.ColumnName, mc.ColumnComment, mc.AiComment, mc.DataType)
	answer, err := llm.complete(prompt)
	if err != nil {
		return err
	}
	code, conf := parseGradeAnswer(answer)
	gradeID, ok := codeToID[code]
	if !ok {
		gradeID = codeToID["GENERAL"]
	}
	if gradeID == 0 {
		return fmt.Errorf("无效分级代码: %s", code)
	}
	return upsertAssetGrade(dsID, mc.DatabaseName, mc.TableNameX, mc.ColumnName, gradeID, conf)
}

func buildTableGradingPrompt(dbName, tableName, comment, aiComment string) string {
	return fmt.Sprintf(`你是数据安全分级助手。请仅根据下列元数据，在「一般数据 GENERAL」「重要数据 IMPORTANT」「核心数据 CORE」中选一项。
输出必须是单行 JSON，不要 Markdown，格式严格为：
{"grade_code":"GENERAL","confidence":75,"reason":"一句话理由"}

规则：涉及国家安全、国民经济命脉、大规模个人信息且影响重大等倾向更高等级；无法判断时倾向 GENERAL；confidence 为 0-100 的整数。

库名: %s
表名: %s
表注释: %s
AI表注释: %s`, dbName, tableName, nullStr(comment), nullStr(aiComment))
}

func buildColumnGradingPrompt(dbName, tableName, colName, colComment, aiColComment, dataType string) string {
	return fmt.Sprintf(`你是数据安全分级助手。请仅根据下列字段元数据，在「一般数据 GENERAL」「重要数据 IMPORTANT」「核心数据 CORE」中选一项。
输出必须是单行 JSON，不要 Markdown，格式严格为：
{"grade_code":"GENERAL","confidence":70,"reason":"一句话理由"}

身份证、手机号、银行卡、密码、生物特征等敏感字段倾向更高等级；无法判断时倾向 GENERAL；confidence 为 0-100 的整数。

库名: %s
表名: %s
列名: %s
类型: %s
列注释: %s
AI列注释: %s`, dbName, tableName, colName, nullStr(dataType), nullStr(colComment), nullStr(aiColComment))
}

func nullStr(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "(无)"
	}
	return s
}

func parseGradeAnswer(answer string) (code string, conf int8) {
	answer = strings.TrimSpace(answer)
	// 尝试提取 JSON 对象
	if i := strings.Index(answer, "{"); i >= 0 {
		if j := strings.LastIndex(answer, "}"); j > i {
			answer = answer[i : j+1]
		}
	}
	var j aiGradeJSON
	if err := json.Unmarshal([]byte(answer), &j); err == nil && j.GradeCode != "" {
		code = normalizeGradeCode(j.GradeCode)
		if code == "" {
			code = "GENERAL"
		}
		c := j.Confidence
		if c < 0 {
			c = 0
		}
		if c > 100 {
			c = 100
		}
		return code, int8(c)
	}
	// 回退：全文匹配
	code = normalizeGradeCode(answer)
	if code != "" {
		return code, 65
	}
	return "GENERAL", 50
}

func normalizeGradeCode(s string) string {
	u := strings.ToUpper(strings.TrimSpace(s))
	if strings.Contains(u, "CORE") {
		return "CORE"
	}
	if strings.Contains(u, "IMPORTANT") {
		return "IMPORTANT"
	}
	if strings.Contains(u, "GENERAL") {
		return "GENERAL"
	}
	re := regexp.MustCompile(`(?i)\b(GENERAL|IMPORTANT|CORE)\b`)
	if m := re.FindStringSubmatch(s); len(m) > 1 {
		return strings.ToUpper(m[1])
	}
	return ""
}

func upsertAssetGrade(datasourceID int, databaseName, tableName, columnName string, gradeID int64, confidence int8) error {
	now := time.Now()
	var row model.DataAssetSecurityGrade
	q := database.DB.Where("datasource_id = ? AND database_name = ? AND table_name = ? AND column_name = ?",
		datasourceID, databaseName, tableName, columnName)
	err := q.First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = model.DataAssetSecurityGrade{
			DatasourceId: datasourceID,
			DatabaseName: databaseName,
			TableNameX:   tableName,
			ColumnName:   columnName,
			GradeId:      gradeID,
			AssignSource: aiAssignSource,
			Confidence:   &confidence,
			Remark:       "",
			CreatedBy:    operatorSystem,
			UpdatedBy:    operatorSystem,
			GmtCreated:   now,
			GmtUpdated:   now,
		}
		if err := database.DB.Create(&row).Error; err != nil {
			return err
		}
		database.DB.Create(&model.DataAssetSecurityGradeLog{
			AssetId:    row.Id,
			GradeIdOld: nil,
			GradeIdNew: gradeID,
			Action:     "create",
			Operator:   operatorSystem,
			GmtCreated: now,
		})
		return nil
	}
	if err != nil {
		return err
	}
	prevGrade := row.GradeId
	row.GradeId = gradeID
	row.AssignSource = aiAssignSource
	row.Confidence = &confidence
	row.UpdatedBy = operatorSystem
	row.GmtUpdated = now
	if err := database.DB.Save(&row).Error; err != nil {
		return err
	}
	if prevGrade != gradeID {
		og := prevGrade
		database.DB.Create(&model.DataAssetSecurityGradeLog{
			AssetId:    row.Id,
			GradeIdOld: &og,
			GradeIdNew: gradeID,
			Action:     "update",
			Operator:   operatorSystem,
			GmtCreated: now,
		})
	}
	return nil
}
