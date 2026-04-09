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

package dashboard

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// var db = database.InitConnect()

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//允许跨域访问
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func EventWS(c *gin.Context) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		//fmt.Printf("websocket read message:  %s\n" ,msg)

		var mydb = database.DB
		var ckdb = database.CK

		var d = make(map[string]interface{})

		//监控节点数
		var datasourceCount int64
		mydb.Model(&model.Datasource{}).Where("monitor_enable=?", 1).Count(&datasourceCount)
		d["datasourceCount"] = datasourceCount

		//今日事件数
		var todayEventCount int64
		ckdb.Model(&model.Event{}).Where("event_time>toStartOfDay(toDate(now()))").Count(&todayEventCount)
		d["todayEventCount"] = todayEventCount

		//今日告警数
		var todayAlarmCount int64
		mydb.Model(&model.AlarmEvent{}).Where("gmt_created>date_format(now(),'%Y-%m-%d')").Count(&todayAlarmCount)
		d["todayAlarmCount"] = todayAlarmCount

		//今日SQL查询数
		var todaySqlQueryCount int64
		mydb.Model(&model.QueryLog{}).Where("gmt_created>date_format(now(),'%Y-%m-%d')").Count(&todaySqlQueryCount)
		d["todaySqlQueryCount"] = todaySqlQueryCount

		//今日SQL查询拦截数
		var todaySqlQueryInterceptCount int64
		mydb.Model(&model.QueryLog{}).Where("gmt_created>date_format(now(),'%Y-%m-%d')").Where("status=?", "intercept").Count(&todaySqlQueryInterceptCount)
		d["todaySqlQueryInterceptCount"] = todaySqlQueryInterceptCount

		//数据保护次数
		sensitiveQueryCount, _ := database.QueryAll("select count(*) as count from query_log where `database` in (select database_name from sensitive_meta) limit 1;")
		d["sensitiveQueryCount"] = sensitiveQueryCount[0]["count"]

		//近15分钟事件数
		var currentEventCount int64
		ckdb.Model(&model.Event{}).Where("event_time>date_sub(now(),interval 15 minute)").Count(&currentEventCount)
		d["currentEventCount"] = currentEventCount

		//近15分钟告警数
		var currentAlarmCount int64
		mydb.Model(&model.AlarmEvent{}).Where("gmt_created>date_sub(now(),interval 15 minute)").Count(&currentAlarmCount)
		d["currentAlarmCount"] = currentAlarmCount

		//近1小时告警通知数
		var currentAlarmNoticeCount int64
		mydb.Model(&model.AlarmSendLog{}).Where("gmt_created>date_sub(now(),interval 60 minute)").Count(&currentAlarmNoticeCount)
		d["currentAlarmNoticeCount"] = currentAlarmNoticeCount

		var eventList []model.Event
		ckdb.Raw("select event_time,event_type,event_group,event_entity,event_key,event_value from events order by event_time desc limit 8").Scan(&eventList)

		//小时事件数量
		var eventHourCount int64
		ckdb.Model(&model.Event{}).Where("event_time>addMinutes(now(),-60)").Count(&eventHourCount)
		d["eventHourCount"] = eventHourCount

		//最新事件时间
		var eventRecord map[string]interface{}
		result := ckdb.Model(&model.Event{}).Order("event_time desc").Take(&eventRecord)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			d["lastEventTime"] = eventRecord["event_time"].(time.Time).Format("2006-01-02 15:04:05")
		} else {
			d["lastEventTime"] = "0000-00-00 00:00:00"
		}

		//最新告警时间
		var alarmRecord map[string]interface{}
		result = mydb.Model(&model.AlarmEvent{}).Order("gmt_created desc").Take(&alarmRecord)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			d["lastAlarmTime"] = alarmRecord["gmt_created"].(time.Time).Format("2006-01-02 15:04:05")
		} else {
			d["lastAlarmTime"] = "0000-00-00 00:00:00"
		}

		//最新查询时间
		var queryRecord map[string]interface{}
		result = mydb.Model(&model.QueryLog{}).Order("id desc").Take(&queryRecord)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			d["lastQueryTime"] = queryRecord["gmt_created"].(time.Time).Format("2006-01-02 15:04:05")
		} else {
			d["lastQueryTime"] = "0000-00-00 00:00:00"
		}

		//最新拦截时间
		var queryInterceptRecord map[string]interface{}
		result = mydb.Model(&model.QueryLog{}).Where("status=?", "intercept").Order("id desc").Take(&queryInterceptRecord)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			d["lastQueryInterceptTime"] = queryInterceptRecord["gmt_created"].(time.Time).Format("2006-01-02 15:04:05")
		} else {
			d["lastQueryInterceptTime"] = "0000-00-00 00:00:00"
		}

		//近15分钟每分钟实时事件图表
		type EventLine struct {
			X string
			Y int
		}
		var eventLineResult []EventLine
		ckdb.Model(&model.Event{}).Select("toString(toStartOfInterval(event_time,interval 1 minute)) as x ,count(*) as y").Where("event_time >addMinutes(now(),-15)").Group("x").Order("x asc").Find(&eventLineResult)
		lineDataList := make([]map[string]interface{}, 0)
		for _, item := range eventLineResult {
			lineDataList = append(lineDataList, map[string]interface{}{"x": item.X, "y": item.Y})
		}
		d["eventMinuteData"] = lineDataList

		//今日每分钟实时告警图表
		type AlarmLine struct {
			X string
			Y int
		}
		var alarmLineResult []AlarmLine
		mydb.Model(&model.AlarmEvent{}).Select("date_format(event_time,'%Y-%m-%d %H:%i:00') as x ,count(*) as y").Where("gmt_created>date_format(now(),'%Y-%m-%d')").Group("x").Order("x asc").Find(&alarmLineResult)
		lineDataList = make([]map[string]interface{}, 0)
		for _, item := range alarmLineResult {
			lineDataList = append(lineDataList, map[string]interface{}{"x": item.X, "y": item.Y})
		}
		d["alarmMinuteData"] = lineDataList

		//今日每分钟查询量图表
		type QueryNumberLine struct {
			X string
			Y int
		}
		var queryNumberResult []QueryNumberLine
		mydb.Model(&model.QueryLog{}).Select("date_format(gmt_created,'%Y-%m-%d %H:%i:00') as x ,count(*) as y").Where("gmt_created>date_format(now(),'%Y-%m-%d')").Group("x").Order("x asc").Find(&queryNumberResult)
		lineDataList = make([]map[string]interface{}, 0)
		for _, item := range queryNumberResult {
			lineDataList = append(lineDataList, map[string]interface{}{"x": item.X, "y": item.Y})
		}
		d["queryNumberTodayData"] = lineDataList

		//今日每分钟查询拦截量图表
		type QueryInterceptLine struct {
			X string
			Y int
		}
		var queryInterceptResult []QueryInterceptLine
		mydb.Model(&model.QueryLog{}).Select("date_format(gmt_created,'%Y-%m-%d %H:%i:00') as x ,count(*) as y").Where("gmt_created>date_format(now(),'%Y-%m-%d')").Where("status=?", "intercept").Group("x").Order("x asc").Find(&queryInterceptResult)
		lineDataList = make([]map[string]interface{}, 0)
		for _, item := range queryInterceptResult {
			lineDataList = append(lineDataList, map[string]interface{}{"x": item.X, "y": item.Y})
		}
		d["queryInterceptTodayData"] = lineDataList

		//1小时内每5分钟实时事件图表
		type EventHourLine struct {
			X string
			Y int
		}
		var eventhourLineResult []EventHourLine
		ckdb.Model(&model.Event{}).Select("toString(toStartOfInterval(event_time,interval 5 minute)) as x ,count(*) as y").Where("event_time >addMinutes(now(),-60)").Group("x").Order("x asc").Find(&eventhourLineResult)
		lineHourDataList := make([]map[string]interface{}, 0)
		for _, item := range eventhourLineResult {
			lineHourDataList = append(lineHourDataList, map[string]interface{}{"x": item.X, "y": item.Y})
		}
		d["eventHourData"] = lineHourDataList

		//小时和分钟指标distinct值
		var hourKeyCount, minKeyCount int64
		ckdb.Model(&model.Event{}).Where("event_time>addMinutes(now(),-60)").Group("event_entity,event_key").Count(&hourKeyCount)
		ckdb.Model(&model.Event{}).Where("event_time>addMinutes(now(),-1)").Group("event_entity,event_key").Count(&minKeyCount)
		d["hourKeyCount"] = hourKeyCount
		d["minKeyCount"] = minKeyCount
		if minKeyCount == 0 || hourKeyCount == 0 {
			d["eventKeyPct"] = 0
			d["eventKeyPctStr"] = "0%"
		} else {
			d["eventKeyPct"] = int((float64(minKeyCount) / float64(hourKeyCount)) * 100)
			d["eventKeyPctStr"] = fmt.Sprintf("%d%%", int((float64(minKeyCount)/float64(hourKeyCount))*100))
		}

		//查询拦截占比
		if todaySqlQueryCount == 0 || todaySqlQueryInterceptCount == 0 {
			d["queryInterceptPct"] = 0
			d["queryInterceptPctStr"] = "0%"
		} else {
			d["queryInterceptPct"] = int((float64(todaySqlQueryInterceptCount) / float64(todaySqlQueryCount)) * 100)
			d["queryInterceptPctStr"] = fmt.Sprintf("%d%%", int((float64(todaySqlQueryInterceptCount)/float64(todaySqlQueryCount))*100))
		}

		d["receiveMessage"] = msg
		d["eventList"] = eventList

		//健康状态
		var totalDatasourceCount, monitorDatasourceCount, healthDatasourceCount, failDatasourceCount int64
		mydb.Model(&model.Datasource{}).Count(&totalDatasourceCount)
		mydb.Model(&model.Datasource{}).Where("monitor_enable=?", 1).Count(&monitorDatasourceCount)
		mydb.Model(&model.Datasource{}).Where("monitor_enable=?", 1).Where("status=?", 1).Count(&healthDatasourceCount)
		mydb.Model(&model.Datasource{}).Where("monitor_enable=?", 1).Where("status=?", 0).Count(&failDatasourceCount)
		d["totalDatasourceCount"] = totalDatasourceCount
		d["monitorDatasourceCount"] = monitorDatasourceCount
		d["healthDatasourceCount"] = healthDatasourceCount
		d["failDatasourceCount"] = failDatasourceCount
		if monitorDatasourceCount == 0 || healthDatasourceCount == 0 {
			d["healthPct"] = 0
			d["healthPct2"] = 0
		} else {
			d["healthPct"] = float64(healthDatasourceCount) / float64(monitorDatasourceCount)
			d["healthPct2"] = int((float64(healthDatasourceCount) / float64(monitorDatasourceCount)) * 100)
		}

		j, _ := json.Marshal(d)
		conn.WriteMessage(t, []byte(j))
	}

}

