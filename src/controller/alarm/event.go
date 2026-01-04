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

package alarm

import (
	"bytes"
	"dbmcloud/log"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func EventList(c *gin.Context) {
	var db = database.DB

	if c.Query("event_type") != "" {
		db = db.Where("event_type=?", c.Query("event_type"))
	}
	if c.Query("event_group") != "" {
		db = db.Where("event_group=?", c.Query("event_group"))
	}
	if c.Query("event_entity") != "" {
		db = db.Where("event_entity=?", c.Query("event_entity"))
	}
	if c.Query("event_key") != "" {
		db = db.Where("event_key=?", c.Query("event_key"))
	}
	if c.Query("send_mail") != "" {
		db = db.Where("send_mail=?", c.Query("send_mail"))
	}
	if c.Query("send_phone") != "" {
		db = db.Where("send_phone=?", c.Query("send_phone"))
	}
	if c.Query("start_time") != "" && c.Query("end_time") != "" {
		db = db.Where("event_time >= ?", c.Query("start_time"))
		db = db.Where("event_time <= ?", c.Query("end_time"))
	}
	var count int64
	result := db.Model(&model.AlarmEvent{}).Count(&count)
	if result.Error != nil {
		c.JSON(200, gin.H{"success": false, "msg": "Query Error:" + result.Error.Error()})
		return
	}
	db = db.Order("id desc")
	if c.Query("current") != "" && c.Query("pageSize") != "" {
		offset, _ := strconv.Atoi(c.DefaultQuery("current", "0"))
		limit, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
		if offset <= 1 {
			offset = 0
		} else {
			offset--
		}
		db.Offset(offset * limit).Limit(limit)
	}
	var dataList []model.AlarmEvent
	result = db.Find(&dataList)
	if result.Error != nil {
		c.JSON(200, gin.H{"success": false, "msg": "Query Error:" + result.Error.Error()})
		return
	}
	log.Info("----> ", zap.Int64("count", count))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
		"data":    dataList,
		"total":   count,
	})
	return

}

type Result struct {
	EventType   string
	EventKey    string
	EventDetail string
}

type Suggest struct {
	Content string
}

// 去除json中的转义字符
func disableEscapeHtml(data interface{}) (string, error) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	if err := jsonEncoder.Encode(data); err != nil {
		return "", err
	}
	return bf.String(), nil
}

type EventDescription struct {
	Description string
}

type AlarmSuggest struct {
	Content string
}

func EventDetail(c *gin.Context) {
	uuid := c.Query("uuid")
	if uuid == "" {
		c.JSON(200, gin.H{"success": false, "msg": "Params Error."})
		return
	}

	var db = database.DB
	var result Result
	db.Raw("select event_type,event_key,event_detail from alarm_event where event_uuid=? ", uuid).Scan(&result)
	var resultJson []map[string]interface{}
	detail := strings.Replace(result.EventDetail, "\n", "\\n", -1)
	err := json.Unmarshal([]byte(detail), &resultJson)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(resultJson)
	columns := make([]map[string]string, 0)
	if len(resultJson) > 0 {
		for key, _ := range resultJson[0] {
			columns = append(columns, map[string]string{"title": key, "dataIndex": key})
		}
	}
	/*/
	get event description
	*/
	var eventDescription EventDescription
	db.Raw("select description from event_description where event_type=?  and event_key=? limit 1", result.EventType, result.EventKey).Scan(&eventDescription)
	/*/
	get alarm suggest
	*/
	var alarmSuggest AlarmSuggest
	db.Raw("select content from alarm_suggest where event_type=?  and event_key=? limit 1", result.EventType, result.EventKey).Scan(&alarmSuggest)
	//fmt.Println(columns)
	//fmt.Println(resultJson)
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"msg":         "OK",
		"data":        resultJson,
		"total":       len(resultJson),
		"columns":     columns,
		"description": eventDescription.Description,
		"suggest":     alarmSuggest.Content,
	})
	return

}

type updStatusData struct {
	Ids    []int64 `json:"ids"`
	Status int     `json:"status"`
}

func PutBatchUpdateStatus(c *gin.Context) {
	var updData updStatusData
	err := c.BindJSON(&updData)
	if err != nil {
		c.JSON(200, gin.H{"success": false, "msg": "update body error. " + err.Error()})
		return
	}

	result := database.DB.Model(&model.AlarmEvent{}).Where("id IN (?)", updData.Ids).Updates(map[string]interface{}{"status": updData.Status})
	if result.Error != nil {
		c.JSON(200, gin.H{"success": false, "msg": "update db users error " + result.Error.Error()})
		return
	}
	content := "开始跟进"
	if updData.Status == 2 {
		content = "处理完成"
	}
	for k, _ := range updData.Ids {
		var tracks model.AlarmTrack
		userId, _ := c.Get("userId")
		tracks.UserId, _ = userId.(int64)
		tracks.AlarmId = updData.Ids[k]
		tracks.Content = content
		database.DB.Create(&tracks)
	}
	c.JSON(200, gin.H{"success": true})
	return
}

func EventAnalysis(c *gin.Context) {
	alarmCount, _ := database.QueryAll("SELECT COUNT(*) as count FROM alarm_event LIMIT 1")
	alarmLastTime, _ := database.QueryAll("SELECT gmt_created FROM alarm_event order by id desc LIMIT 1 ")
	alarmTodayCount, _ := database.QueryAll("SELECT COUNT(*) as count FROM alarm_event WHERE gmt_created>= DATE_FORMAT(NOW(), '%Y-%m-%d 00:00:00') LIMIT 1")
	alarmHourCount, _ := database.QueryAll("SELECT COUNT(*) as count FROM alarm_event WHERE gmt_created>= date_sub(NOW(),interval 1 hour) LIMIT 1")
	alarmLast7DayData, _ := database.QueryAll("select date_format(gmt_created,'%Y%m%d') as x,count(*) as y from alarm_event where gmt_created>= date_format( date_sub(NOW(), interval 14 day) ,'%Y-%m-%d 00:00:00') group by x order by x asc")
	alarmTodayData, _ := database.QueryAll("select date_format(gmt_created,'%H') as x,count(*) as y from alarm_event where gmt_created>= date_format(now() ,'%Y-%m-%d 00:00:00') group by x order by x asc")
	alarmTagData, _ := database.QueryAll("select concat(event_type,'[',event_key,']') as name,200 as value from alarm_event where gmt_created>= date_format( date_sub(NOW(), interval 7 day) ,'%Y-%m-%d 00:00:00') group by name order by value desc")
	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"msg":               "OK",
		"alarmCount":        alarmCount[0]["count"],
		"alarmLastTime":     alarmLastTime[0]["gmt_created"],
		"alarmTodayCount":   alarmTodayCount[0]["count"],
		"alarmHourCount":    alarmHourCount[0]["count"],
		"alarmLast7DayData": alarmLast7DayData,
		"alarmTodayData":    alarmTodayData,
		"alarmTagData":      alarmTagData,
	})
	return
}
