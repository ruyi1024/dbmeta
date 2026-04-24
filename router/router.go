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

package router

import (
	"dbmeta-core/src/controller/ai"
	"dbmeta-core/src/controller/config"
	"dbmeta-core/src/controller/dashboard"
	"dbmeta-core/src/controller/data"
	"dbmeta-core/src/controller/dataquality"
	"dbmeta-core/src/controller/datasource"
	"dbmeta-core/src/controller/edition"
	"dbmeta-core/src/controller/favorite"
	"dbmeta-core/src/controller/grading"
	"dbmeta-core/src/controller/meta"
	"dbmeta-core/src/controller/pumpkin"
	"dbmeta-core/src/controller/query"
	"dbmeta-core/src/controller/task"
	"dbmeta-core/src/controller/users"
	"dbmeta-core/src/middleware"
	"dbmeta-core/src/module"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	r := gin.New()
	// session
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("dbmeta-session", store))
	r.Use(middleware.Auth())

	v1 := r.Group("api/v1")
	{
		v1.GET("/currentUser", users.CurrentUser)
		v1.GET("/edition", edition.GetEdition)
		v1.POST("/login/account", users.Login)
		v1.GET("/login/outLogin", users.Logout)
		v1.GET("/users/manager/lists", users.GetUsers)
		v1.POST("/users/manager/lists", users.PostUser)
		v1.PUT("/users/manager/lists", users.PutUser)
		v1.DELETE("/users/manager/lists", users.DeleteUser)

		v1.GET("/datasource/list", datasource.List)
		v1.POST("/datasource/list", datasource.List)
		v1.PUT("/datasource/list", datasource.List)
		v1.DELETE("/datasource/list", datasource.List)
		v1.POST("/datasource/check", datasource.Check)

		v1.GET("/datasource_type/list", datasource.TypeList)
		v1.POST("/datasource_type/list", datasource.TypeList)
		v1.PUT("/datasource_type/list", datasource.TypeList)
		v1.DELETE("/datasource_type/list", datasource.TypeList)

		v1.GET("/datasource_idc/list", datasource.IdcList)
		v1.POST("/datasource_idc/list", datasource.IdcList)
		v1.PUT("/datasource_idc/list", datasource.IdcList)
		v1.DELETE("/datasource_idc/list", datasource.IdcList)

		v1.GET("/datasource_env/list", datasource.EnvList)
		v1.POST("/datasource_env/list", datasource.EnvList)
		v1.PUT("/datasource_env/list", datasource.EnvList)
		v1.DELETE("/datasource_env/list", datasource.EnvList)
		v1.GET("/setting/notice", config.NoticeGet)
		v1.PUT("/setting/notice/mail", config.NoticeMailPut)
		v1.PUT("/setting/notice/aliyun", config.NoticeAliyunPut)
		v1.PUT("/setting/notice/wechat", config.NoticeWechatPut)

		v1.GET("/task/option", task.OptionList)
		v1.POST("/task/option", task.OptionList)
		v1.PUT("/task/option", task.OptionList)
		v1.DELETE("/task/option", task.OptionList)
		v1.GET("/task/log", task.TaskLogList)
		v1.GET("/task/log/stats", task.TaskLogStats)
		v1.GET("/task/today/stats", task.TaskTodayStats)

		// 数据洞察（大模型分析任务）API 由企业版 insight 模块注册，见 dbmeta-enterprise/insight。

		// 数据告警相关接口
		v1.GET("/data/alarm/list", data.DataAlarmList)
		v1.POST("/data/alarm/create", data.DataAlarmList)
		v1.PUT("/data/alarm/update", data.DataAlarmList)
		v1.DELETE("/data/alarm/delete/:id", data.DataAlarmList)
		v1.PUT("/data/alarm/toggle-status", data.ToggleDataAlarmStatus)
		v1.POST("/data/alarm/execute", data.ExecuteDataAlarm)
		v1.GET("/data/alarm/logs", data.DataAlarmLogs)
		v1.GET("/data/alarm/report/:id", data.GetDataAlarmReport)
		v1.GET("/data/alarm/detail/:id", data.GetDataAlarmDetail)
		v1.POST("/data/alarm/test-sql", data.TestSqlQuery)
		v1.GET("/data/alarm/datasource-type", data.GetDatasourceTypeList)
		v1.GET("/data/alarm/datasource", data.GetDatasourceList)
		v1.GET("/data/alarm/database", data.GetDatabaseList)

		// 数据安全相关 API（privilege / sensitive / safe）由企业版 security 插件注册，见 dbmeta-enterprise/security。

		v1.GET("/query/datasource_type", query.DataSourceTypeList)
		v1.GET("/query/datasource", query.DataSourceList)
		v1.GET("/query/database", query.DatabaseList)
		v1.GET("/query/table", query.TableList)
		v1.POST("/query/doQuery", query.DoQuery)
		// POST /query/writeLog 由企业版 audit 插件注册（导出等场景写审计）。

		v1.GET("/favorite/list", favorite.List)
		v1.POST("/favorite/list", favorite.List)
		v1.PUT("/favorite/list", favorite.List)
		v1.DELETE("/favorite/list", favorite.List)

		v1.GET("/meta/instance/list", meta.InstanceList)
		v1.GET("/meta/database/list", meta.DatabaseList)
		v1.PUT("/meta/database/list", meta.DatabaseList)
		v1.GET("/meta/business-info/list", meta.BusinessInfo)
		v1.POST("/meta/business-info/list", meta.BusinessInfo)
		v1.PUT("/meta/business-info/list", meta.BusinessInfo)
		v1.DELETE("/meta/business-info/:id", meta.BusinessInfoDelete)
		v1.GET("/meta/database-business/list", meta.DatabaseBusiness)
		v1.POST("/meta/database-business/list", meta.DatabaseBusiness)
		v1.PUT("/meta/database-business/list", meta.DatabaseBusiness)
		v1.POST("/meta/database-business/batch-sync", meta.DatabaseBusinessBatchSync)
		v1.DELETE("/meta/database-business/:id", meta.DatabaseBusinessDelete)
		v1.GET("/meta/table/list", meta.TableList)
		v1.PUT("/meta/table/batch-update-ai-fixed", meta.BatchUpdateAiFixed)
		v1.PUT("/meta/table/update-ai-comment", meta.UpdateTableAiComment)
		v1.GET("/meta/column/list", meta.ColumnList)
		v1.PUT("/meta/column/batch-update-ai-fixed", meta.ColumnBatchUpdateAiFixed)
		v1.PUT("/meta/column/update-ai-comment", meta.UpdateColumnAiComment)
		v1.GET("/meta/dashboard/info", meta.DashboardInfo)
		v1.GET("/meta/quality/info", meta.QualityInfo)

		v1.GET("/meta/env/list", meta.MetaEnvList)
		v1.POST("/meta/env/list", meta.MetaEnvList)
		v1.PUT("/meta/env/list", meta.MetaEnvList)
		v1.DELETE("/meta/env/list", meta.MetaEnvList)

		v1.GET("/task/list", task.TaskList)
		v1.POST("/task/option/execute", task.ExecuteTask)

		// 数据查询审计 /audit/query_log 由企业版 audit 插件注册。

		// 数据质量相关接口
		v1.GET("/dataquality/dashboard/info", dataquality.DashboardInfo)
		v1.GET("/dataquality/issues", dataquality.GetIssues)
		v1.PUT("/dataquality/issues/status", dataquality.UpdateIssueStatus)
		v1.GET("/dataquality/rules", dataquality.GetRules)
		v1.POST("/dataquality/rules", dataquality.CreateRule)
		v1.PUT("/dataquality/rules", dataquality.UpdateRule)
		v1.DELETE("/dataquality/rules/:id", dataquality.DeleteRule)
		v1.GET("/dataquality/tasks", dataquality.GetTasks)
		v1.POST("/dataquality/tasks", dataquality.CreateTask)
		v1.PUT("/dataquality/tasks/status", dataquality.UpdateTaskStatus)
		v1.POST("/dataquality/tasks/execute", dataquality.ExecuteTask)
		v1.DELETE("/dataquality/tasks/:id", dataquality.DeleteTask)

		// 数据分级
		v1.GET("/grading/grades", grading.ListGrades)
		v1.PUT("/grading/grades", grading.UpdateGrade)
		v1.GET("/grading/assets", grading.ListAssets)
		v1.POST("/grading/assets", grading.CreateAsset)
		v1.PUT("/grading/assets", grading.UpdateAsset)
		v1.DELETE("/grading/assets/:id", grading.DeleteAsset)
		v1.GET("/grading/logs", grading.ListLogs)

		v1.GET("/dashbaord/websocket", dashboard.EventWS)
		v1.GET("/dashbaord/info", dashboard.MetaInfo)

		// 数据容量相关接口
		v1.GET("/pumpkin/capacity/stats", pumpkin.GetCapacityStats)
		v1.GET("/pumpkin/capacity/database/type-distribution", pumpkin.GetDatabaseTypeDistribution)
		v1.GET("/pumpkin/capacity/database/top10/chart", pumpkin.GetDatabaseCapacityTop10Chart)
		v1.GET("/pumpkin/capacity/database/top10", pumpkin.GetDatabaseCapacityTop10)
		v1.GET("/pumpkin/capacity/table/top10", pumpkin.GetTableCapacityTop10)
		v1.GET("/pumpkin/capacity/table/growth", pumpkin.GetTableCapacityGrowth)
		v1.GET("/pumpkin/capacity/table/fragmentation/top10", pumpkin.GetTableFragmentationTop10)
		v1.GET("/pumpkin/capacity/table/rows/top10", pumpkin.GetTableRowsTop10)

		// AI相关接口
		v1.POST("/ai/chat", ai.Chat)
		v1.POST("/ai/dify", ai.DifyChat)
		v1.GET("/ai/agents", ai.GetAgents)
		v1.POST("/ai/feedback", ai.SubmitFeedback)
		v1.GET("/ai/feedback/stats", ai.GetFeedbackStats)
		v1.POST("/database/analysis", ai.DatabaseAnalysis)

		// AI Chat查询相关接口
		v1.POST("/ai/chat/query", ai.ChatQuery)
		v1.GET("/ai/chat/sessions", ai.GetSessions)
		v1.POST("/ai/chat/sessions", ai.CreateSession)
		v1.POST("/ai/dbquery", ai.DbQuery)
		v1.DELETE("/ai/chat/sessions/:sessionId", ai.DeleteSession)
		v1.GET("/ai/chat/sessions/:sessionId/messages", ai.GetSessionMessages)
		v1.PUT("/ai/chat/sessions/:sessionId/title", ai.UpdateSessionTitle)
		v1.GET("/ai/chat/rules", ai.GetRules)
		v1.GET("/ai/chat/rules/recommended", ai.GetRecommendedRules)
		v1.POST("/ai/chat/rules", ai.CreateRule)
		v1.PUT("/ai/chat/rules/:id", ai.UpdateRule)
		v1.DELETE("/ai/chat/rules/:id", ai.DeleteRule)

		// 默认模型（不可使用 /ai/models/... 前缀，否则与下方 /ai/models/:id 在 Gin 路由树中冲突）
		v1.GET("/ai/model-defaults", ai.GetAIModelDefaults)
		v1.PUT("/ai/model-defaults", ai.UpdateAIModelDefaults)

		// AI模型管理相关接口
		modelsGroup := v1.Group("/ai/models")
		{
			modelsGroup.GET("", ai.GetModels)
			modelsGroup.GET("enabled", ai.GetEnabledModels)
			modelsGroup.POST("", ai.CreateModel)
			modelsGroup.PUT(":id", ai.UpdateModel)
			modelsGroup.DELETE(":id", ai.DeleteModel)
			modelsGroup.PUT(":id/toggle", ai.ToggleModel)
			modelsGroup.POST(":id/test", ai.TestModel)
		}
		// 测试配置接口使用不同的路径前缀，避免与 :id 路由冲突
		v1.POST("/ai/model/test-config", ai.TestModelConfig)

		// 外部模块路由扩展入口（如企业版插件）。
		module.RegisterRoutes(v1)
	}

	return r
}
