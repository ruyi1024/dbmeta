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

package meta

import (
	"dbmcloud/src/database"
	"dbmcloud/src/utils"
	"net/http"
	"unicode"

	"github.com/gin-gonic/gin"
)

// 检测中文字符数量
func countChineseCharacters(text string) int {
	if text == "" {
		return 0
	}

	count := 0
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			count++
		}
	}
	return count
}

// 判断注释是否准确（包含3个及以上中文字符）
func isCommentAccurate(comment string) bool {
	return countChineseCharacters(comment) >= 3
}

// 保留2位小数
func roundToTwoDecimals(value float64) float64 {
	return float64(int(value*100)) / 100
}

func QualityInfo(c *gin.Context) {
	// 基础统计数据
	databaseCount, _ := database.QueryAll("select count(*) as count from meta_database where is_deleted = 0 limit 1")
	tableCount, _ := database.QueryAll("select count(*) as count from meta_table limit 1")
	columnCount, _ := database.QueryAll("select count(*) as count from meta_column limit 1")

	// 数据库业务关联率（在 meta_database_business 中存在 database_name 即视为已关联业务信息）
	databaseBusinessCount, _ := database.QueryAll(`
		select count(*) as count from meta_database d
		where d.is_deleted = 0
		and exists (
			select 1 from meta_database_business b where b.database_name = d.database_name limit 1
		)
		limit 1
	`)

	var databaseBusinessRate float64 = 0
	if utils.StrToInt(databaseCount[0]["count"].(string)) > 0 {
		databaseBusinessRate = float64(utils.StrToInt(databaseBusinessCount[0]["count"].(string))) / float64(utils.StrToInt(databaseCount[0]["count"].(string))) * 100
		// 保留2位小数
		databaseBusinessRate = roundToTwoDecimals(databaseBusinessRate)
	}

	// 数据表注释完备率（基于table_comment）
	tableCommentCount, _ := database.QueryAll(`
		select count(*) as count from meta_table 
		where table_comment is not null and table_comment != ''
		limit 1
	`)

	var tableCommentRate float64 = 0
	if utils.StrToInt(tableCount[0]["count"].(string)) > 0 {
		tableCommentRate = float64(utils.StrToInt(tableCommentCount[0]["count"].(string))) / float64(utils.StrToInt(tableCount[0]["count"].(string))) * 100
		// 保留2位小数
		tableCommentRate = roundToTwoDecimals(tableCommentRate)
	}

	// 数据字段注释完备率（基于column_comment）
	columnCommentCount, _ := database.QueryAll(`
		select count(*) as count from meta_column 
		where column_comment is not null and column_comment != ''
		limit 1
	`)

	var columnCommentRate float64 = 0
	if utils.StrToInt(columnCount[0]["count"].(string)) > 0 {
		columnCommentRate = float64(utils.StrToInt(columnCommentCount[0]["count"].(string))) / float64(utils.StrToInt(columnCount[0]["count"].(string))) * 100
		// 保留2位小数
		columnCommentRate = roundToTwoDecimals(columnCommentRate)
	}

	// 获取所有表注释进行准确度分析
	tableComments, _ := database.QueryAll(`
		select table_comment from meta_table 
		where table_comment is not null and table_comment != ''
	`)

	tableAccurateCount := 0
	tableGeneralCount := 0
	tableInaccurateCount := 0

	for _, item := range tableComments {
		comment := item["table_comment"].(string)
		chineseCount := countChineseCharacters(comment)
		// 新规则：纯英文则不准确，1-2个中文则一般，含2个以上中文则准确
		if chineseCount >= 3 {
			tableAccurateCount++
		} else if chineseCount >= 1 {
			tableGeneralCount++
		} else {
			tableInaccurateCount++
		}
	}

	var tableAccuracyRate float64 = 0
	if utils.StrToInt(tableCommentCount[0]["count"].(string)) > 0 {
		tableAccuracyRate = float64(tableAccurateCount) / float64(utils.StrToInt(tableCommentCount[0]["count"].(string))) * 100
		// 保留2位小数
		tableAccuracyRate = roundToTwoDecimals(tableAccuracyRate)
	}

	// 获取所有字段注释进行准确度分析
	columnComments, _ := database.QueryAll(`
		select column_comment from meta_column 
		where column_comment is not null and column_comment != ''
	`)

	columnAccurateCount := 0
	columnGeneralCount := 0
	columnInaccurateCount := 0

	for _, item := range columnComments {
		comment := item["column_comment"].(string)
		chineseCount := countChineseCharacters(comment)
		// 新规则：纯英文则不准确，1-2个中文则一般，含2个以上中文则准确
		if chineseCount >= 3 {
			columnAccurateCount++
		} else if chineseCount >= 1 {
			columnGeneralCount++
		} else {
			columnInaccurateCount++
		}
	}

	var columnAccuracyRate float64 = 0
	if utils.StrToInt(columnCommentCount[0]["count"].(string)) > 0 {
		columnAccuracyRate = float64(columnAccurateCount) / float64(utils.StrToInt(columnCommentCount[0]["count"].(string))) * 100
		// 保留2位小数
		columnAccuracyRate = roundToTwoDecimals(columnAccuracyRate)
	}

	// 数据库业务关联情况饼图数据（同关联率逻辑）
	databaseBusinessData, _ := database.QueryAll(`
		select 
			case 
				when exists (
					select 1 from meta_database_business b where b.database_name = meta_database.database_name limit 1
				) then '已关联业务'
				else '未关联业务'
			end as type,
			count(*) as value
		from meta_database
		where is_deleted = 0
		group by 
			case 
				when exists (
					select 1 from meta_database_business b where b.database_name = meta_database.database_name limit 1
				) then '已关联业务'
				else '未关联业务'
			end
	`)
	databaseQualityDataList := make([]map[string]interface{}, 0)
	for _, item := range databaseBusinessData {
		pieData := make(map[string]interface{})
		pieData["type"] = item["type"].(string)
		pieData["value"] = utils.StrToInt(item["value"].(string))
		databaseQualityDataList = append(databaseQualityDataList, pieData)
	}

	// 数据表注释完备情况饼图数据
	tableQualityData, _ := database.QueryAll(`
		select 
			case 
				when table_comment is not null and table_comment != '' then '有注释'
				else '无注释'
			end as type,
			count(*) as value
		from meta_table 
		group by type
	`)
	tableQualityDataList := make([]map[string]interface{}, 0)
	for _, item := range tableQualityData {
		pieData := make(map[string]interface{})
		pieData["type"] = item["type"].(string)
		pieData["value"] = utils.StrToInt(item["value"].(string))
		tableQualityDataList = append(tableQualityDataList, pieData)
	}

	// 数据字段注释完备情况饼图数据
	columnQualityData, _ := database.QueryAll(`
		select 
			case 
				when column_comment is not null and column_comment != '' then '有注释'
				else '无注释'
			end as type,
			count(*) as value
		from meta_column 
		group by type
	`)
	columnQualityDataList := make([]map[string]interface{}, 0)
	for _, item := range columnQualityData {
		pieData := make(map[string]interface{})
		pieData["type"] = item["type"].(string)
		pieData["value"] = utils.StrToInt(item["value"].(string))
		columnQualityDataList = append(columnQualityDataList, pieData)
	}

	// 表注释准确度分布（实际计算）
	tableCommentAccuracyDataList := []map[string]interface{}{
		{"type": "准确", "value": tableAccurateCount},
		{"type": "一般", "value": tableGeneralCount},
		{"type": "不准确", "value": tableInaccurateCount},
	}

	// 字段注释准确度分布（实际计算）
	columnCommentAccuracyDataList := []map[string]interface{}{
		{"type": "准确", "value": columnAccurateCount},
		{"type": "一般", "value": columnGeneralCount},
		{"type": "不准确", "value": columnInaccurateCount},
	}

	var data map[string]interface{}
	data = make(map[string]interface{})
	data["databaseCount"] = databaseCount[0]["count"]
	data["tableCount"] = tableCount[0]["count"]
	data["columnCount"] = columnCount[0]["count"]
	data["databaseBusinessRate"] = databaseBusinessRate
	data["tableCommentRate"] = tableCommentRate
	data["columnCommentRate"] = columnCommentRate
	data["tableAccuracyRate"] = tableAccuracyRate
	data["columnAccuracyRate"] = columnAccuracyRate
	data["databaseQualityDataList"] = databaseQualityDataList
	data["tableQualityDataList"] = tableQualityDataList
	data["columnQualityDataList"] = columnQualityDataList
	data["tableCommentAccuracyDataList"] = tableCommentAccuracyDataList
	data["columnCommentAccuracyDataList"] = columnCommentAccuracyDataList

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
