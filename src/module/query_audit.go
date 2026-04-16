package module

// QueryAuditWriter 将一次 SQL 查询行为写入 query_log；仅由企业版 audit 插件注册。
type QueryAuditWriter func(
	username, datasourceType, datasource, queryType, sqlType, databaseName, status string,
	times int64,
	content, doResult string,
)

// WriteQueryLog 企业版在 init 中赋值；未注册时开源版不写审计表。
var WriteQueryLog QueryAuditWriter
