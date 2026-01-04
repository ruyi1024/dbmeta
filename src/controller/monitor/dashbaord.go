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

package monitor

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

		//近1分钟事件数
		var currentEventCount int64
		ckdb.Model(&model.Event{}).Where("event_time>addMinutes(now(),-1)").Count(&currentEventCount)
		d["currentEventCount"] = currentEventCount

		var eventList []model.Event
		ckdb.Raw("select event_time,event_type,event_group,event_entity,event_key,event_value from events order by event_time desc limit 8").Scan(&eventList)

		//实时事件数量
		var eventCount int64
		ckdb.Model(&model.Event{}).Where("event_time>addMinutes(now(),-1)").Count(&eventCount)
		d["eventCount"] = eventCount

		//小时事件数量
		var eventHourCount int64
		ckdb.Model(&model.Event{}).Where("event_time>addMinutes(now(),-60)").Count(&eventHourCount)
		d["eventHourCount"] = eventHourCount

		//最新事件时间
		var eventRecord map[string]interface{}
		ckdb.Model(&model.Event{}).Order("event_time desc").Take(&eventRecord)
		d["lastEventTime"] = eventRecord["event_time"]

		//每分钟实时事件图表
		type EventLine struct {
			X string
			Y int
		}
		var eventLineResult []EventLine
		ckdb.Model(&model.Event{}).Select("toString(toStartOfInterval(event_time,interval 1 minute)) as x ,count(*) as y").Where("event_time >addMinutes(now(),-10)").Group("x").Order("x asc").Find(&eventLineResult)
		lineDataList := make([]map[string]interface{}, 0)
		for _, item := range eventLineResult {
			lineDataList = append(lineDataList, map[string]interface{}{"x": item.X, "y": item.Y})
		}
		d["eventMinuteData"] = lineDataList

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

	var data map[string]interface{}
	data = make(map[string]interface{})

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
	data["eventPieChartData"] = pieDataList

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
	data["eventLineChartData"] = lineDataList

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