func MetaInfo(c *gin.Context) {

	var d = make(map[string]interface{})

	// nodeCount, _ := database.QueryAll("select count(*) as count from datasource limit 1")
	// clusterCount, _ := database.QueryAll("select count(*) as count from meta_clusters limit 1")
	// hostCount, _ := database.QueryAll("select count(*) as count from meta_hosts limit 1")
	// idcCount, _ := database.QueryAll("select count(*) as count from meta_idcs limit 1")
	// envCount, _ := database.QueryAll("select count(*) as count from meta_envs limit 1")
	// moduleCount, _ := database.QueryAll("select count(*) as count from meta_modules limit 1")
	// fmt.Println(nodeCount)
	// var data map[string]interface{}
	// data = make(map[string]interface{})
	// data["nodeCount"] = nodeCount[0]["count"]
	// data["clusterCount"] = clusterCount[0]["count"]
	// data["hostCount"] = hostCount[0]["count"]
	// data["idcCount"] = idcCount[0]["count"]
	// data["envCount"] = envCount[0]["count"]
	// data["moduleCount"] = moduleCount[0]["count"]

	//var mydb = database.DB
	var ckdb = database.CK

	//pie chart data
	type EventPie struct {
		EventType string
		Count     int
	}
	var eventPieResult []EventPie
	ckdb.Model(&model.Event{}).Select("event_type as event_type,count(*) as count").Where("event_time>addMinutes(now(),-5)").Group("event_type").Find(&eventPieResult)
	pieDataList := make([]map[string]interface{}, 0)
	for _, item := range eventPieResult {
		pieData := make(map[string]interface{})
		//pieData[item["type"].(string)] = utils.StrToInt(item["value"].(string))
		pieData["type"] = item.EventType
		pieData["value"] = item.Count
		pieDataList = append(pieDataList, pieData)
	}
	d["eventPieChartData"] = pieDataList

	//line chart data
	type EventLine struct {
		EventType string
		LineTime  string
		Count     int
	}
	var eventLineResult []EventLine
	ckdb.Model(&model.Event{}).Select("toString(toStartOfInterval(event_time,interval 1 minute)) as line_time ,event_type as event_type,count(*) as count").Where("event_time >addHours(now(),-1)").Group("line_time,event_type").Order("line_time asc").Find(&eventLineResult)
	//fmt.Println(eventLineResult)
	lineDataList := make([]map[string]interface{}, 0)
	for _, item := range eventLineResult {
		lineDataList = append(lineDataList, map[string]interface{}{"time": item.LineTime, "value": item.Count, "category": item.EventType})
	}
	d["eventLineChartData"] = lineDataList

	datasourcePieData, _ := database.QueryAll("select type,count(*) value from datasource  group by type order by value desc limit 30")
	datasourcePieDataList := make([]map[string]interface{}, 0)
	for _, item := range datasourcePieData {
		pieData := make(map[string]interface{})
		pieData["type"] = item["type"].(string)
		pieData["value"] = utils.StrToInt(item["value"].(string))
		datasourcePieDataList = append(datasourcePieDataList, pieData)
	}
	d["datasourcePieDataList"] = datasourcePieDataList

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    d,
	})
}
