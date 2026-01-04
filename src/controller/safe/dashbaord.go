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

package safe

import (
	"dbmcloud/src/database"
	"dbmcloud/src/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func DashboardInfo(c *gin.Context) {

	todayQueryCount, _ := database.QueryAll("select count(*) as count from query_log where gmt_created>current_date() limit 1;")
	totalQueryCount, _ := database.QueryAll("select count(*) as count from query_log limit 1;")
	totalInterceptCount, _ := database.QueryAll("select count(*) as count from query_log where status='intercept' limit 1;")

	sensitiveDatabaseCount, _ := database.QueryAll("select count(distinct database_name) as count from sensitive_meta limit 1;")
	sensitiveTableCount, _ := database.QueryAll("select count(distinct database_name,table_name) as count from sensitive_meta limit 1;")
	sensitiveColumnCount, _ := database.QueryAll("select count(distinct database_name,table_name,column_name) as count from sensitive_meta limit 1;")
	sensitiveQueryCount, _ := database.QueryAll("select count(*) as count from query_log where `database` in (select database_name from sensitive_meta) limit 1;")

	queryStatusPieData, _ := database.QueryAll("select case status when 'failed' then '执行失败' when 'intercept' then '风险拦截' when 'success' then '执行成功' else '其他' end as type,count(*) value from query_log group by status order by count(*) desc limit 10")
	queryStatusPieDataList := make([]map[string]interface{}, 0)
	for _, item := range queryStatusPieData {
		pieData := make(map[string]interface{})
		pieData["type"] = item["type"].(string)
		pieData["value"] = utils.StrToInt(item["value"].(string))
		queryStatusPieDataList = append(queryStatusPieDataList, pieData)
	}

	queryTypePieData, _ := database.QueryAll("select case query_type when 'copyData' then '复制数据内容' when 'doExplain' then '查看执行计划' when 'execute' then '执行SQL查询命令' when 'exportExcel' then '导出Excel文件' when 'showColumn' then '查询字段信息' when 'showCreate' then '查询结构信息' when 'showIndex' then '查询索引信息' when 'showTableSize' then '查询表容量' else '其他' end as type,count(*) value from query_log group by query_type order by count(*) desc limit 30")
	queryTypePieDataList := make([]map[string]interface{}, 0)
	for _, item := range queryTypePieData {
		pieData := make(map[string]interface{})
		pieData["type"] = item["type"].(string)
		pieData["value"] = utils.StrToInt(item["value"].(string))
		queryTypePieDataList = append(queryTypePieDataList, pieData)
	}

	sensitiveDsTypePieData, _ := database.QueryAll("select distinct datasource_type type,count(*) value from sensitive_meta group by datasource_type order by count(*) desc limit 20")
	sensitiveDsTypePieDataList := make([]map[string]interface{}, 0)
	for _, item := range sensitiveDsTypePieData {
		pieData := make(map[string]interface{})
		pieData["type"] = item["type"].(string)
		pieData["value"] = utils.StrToInt(item["value"].(string))
		sensitiveDsTypePieDataList = append(sensitiveDsTypePieDataList, pieData)
	}

	sensitiveTypePieData, _ := database.QueryAll("select distinct rule_name type,count(*) value from sensitive_meta group by rule_name order by count(*) desc limit 50")
	sensitiveTypePieDataList := make([]map[string]interface{}, 0)
	for _, item := range sensitiveTypePieData {
		pieData := make(map[string]interface{})
		pieData["type"] = item["type"].(string)
		pieData["value"] = utils.StrToInt(item["value"].(string))
		sensitiveTypePieDataList = append(sensitiveTypePieDataList, pieData)
	}

	queryDayLineData, _ := database.QueryAll("select case status when 'failed' then '执行失败' when 'intercept' then '风险拦截' when 'success' then '执行成功' else '其他' end as status,date_format(gmt_created,'%Y.%m.%d') as timeline,count(*) as count from query_log where gmt_created>date_sub(now(),interval 15 day) group by status,timeline order by status,timeline asc limit 50")
	queryDayLineDataList := make([]map[string]interface{}, 0)
	for _, item := range queryDayLineData {
		queryDayLineDataList = append(queryDayLineDataList, map[string]interface{}{"time": item["timeline"], "value": utils.StrToInt(item["count"].(string)), "category": item["status"].(string)})
	}

	query15DayLineData, _ := database.QueryAll(`select a.timeline as timeline,ifnull(b.count,'0') as count,'15天查询量' as 'category' from (
		SELECT date_sub(curdate(), interval 0 day) as timeline
			union all SELECT date_sub(curdate(), interval 1 day) as timeline
			union all SELECT date_sub(curdate(), interval 2 day) as timeline
			union all SELECT date_sub(curdate(), interval 3 day) as timeline
			union all SELECT date_sub(curdate(), interval 4 day) as timeline
			union all SELECT date_sub(curdate(), interval 5 day) as timeline
			union all SELECT date_sub(curdate(), interval 6 day) as timeline
			union all SELECT date_sub(curdate(), interval 7 day) as timeline
			union all SELECT date_sub(curdate(), interval 8 day) as timeline
			union all SELECT date_sub(curdate(), interval 9 day) as timeline
			union all SELECT date_sub(curdate(), interval 10 day) as timeline
			union all SELECT date_sub(curdate(), interval 11 day) as timeline
			union all SELECT date_sub(curdate(), interval 12 day) as timeline
			union all SELECT date_sub(curdate(), interval 13 day) as timeline
			union all SELECT date_sub(curdate(), interval 14 day) as timeline
	) a
	left join (select date_format(gmt_created,'%Y-%m-%d') as timeline,count(*) as count from query_log 
	where gmt_created>date_sub(now(),interval 15 day) group by timeline order by timeline asc limit 50)  b 
	on a.timeline=b.timeline order by a.timeline asc limit 100;
	`)
	query15DayLineDataList := make([]map[string]interface{}, 0)
	for _, item := range query15DayLineData {
		query15DayLineDataList = append(query15DayLineDataList, map[string]interface{}{"time": item["timeline"].(time.Time).Format("2006-01-02"), "value": utils.StrToInt(item["count"].(string)), "category": item["category"].(string)})
	}

	query15DayInterceptLineData, _ := database.QueryAll(`select a.timeline as timeline,ifnull(b.count,'0') as count,'15天拦截量' as 'category' from (
		SELECT date_sub(curdate(), interval 0 day) as timeline
			union all SELECT date_sub(curdate(), interval 1 day) as timeline
			union all SELECT date_sub(curdate(), interval 2 day) as timeline
			union all SELECT date_sub(curdate(), interval 3 day) as timeline
			union all SELECT date_sub(curdate(), interval 4 day) as timeline
			union all SELECT date_sub(curdate(), interval 5 day) as timeline
			union all SELECT date_sub(curdate(), interval 6 day) as timeline
			union all SELECT date_sub(curdate(), interval 7 day) as timeline
			union all SELECT date_sub(curdate(), interval 8 day) as timeline
			union all SELECT date_sub(curdate(), interval 9 day) as timeline
			union all SELECT date_sub(curdate(), interval 10 day) as timeline
			union all SELECT date_sub(curdate(), interval 11 day) as timeline
			union all SELECT date_sub(curdate(), interval 12 day) as timeline
			union all SELECT date_sub(curdate(), interval 13 day) as timeline
			union all SELECT date_sub(curdate(), interval 14 day) as timeline
	) a
	left join (select date_format(gmt_created,'%Y-%m-%d') as timeline,count(*) as count from query_log 
	where gmt_created>date_sub(now(),interval 15 day)  and status='intercept' group by timeline order by timeline asc limit 50)  b 
	on a.timeline=b.timeline order by a.timeline asc limit 100;
	`)
	query15DayInterceptLineDataList := make([]map[string]interface{}, 0)
	for _, item := range query15DayInterceptLineData {
		query15DayInterceptLineDataList = append(query15DayInterceptLineDataList, map[string]interface{}{"time": item["timeline"].(time.Time).Format("2006-01-02"), "value": utils.StrToInt(item["count"].(string)), "category": item["category"].(string)})
	}

	queryMonthBarData, _ := database.QueryAll(`select a.timeline as timeline,ifnull(b.count,'0') as count,'月查询量' as 'category' from (
		select date_format(date_add(date_format(now(),'%Y-01-01'),interval row month),'%Y%m')  timeline from (
select @row := @row+1 as row from 
(select 0 union all select 1 union all  select 2 union all select 3 union all select 4 union all select 5 union all select 6 union all select 7 union all select 8 union all select 9 ) t,
(select 0 union all select 1 union all  select 2 union all select 3 union all select 4 union all select 5 union all select 6 union all select 7 union all select 8 union all select 9 ) t2,
(select @row:=-1) r
) se
where date_format(date_add('2024-01-01',interval row month),'%Y%m')<=date_format(date_format(now(),'%Y-12-01'),'%Y%m')
	) a
	left join (select date_format(gmt_created,'%Y%m') as timeline,count(*) as count from query_log  where gmt_created>date_format(now(),'%Y')   group by timeline order by timeline asc)  b 
	on a.timeline=b.timeline order by a.timeline asc limit 100
	`)
	queryMonthInterceptBarData, _ := database.QueryAll(`select a.timeline as timeline,ifnull(b.count,'0') as count,'月拦截量' as 'category' from (
		select date_format(date_add(date_format(now(),'%Y-01-01'),interval row month),'%Y%m')  timeline from (
select @row := @row+1 as row from 
(select 0 union all select 1 union all  select 2 union all select 3 union all select 4 union all select 5 union all select 6 union all select 7 union all select 8 union all select 9 ) t,
(select 0 union all select 1 union all  select 2 union all select 3 union all select 4 union all select 5 union all select 6 union all select 7 union all select 8 union all select 9 ) t2,
(select @row:=-1) r
) se
where date_format(date_add('2024-01-01',interval row month),'%Y%m')<=date_format(date_format(now(),'%Y-12-01'),'%Y%m')
	) a
	left join (select date_format(gmt_created,'%Y%m') as timeline,count(*) as count from query_log  where gmt_created>date_format(now(),'%Y') and status='intercept' group by timeline order by timeline asc)  b 
	on a.timeline=b.timeline order by a.timeline asc limit 100
	`)
	queryMonthBarDataList := make([]map[string]interface{}, 0)
	for _, item := range queryMonthBarData {
		queryMonthBarDataList = append(queryMonthBarDataList, map[string]interface{}{"time": item["timeline"].(string), "value": utils.StrToInt(item["count"].(string)), "category": item["category"].(string)})
	}

	for _, item := range queryMonthInterceptBarData {
		queryMonthBarDataList = append(queryMonthBarDataList, map[string]interface{}{"time": item["timeline"].(string), "value": utils.StrToInt(item["count"].(string)), "category": item["category"].(string)})
	}

	queryNewInterceptDataList, _ := database.QueryAll("select username,datasource_type,`database`,sql_type,result,date_format(gmt_created,'%Y-%m-%d') as gmt_created from query_log where status='intercept' order by id desc limit 10;")
	queryNewSensitiveDataList, _ := database.QueryAll("select rule_name,database_name,table_name,column_name,datasource_type,date_format(gmt_created,'%Y-%m-%d') as gmt_created from sensitive_meta where  status=1 order by id desc limit 10;")

	var data = make(map[string]interface{})

	data["todayQueryCount"] = todayQueryCount[0]["count"]
	data["totalQueryCount"] = totalQueryCount[0]["count"]
	data["totalInterceptCount"] = totalInterceptCount[0]["count"]
	data["sensitiveDatabaseCount"] = sensitiveDatabaseCount[0]["count"]
	data["sensitiveTableCount"] = sensitiveTableCount[0]["count"]
	data["sensitiveColumnCount"] = sensitiveColumnCount[0]["count"]
	data["sensitiveQueryCount"] = sensitiveQueryCount[0]["count"]

	data["queryStatusPieDataList"] = queryStatusPieDataList
	data["queryTypePieDataList"] = queryTypePieDataList
	data["sensitiveDsTypePieDataList"] = sensitiveDsTypePieDataList
	data["sensitiveTypePieDataList"] = sensitiveTypePieDataList

	data["queryDayLineDataList"] = queryDayLineDataList
	data["query15DayLineDataList"] = query15DayLineDataList
	data["query15DayInterceptLineDataList"] = query15DayInterceptLineDataList
	data["queryMonthBarDataList"] = queryMonthBarDataList

	data["queryNewInterceptDataList"] = queryNewInterceptDataList
	data["queryNewSensitiveDataList"] = queryNewSensitiveDataList

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
