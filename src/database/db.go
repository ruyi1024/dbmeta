/*
Copyright 2014-2022 The Lepus Team Group, website: https://www.lepus.cc
Licensed under the GNU General Public License, Version 3.0 (the "GPLv3 License");
You may not use this file except in compliance with the License.
You may obtain a copy of the License at
    https://www.gnu.org/licenses/gpl-3.0.html
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
Special note:
Please do not use this source code for any commercial purpose,
or use it for commercial purposes after secondary development, otherwise you may bear legal risks.
*/

package database

import (
	"database/sql"
	"dbmcloud/log"
	"dbmcloud/setting"
	"dbmcloud/src/aes"
	"dbmcloud/src/model"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	//_ "github.com/ClickHouse/clickhouse-go"
	//_ "github.com/go-sql-driver/mysql"
	//_ "github.com/lib/pq"
	"github.com/go-redis/redis"
)

var DB *gorm.DB
var CK *gorm.DB
var SQL *sql.DB
var RDS *redis.Client

func InitDb() *gorm.DB {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			os.Exit(0)

		}
	}()

	ds := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", setting.Setting.User, setting.Setting.Password, setting.Setting.Host, setting.Setting.Port, setting.Setting.Database)
	log.Info("debug mysql: " + fmt.Sprintf("%s", ds))
	sqlDB, err := sql.Open("mysql", ds)
	if err != nil {
		log.Error("open database error", zap.Error(err))
		panic(fmt.Sprintln("open database error.", zap.Error(err)))
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: log.NewGormLogger(zapcore.InfoLevel, zapcore.InfoLevel, time.Millisecond*200),
	})
	if err != nil {
		log.Error("grom open database error", zap.Error(err))
		panic(fmt.Sprintln("grom open database error.", zap.Error(err)))
	}

	if !db.Migrator().HasTable(&model.Users{}) {
		if err = db.AutoMigrate(&model.Users{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		result := db.Create(&model.Users{Id: 1, Username: "admin", ChineseName: "管理员", Password: "a8a0d32f1abefd3fa996321d5e72c6d6", Admin: true})
		if result.Error != nil {
			panic(result.Error)
		}
	}

	if !db.Migrator().HasTable(&model.Token{}) {
		if err = db.AutoMigrate(&model.Token{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.DatasourceType{}) {
		if err = db.AutoMigrate(&model.DatasourceType{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		db.Create(&model.DatasourceType{Id: 1, Name: "MySQL", Sort: 1, Enable: 1})
		db.Create(&model.DatasourceType{Id: 2, Name: "MariaDB", Sort: 2, Enable: 1})
		db.Create(&model.DatasourceType{Id: 3, Name: "GreatSQL", Sort: 3, Enable: 1})
		db.Create(&model.DatasourceType{Id: 4, Name: "TiDB", Sort: 4, Enable: 1})
		db.Create(&model.DatasourceType{Id: 5, Name: "Doris", Sort: 5, Enable: 1})
		db.Create(&model.DatasourceType{Id: 6, Name: "OceanBase", Sort: 6, Enable: 1})
		db.Create(&model.DatasourceType{Id: 7, Name: "ClickHouse", Sort: 7, Enable: 1})
		db.Create(&model.DatasourceType{Id: 8, Name: "Oracle", Sort: 8, Enable: 1})
		db.Create(&model.DatasourceType{Id: 9, Name: "PostgreSQL", Sort: 9, Enable: 1})
		db.Create(&model.DatasourceType{Id: 10, Name: "SQLServer", Sort: 10, Enable: 1})
		db.Create(&model.DatasourceType{Id: 11, Name: "MongoDB", Sort: 11, Enable: 1})
		db.Create(&model.DatasourceType{Id: 12, Name: "Redis", Sort: 12, Enable: 1})

	}

	if !db.Migrator().HasTable(&model.Idc{}) {
		if err = db.AutoMigrate(&model.Idc{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		db.Create(&model.Idc{Id: 1, IdcKey: "default", IdcName: "默认机房", Description: "默认未分类机房"})
	}

	if !db.Migrator().HasTable(&model.Env{}) {
		if err = db.AutoMigrate(&model.Env{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		db.Create(&model.Env{Id: 1, EnvKey: "dev", EnvName: "开发环境", Description: "业务功能开发环境"})
		db.Create(&model.Env{Id: 2, EnvKey: "test", EnvName: "测试环境", Description: "业务功能测试环境"})
		db.Create(&model.Env{Id: 3, EnvKey: "pre", EnvName: "预发环境", Description: "准生产验证环境"})
		db.Create(&model.Env{Id: 4, EnvKey: "prod", EnvName: "生产环境", Description: "线上业务运行环境"})
	}

	if !db.Migrator().HasTable(&model.Datasource{}) {
		if err = db.AutoMigrate(&model.Datasource{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		aesPassword, _ := aes.AesPassEncode(setting.Setting.Password, setting.Setting.DbPassKey)
		aesRdsPassword, _ := aes.AesPassEncode(setting.Setting.RedisPassword, setting.Setting.DbPassKey)
		db.Create(&model.Datasource{Id: 1, Name: "DBMETA-MySQL", GroupName: "Lepus", Idc: "default", Env: "prod", Type: "MySQL", Host: setting.Setting.Host, Port: setting.Setting.Port, User: setting.Setting.User, Pass: aesPassword, Enable: 1, DbmetaEnable: 1, ExecuteEnable: 1, MonitorEnable: 1, AlarmEnable: 1})
		db.Create(&model.Datasource{Id: 3, Name: "DBMETA-Redis", GroupName: "Lepus", Idc: "default", Env: "prod", Type: "Redis", Host: setting.Setting.RedisHost, Port: setting.Setting.RedisPort, User: "", Pass: aesRdsPassword, Enable: 1, DbmetaEnable: 0, ExecuteEnable: 1, MonitorEnable: 1, AlarmEnable: 1})

	}

	if !db.Migrator().HasTable(&model.MetaDatabase{}) {
		if err = db.AutoMigrate(&model.MetaDatabase{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.MetaTable{}) {
		if err = db.AutoMigrate(&model.MetaTable{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.MetaColumn{}) {
		if err = db.AutoMigrate(&model.MetaColumn{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	// 数据质量相关表 - 直接执行 AutoMigrate（GORM 会自动处理表是否存在）
	if err = db.AutoMigrate(&model.DataQualityAssessment{}); err != nil {
		log.Error("db sync DataQualityAssessment error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataQualityIssue{}); err != nil {
		log.Error("db sync DataQualityIssue error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataQualityAiAnalysis{}); err != nil {
		log.Error("db sync DataQualityAiAnalysis error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataQualityAiInsight{}); err != nil {
		log.Error("db sync DataQualityAiInsight error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataQualityAiRecommendation{}); err != nil {
		log.Error("db sync DataQualityAiRecommendation error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataQualityDistribution{}); err != nil {
		log.Error("db sync DataQualityDistribution error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataQualityRule{}); err != nil {
		log.Error("db sync DataQualityRule error.", zap.Error(err))
	}
	// 只有在规则表为空时才添加默认规则
	var ruleCount int64
	db.Model(&model.DataQualityRule{}).Count(&ruleCount)
	if ruleCount == 0 {
		// 添加默认的数据质量规则
		// 完整性规则
		db.Create(&model.DataQualityRule{
			RuleName:   "字段空值率检查",
			RuleType:   "完整性",
			RuleDesc:   "检查字段的空值率，当空值率超过阈值时触发告警",
			RuleConfig: `{"max_null_rate": 0.2, "check_columns": ["*"]}`,
			Threshold:  20.0,
			Severity:   "high",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "必填字段完整性检查",
			RuleType:   "完整性",
			RuleDesc:   "检查关键必填字段是否存在空值，如主键、外键、业务关键字段等",
			RuleConfig: `{"required_fields": ["id", "user_id", "order_id"], "strict_mode": true}`,
			Threshold:  0.0,
			Severity:   "high",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "表记录完整性检查",
			RuleType:   "完整性",
			RuleDesc:   "检查数据表记录数量是否异常，如记录数突然减少或为0",
			RuleConfig: `{"min_record_count": 1, "check_trend": true}`,
			Threshold:  0.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})

		// 准确性规则
		db.Create(&model.DataQualityRule{
			RuleName:   "邮箱格式验证",
			RuleType:   "准确性",
			RuleDesc:   "验证邮箱字段格式是否符合标准邮箱格式",
			RuleConfig: `{"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", "columns": ["email", "mail"]}`,
			Threshold:  5.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "手机号格式验证",
			RuleType:   "准确性",
			RuleDesc:   "验证手机号字段格式是否符合中国手机号规范（11位数字，1开头）",
			RuleConfig: `{"pattern": "^1[3-9]\\d{9}$", "columns": ["phone", "mobile", "tel"]}`,
			Threshold:  5.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "身份证号格式验证",
			RuleType:   "准确性",
			RuleDesc:   "验证身份证号字段格式是否符合18位身份证号规范",
			RuleConfig: `{"pattern": "^[1-9]\\d{5}(18|19|20)\\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\\d|3[01])\\d{3}[0-9Xx]$", "columns": ["id_card", "idcard", "identity_card"]}`,
			Threshold:  5.0,
			Severity:   "high",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "数值范围检查",
			RuleType:   "准确性",
			RuleDesc:   "检查数值字段是否在合理范围内，如年龄、金额、百分比等",
			RuleConfig: `{"age": {"min": 0, "max": 150}, "amount": {"min": 0, "max": 999999999}, "percentage": {"min": 0, "max": 100}}`,
			Threshold:  5.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "日期有效性检查",
			RuleType:   "准确性",
			RuleDesc:   "检查日期字段是否有效，如出生日期不能晚于当前日期，创建时间不能晚于更新时间等",
			RuleConfig: `{"check_future_date": true, "check_date_range": true, "max_future_days": 0}`,
			Threshold:  5.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})

		// 唯一性规则
		db.Create(&model.DataQualityRule{
			RuleName:   "主键唯一性检查",
			RuleType:   "唯一性",
			RuleDesc:   "检查主键字段是否存在重复值，确保主键唯一性",
			RuleConfig: `{"check_primary_key": true, "allow_null": false}`,
			Threshold:  0.0,
			Severity:   "high",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "业务唯一性检查",
			RuleType:   "唯一性",
			RuleDesc:   "检查业务唯一字段是否存在重复值，如用户ID、订单号、业务编号等",
			RuleConfig: `{"unique_fields": ["user_id", "order_no", "business_code"], "check_combination": false}`,
			Threshold:  0.0,
			Severity:   "high",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "组合唯一性检查",
			RuleType:   "唯一性",
			RuleDesc:   "检查多个字段的组合是否唯一，如用户ID+产品ID的组合唯一性",
			RuleConfig: `{"unique_combinations": [["user_id", "product_id"], ["order_id", "item_id"]]}`,
			Threshold:  0.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})

		// 一致性规则
		db.Create(&model.DataQualityRule{
			RuleName:   "外键一致性检查",
			RuleType:   "一致性",
			RuleDesc:   "检查外键字段的值是否在关联表中存在，确保外键引用完整性",
			RuleConfig: `{"check_foreign_keys": true, "strict_mode": true}`,
			Threshold:  0.0,
			Severity:   "high",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "数据关联一致性检查",
			RuleType:   "一致性",
			RuleDesc:   "检查关联表之间的数据一致性，如订单表和订单明细表的金额汇总是否一致",
			RuleConfig: `{"check_relations": true, "aggregation_checks": [{"parent": "orders", "child": "order_items", "parent_field": "total_amount", "child_field": "amount", "operation": "sum"}]}`,
			Threshold:  1.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "枚举值一致性检查",
			RuleType:   "一致性",
			RuleDesc:   "检查枚举类型字段的值是否在预定义的枚举值范围内",
			RuleConfig: `{"enum_fields": {"status": ["active", "inactive", "pending"], "type": ["type1", "type2", "type3"]}}`,
			Threshold:  5.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})

		// 及时性规则
		db.Create(&model.DataQualityRule{
			RuleName:   "数据更新时效性检查",
			RuleType:   "及时性",
			RuleDesc:   "检查数据表的最后更新时间，如果超过指定时间未更新则告警",
			RuleConfig: `{"max_update_interval_hours": 24, "check_fields": ["updated_at", "gmt_updated", "modify_time"]}`,
			Threshold:  10.0,
			Severity:   "low",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "数据同步延迟检查",
			RuleType:   "及时性",
			RuleDesc:   "检查主从库或数据同步场景下的数据同步延迟情况",
			RuleConfig: `{"max_sync_delay_seconds": 300, "check_replication": true}`,
			Threshold:  5.0,
			Severity:   "medium",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "数据采集时效性检查",
			RuleType:   "及时性",
			RuleDesc:   "检查数据采集任务的执行时效性，确保数据及时采集",
			RuleConfig: `{"max_collect_interval_minutes": 60, "check_collect_tasks": true}`,
			Threshold:  10.0,
			Severity:   "low",
			Enabled:    1,
			CreatedBy:  "system",
		})

		// 其他通用规则
		db.Create(&model.DataQualityRule{
			RuleName:   "字符串长度检查",
			RuleType:   "准确性",
			RuleDesc:   "检查字符串字段长度是否符合预期范围",
			RuleConfig: `{"name": {"min": 1, "max": 100}, "description": {"min": 0, "max": 500}, "code": {"min": 1, "max": 50}}`,
			Threshold:  5.0,
			Severity:   "low",
			Enabled:    1,
			CreatedBy:  "system",
		})
		db.Create(&model.DataQualityRule{
			RuleName:   "数值精度检查",
			RuleType:   "准确性",
			RuleDesc:   "检查数值字段的小数位数是否符合精度要求",
			RuleConfig: `{"price": {"decimal_places": 2}, "rate": {"decimal_places": 4}, "amount": {"decimal_places": 2}}`,
			Threshold:  5.0,
			Severity:   "low",
			Enabled:    1,
			CreatedBy:  "system",
		})
	}
	if err = db.AutoMigrate(&model.DataQualityTask{}); err != nil {
		log.Error("db sync DataQualityTask error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataQualityHistory{}); err != nil {
		log.Error("db sync DataQualityHistory error.", zap.Error(err))
	}

	// 数据分类分级（GB 通用三类：一般 / 重要 / 核心）
	if err = db.AutoMigrate(&model.DataSecurityGrade{}); err != nil {
		log.Error("db sync DataSecurityGrade error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataAssetSecurityGrade{}); err != nil {
		log.Error("db sync DataAssetSecurityGrade error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataAssetSecurityGradeLog{}); err != nil {
		log.Error("db sync DataAssetSecurityGradeLog error.", zap.Error(err))
	}
	var dataGradeCount int64
	db.Model(&model.DataSecurityGrade{}).Count(&dataGradeCount)
	if dataGradeCount == 0 {
		now := time.Now()
		db.Create(&model.DataSecurityGrade{
			GradeCode:   "GENERAL",
			GradeName:   "一般数据",
			LevelOrder:  1,
			Description: "危害程度相对较低，按组织数据安全制度管理",
			StandardRef: "GB 国家标准体系-通用分级(三类)",
			Enable:      1,
			GmtCreated:  now,
			GmtUpdated:  now,
		})
		db.Create(&model.DataSecurityGrade{
			GradeCode:   "IMPORTANT",
			GradeName:   "重要数据",
			LevelOrder:  2,
			Description: "一旦遭到篡改、破坏、泄露或者非法获取、非法利用，可能对经济运行、公共利益、个人权益等造成较大危害",
			StandardRef: "GB 国家标准体系-通用分级(三类)",
			Enable:      1,
			GmtCreated:  now,
			GmtUpdated:  now,
		})
		db.Create(&model.DataSecurityGrade{
			GradeCode:   "CORE",
			GradeName:   "核心数据",
			LevelOrder:  3,
			Description: "危害程度高于重要数据，对国家安全、经济运行、社会秩序、公共利益等影响更为重大，需采取更严格保护措施",
			StandardRef: "GB 国家标准体系-通用分级(三类)",
			Enable:      1,
			GmtCreated:  now,
			GmtUpdated:  now,
		})
	}

	if !db.Migrator().HasTable(&model.TaskOption{}) {
		if err = db.AutoMigrate(&model.TaskOption{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	// 初始化默认任务配置（如果不存在则创建）
	defaultTasks := []model.TaskOption{
		{TaskKey: "recycle_token", TaskName: "回收用户令牌", TaskDescription: "回收用户过期的ToKen", Crontab: "* * * * *"},
		{TaskKey: "revoke_privileage", TaskName: "回收用户权限", TaskDescription: "检查用户查询数据库权限是否过期，并回收权限", Crontab: "1 * * * *"},
		{TaskKey: "check_datasource", TaskName: "监测数据源状态", TaskDescription: "监测数据源连接状态是否正常", Crontab: "@every 30s"},
		{TaskKey: "gather_dbmeta", TaskName: "采集元数据信息", TaskDescription: "采集数据库、数据表、数据列等元数据信息", Crontab: "*/3 * * * *"},
		{TaskKey: "gather_sensitive", TaskName: "敏感数据探测分析", TaskDescription: "分析数据库数据，监测敏感信息", Crontab: "*/5 * * * *"},
		{TaskKey: "ai_general_table_comment", TaskName: "AI生成表注释", TaskDescription: "接入AI大模型，自动为缺失注释的数据表生成AI注释", Crontab: "*/30 * * * *"},
		{TaskKey: "ai_general_column_comment", TaskName: "AI生成字段注释", TaskDescription: "接入AI大模型，自动为缺失注释的数据字段生成AI注释", Crontab: "*/30 * * * *"},
		{TaskKey: "ai_apply_table_comment", TaskName: "AI应用表注释", TaskDescription: "将待应用的AI注释应用到实际数据表", Crontab: "*/30 * * * *"},
		{TaskKey: "ai_apply_column_comment", TaskName: "AI应用字段注释", TaskDescription: "将待应用的AI注释应用到实际数据字段", Crontab: "*/30 * * * *"},
		{TaskKey: "data_quality_ai_analysis", TaskName: "数据质量AI分析", TaskDescription: "对数据质量评估结果进行AI智能分析，生成洞察和优化建议", Crontab: "0 * * * *"},
		{TaskKey: "gather_pumpkin", TaskName: "容量数据采集", TaskDescription: "采集数据库容量数据", Crontab: "0 * * * *"},
		{TaskKey: "gather_pumpkin_growth", TaskName: "容量增长分析", TaskDescription: "分析数据库容量增长情况", Crontab: "*/30 * * * *"},
		{TaskKey: "ai_grading_batch", TaskName: "AI数据分级批处理", TaskDescription: "对无分级或低置信度(仅AI)的表/列调用大模型自动标注安全分级", Crontab: "*/30 * * * *"},
	}

	for _, task := range defaultTasks {
		var existingTask model.TaskOption
		result := db.Where("task_key = ?", task.TaskKey).First(&existingTask)
		if result.Error != nil {
			// 如果不存在，创建
			db.Create(&task)
		}
	}

	if !db.Migrator().HasTable(&model.TaskHeartbeat{}) {
		if err = db.AutoMigrate(&model.TaskHeartbeat{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	// 初始化默认任务心跳记录（如果不存在则创建）
	t, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))
	defaultHeartbeats := []model.TaskHeartbeat{
		{HeartbeatKey: "recycle_token", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "revoke_privileage", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "check_datasource", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "gather_dbmeta", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "gather_sensitive", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "gather_pumpkin", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "ai_general_table_comment", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "ai_general_column_comment", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "ai_apply_table_comment", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "ai_apply_column_comment", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "data_quality_ai_analysis", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "gather_pumpkin_growth", HeartbeatTime: t, HeartbeatEndTime: t},
		{HeartbeatKey: "ai_grading_batch", HeartbeatTime: t, HeartbeatEndTime: t},
	}

	for _, heartbeat := range defaultHeartbeats {
		var existingHeartbeat model.TaskHeartbeat
		result := db.Where("heartbeat_key = ?", heartbeat.HeartbeatKey).First(&existingHeartbeat)
		if result.Error != nil {
			// 如果不存在，创建
			db.Create(&heartbeat)
		}
	}

	if !db.Migrator().HasTable(&model.TaskLog{}) {
		if err = db.AutoMigrate(&model.TaskLog{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.Favorite{}) {
		if err = db.AutoMigrate(&model.Favorite{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.Privilege{}) {
		if err = db.AutoMigrate(&model.Privilege{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.QueryLog{}) {
		if err = db.AutoMigrate(&model.QueryLog{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.SensitiveRule{}) {
		if err = db.AutoMigrate(&model.SensitiveRule{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		db.Create(&model.SensitiveRule{RuleKey: "mobile", RuleName: "手机号码", RuleType: "data", RuleExpress: "^1[356789]\\d{9}$|^\\+861\\d{10}$", Level: 1, Status: 1})
		db.Create(&model.SensitiveRule{RuleKey: "id_number", RuleName: "身份证号", RuleType: "data", RuleExpress: "^([1-9]\\d{5}[12]\\d{3}(0[1-9]|1[012])(0[1-9]|[12][0-9]|3[01])\\d{3}[0-9xX])$", Level: 1, Status: 1})
		db.Create(&model.SensitiveRule{RuleKey: "email", RuleName: "电子邮箱", RuleType: "data", RuleExpress: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+.[a-zA-Z]{2,}$", Level: 1, Status: 1})
		db.Create(&model.SensitiveRule{RuleKey: "bank_card", RuleName: "银行卡号", RuleType: "data", RuleExpress: "^[6]\\d{18}", Level: 1, Status: 1})
		db.Create(&model.SensitiveRule{RuleKey: "car_number", RuleName: "车牌号", RuleType: "data", RuleExpress: "^[\\x{4e00}-\\x{9fa2}][A-Z][0-9A-Z]{5}", Level: 0, Status: 1})
		db.Create(&model.SensitiveRule{RuleKey: "address", RuleName: "住址地址", RuleType: "data", RuleExpress: "[\\x{4e00}-\\x{9fa5}]{2,5}[市][\\x{4e00}-\\x{9fa5}]{2,5}[区](.+)[0-9]{1,4}[号]", Level: 0, Status: 1})
		db.Create(&model.SensitiveRule{RuleKey: "ip", RuleName: "IP地址", RuleType: "data", RuleExpress: "^(\\d{1,3}\\.){3}\\d{1,3}$", Level: 0, Status: 1})
		db.Create(&model.SensitiveRule{RuleKey: "ipport", RuleName: "IP端口服务", RuleType: "data", RuleExpress: "^(\\d{1,3}\\.){3}\\d{1,3}\\:\\d{2,6}$", Level: 0, Status: 1})
		db.Create(&model.SensitiveRule{RuleKey: "realname", RuleName: "姓名", RuleType: "data", RuleExpress: "^[\\x{4e00}-\\x{9fa2}]{2,3}$", Level: 1, Status: -1})
		db.Create(&model.SensitiveRule{RuleKey: "username", RuleName: "用户名", RuleType: "column", RuleExpress: "user|username|user_name", Level: 0, Status: -1})
		db.Create(&model.SensitiveRule{RuleKey: "password", RuleName: "密码", RuleType: "column", RuleExpress: "pass|password|pass_word", Level: 1, Status: -1})
	}

	if !db.Migrator().HasTable(&model.SensitiveMeta{}) {
		if err = db.AutoMigrate(&model.SensitiveMeta{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}
	if !db.Migrator().HasTable(&model.Event{}) {
		if err = db.AutoMigrate(&model.Event{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.EventDescription{}) {
		if err = db.AutoMigrate(&model.EventDescription{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.EventGlobal{}) {
		if err = db.AutoMigrate(&model.EventGlobal{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.AlarmChannel{}) {
		if err = db.AutoMigrate(&model.AlarmChannel{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		db.Create(&model.AlarmChannel{ID: 1, Name: "默认渠道", Description: "默认通知事件发送渠道", Enable: 1})
	}

	if !db.Migrator().HasTable(&model.AlarmLevel{}) {
		if err = db.AutoMigrate(&model.AlarmLevel{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		db.Create(&model.AlarmLevel{ID: 1, LevelName: "停服", Description: "服务不可用", Enable: 1})
		db.Create(&model.AlarmLevel{ID: 2, LevelName: "严重", Description: "紧急的严重问题", Enable: 1})
		db.Create(&model.AlarmLevel{ID: 3, LevelName: "警告", Description: "不紧急的重要信息", Enable: 1})
		db.Create(&model.AlarmLevel{ID: 4, LevelName: "提醒", Description: "不紧急不严重需要关注的信息", Enable: 1})

	}

	if !db.Migrator().HasTable(&model.AlarmRule{}) {
		if err = db.AutoMigrate(&model.AlarmRule{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
		db.Create(&model.AlarmRule{Title: "MySQL数据源监测失败", EventType: "MySQL", EventKey: "datasourceCheck", AlarmRule: "!=", AlarmValue: "1", LevelId: 1, Enable: 1})
	}

	if !db.Migrator().HasTable(&model.AlarmEvent{}) {
		if err = db.AutoMigrate(&model.AlarmEvent{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.AlarmSendLog{}) {
		if err = db.AutoMigrate(&model.AlarmSendLog{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.AlarmSuggest{}) {
		if err = db.AutoMigrate(&model.AlarmSuggest{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	if !db.Migrator().HasTable(&model.AlarmTrack{}) {
		if err = db.AutoMigrate(&model.AlarmTrack{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	// Chat sessions table
	if !db.Migrator().HasTable(&model.ChatSession{}) {
		if err = db.AutoMigrate(&model.ChatSession{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	// Chat messages table
	if !db.Migrator().HasTable(&model.ChatMessage{}) {
		if err = db.AutoMigrate(&model.ChatMessage{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	// Semantic SQL rules table
	// 即使表已存在，也执行AutoMigrate以添加新字段
	if err = db.AutoMigrate(&model.SemanticSqlRule{}); err != nil {
		log.Error("db sync error.", zap.Error(err))
	}

	// Analysis Task table
	// 即使表已存在，也执行AutoMigrate以添加新字段
	if err = db.AutoMigrate(&model.AnalysisTask{}); err != nil {
		log.Error("db sync error.", zap.Error(err))
	}

	// Analysis Task Log table
	if !db.Migrator().HasTable(&model.AnalysisTaskLog{}) {
		if err = db.AutoMigrate(&model.AnalysisTaskLog{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}

	// AI models table
	if !db.Migrator().HasTable(&model.AIModel{}) {
		if err = db.AutoMigrate(&model.AIModel{}); err != nil {
			log.Error("db sync error.", zap.Error(err))
		}
	}
	if err = db.AutoMigrate(&model.AIModelDefault{}); err != nil {
		log.Error("db sync AIModelDefault error.", zap.Error(err))
	}

	// Pumpkin growth tables
	if err = db.AutoMigrate(&model.PumpkinTableGrowth{}); err != nil {
		log.Error("db sync PumpkinTableGrowth error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.PumpkinDatabaseGrowth{}); err != nil {
		log.Error("db sync PumpkinDatabaseGrowth error.", zap.Error(err))
	}

	// Data Alarm tables
	// 即使表已存在，也执行AutoMigrate以添加新字段
	if err = db.AutoMigrate(&model.DataAlarm{}); err != nil {
		log.Error("db sync DataAlarm error.", zap.Error(err))
	}
	if err = db.AutoMigrate(&model.DataAlarmLog{}); err != nil {
		log.Error("db sync DataAlarmLog error.", zap.Error(err))
	}

	return db
}

func InitConnect() *sql.DB {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			os.Exit(0)

		}
	}()
	ds := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", setting.Setting.User, setting.Setting.Password, setting.Setting.Host, setting.Setting.Port, setting.Setting.Database)
	db, err := sql.Open("mysql", ds)
	if err != nil {
		log.Error(fmt.Sprintln("Init mysql connect err,", err))
		panic(fmt.Sprintln("Init mysql connect err,", err))
	}
	if err := db.Ping(); err != nil {
		log.Error(fmt.Sprintln("Init mysql ping err,", err))
		panic(fmt.Sprintln("Init mysql ping err,", err))
	}

	return db
}

func QueryAll(sql string) ([]map[string]interface{}, error) {
	rows, err := SQL.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var list []map[string]interface{}
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			continue
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]
			b, ok := v.([]byte)
			if ok {
				entry[col] = string(b)
			} else {
				entry[col] = v
			}
		}
		list = append(list, entry)
	}
	return list, nil
}

// 使用option结构体创建可选参数
type Option struct {
	f func(*options)
}

type options struct {
	driver   string
	host     string
	port     string
	username string
	password string
	database string
	sid      string
	timeout  int
}

func WithDriver(driver string) Option {
	return Option{func(op *options) {
		op.driver = driver
	}}
}
func WithHost(host string) Option {
	return Option{func(op *options) {
		op.host = host
	}}
}
func WithPort(port string) Option {
	return Option{func(op *options) {
		op.port = port
	}}
}
func WithUsername(username string) Option {
	return Option{func(op *options) {
		op.username = username
	}}
}
func WithPassword(password string) Option {
	return Option{func(op *options) {
		op.password = password
	}}
}
func WithDatabase(database string) Option {
	return Option{func(op *options) {
		op.database = database
	}}
}
func WithSid(sid string) Option {
	return Option{func(op *options) {
		op.sid = sid
	}}
}
func WithTimeout(timeout int) Option {
	return Option{func(op *options) {
		op.timeout = timeout
	}}
}

// 使用结构体动态传入参数
func Connect(ops ...Option) (*sql.DB, error) {
	//set option
	opt := &options{}
	for _, do := range ops {
		do.f(opt)
	}
	//不同数据库构造不同的url
	var url string
	if opt.driver == "mysql" {
		url = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=5s&readTimeout=10s", opt.username, opt.password, opt.host, opt.port, opt.database)
	}
	if opt.driver == "postgres" {
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", opt.host, opt.port, opt.username, opt.password, opt.database)
	}
	if opt.driver == "clickhouse" {
		url = fmt.Sprintf("tcp://%s:%s/%s?username=%s&password=%s&read_timeout=30s", opt.host, opt.port, opt.database, opt.username, opt.password)
	}
	if opt.driver == "oracle" {
		url = fmt.Sprintf(`user="%s" password="%s" connectString="%s:%s/%s"`, opt.username, opt.password, opt.host, opt.port, opt.sid)
	}
	if opt.driver == "mssql" {
		url = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=disable;connection timeout=6;", opt.host, opt.username, opt.password, opt.port, opt.database)
	}
	//连接数据库
	db, err := sql.Open(opt.driver, url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func Execute(db *sql.DB, sql string) (rowsAffected int64, err error) {
	res, err := db.Exec(sql)
	if err != nil {
		return 0, err
	}
	rowsAffected, _ = res.RowsAffected()
	return rowsAffected, nil
}

func QueryRemote(db *sql.DB, sql string) ([]map[string]interface{}, error) {
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var list []map[string]interface{}
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			continue
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]
			b, ok := v.([]byte)
			if ok {
				entry[col] = string(b)
				//entry[col] = b
			} else {
				entry[col] = v
			}
		}
		list = append(list, entry)
	}
	return list, nil
}

/*
QueryRemoteNew方法会返回columns，columns顺序是稳定的
*/
func QueryRemoteNew(db *sql.DB, sql string) ([]string, []map[string]interface{}, error) {
	rows, err := db.Query(sql)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var list []map[string]interface{}
	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			continue
		}

		entry := make(map[string]interface{})
		for i, col := range columns {
			fmt.Println(col)
			v := values[i]
			b, ok := v.([]byte)
			if ok {
				entry[col] = string(b)
				//entry[col] = b
			} else {
				entry[col] = v
			}
		}
		list = append(list, entry)
	}
	return columns, list, nil
}

func InitRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", setting.Setting.RedisHost, setting.Setting.RedisPort),
		Password:     setting.Setting.RedisPassword, // no password set
		DB:           0,                             // use default DB
		PoolSize:     128,
		ReadTimeout:  time.Millisecond * time.Duration(2000),
		WriteTimeout: time.Millisecond * time.Duration(2000),
		IdleTimeout:  time.Second * time.Duration(86400),
		MaxRetries:   3,
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Error("open redis client error", zap.Error(err))
		panic(fmt.Sprintln("open redis client error.", zap.Error(err)))
	}
	return redisClient
}

func InitRedisCluster(host, port, password string) (*redis.ClusterClient, error) {
	redisClusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{fmt.Sprintf("%s:%s", host, port)},
		Password:     password, // no password set
		PoolSize:     1000,
		ReadTimeout:  time.Millisecond * time.Duration(200),
		WriteTimeout: time.Millisecond * time.Duration(200),
		IdleTimeout:  time.Second * time.Duration(600),
	})
	_, err := redisClusterClient.Ping().Result()
	if err != nil {
		return nil, err
	}
	return redisClusterClient, nil
}
