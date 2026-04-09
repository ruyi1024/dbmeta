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

package event

import (
	"bytes"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

/*
init clickhouse db
*/

func List(c *gin.Context) {
	//var db = database.CK
	//var extWhere string

	// countSql := fmt.Sprintf("select count(*) from events %s", extWhere)
	// data, err := database.QueryAll(countSql)

	// count := data[0]["count()"]

	// sql := fmt.Sprintf("select event_time,event_type,event_group,event_entity,event_tag,event_key,toString(event_value) event_value,event_unit,event_uuid from events %s", extWhere)
	// fmt.Println(sql)
	// dataList, err := database.QueryAll(sql)
	// if err != nil {
	// 	c.JSON(200, gin.H{"success": false, "errorMsg": err})
	// 	return
	// }

	var db = database.CK
	var dataList = []map[string]interface{}{}
	if c.Query("typeKeyword") != "" {
		db = db.Where("event_type=? ", c.Query("typeKeyword"))
	}
	if c.Query("groupKeyword") != "" {
		db = db.Where("event_group=? ", c.Query("groupKeyword"))
	}
	if c.Query("startTime") != "" && c.Query("endTime") != "" {
		db = db.Where("event_time>=? and event_time<=?", c.Query("startTime"), c.Query("endTime"))
	} else {
		db = db.Where("event_time>addMinutes(now(),-10) ")
	}
	if c.Query("eventEntityKeyword") != "" {
		db = db.Where("event_entity in ? ", strings.Split(string(c.Query("eventEntityKeyword")), ","))
	}
	if c.Query("eventKeyKeyword") != "" {
		db = db.Where("event_key in ? ", strings.Split(string(c.Query("eventKeyKeyword")), ","))
	}

	var count int64
	db.Model(&model.Event{}).Count(&count)

	if c.Query("sorterField") != "" {
		if c.DefaultQuery("sorterOrder", "") == "ascend" {
			db = db.Order(fmt.Sprintf("%s asc", c.Query("sorterField")))
		}
		if c.DefaultQuery("sorterOrder", "") == "descend" {
			db = db.Order(fmt.Sprintf("%s desc", c.Query("sorterField")))
		}
	}
	if c.Query("limit") != "" && c.Query("offset") != "" {
		db = db.Limit(utils.StrToInt(c.Query("limit"))).Offset(utils.StrToInt(c.Query("offset")))
	} else {
		db = db.Limit(25)
	}

	db.Model(&model.Event{}).Find(&dataList)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"errorCode": 0,
		"data":      dataList,
		"total":     count,
	})
	return
}

func GetAllEventInfoList(c *gin.Context) {
	var db = database.DB
	var dataList []model.EventsDescription
	result := db.Table("event_description").Find(&dataList)

	if result.Error != nil {
		c.JSON(200, gin.H{"success": false, "errorMsg": "Query Error:" + result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"errorCode": 0,
		"data":      dataList,
	})
	return
}

func FilterItems(c *gin.Context) {
	var db = database.DB
	var metaIdc []model.Idc
	var metaEnv []model.Env
	var metaHosts []model.Datasource
	var metaClusters []model.Datasource
	var metaNodes []model.Datasource
	var metaModule []model.DatasourceType
	db.Find(&metaIdc)
	db.Find(&metaEnv)
	//db.Find(&metaHosts)
	//db.Find(&metaClusters)
	db.Find(&metaNodes)
	db.Find(&metaModule)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"errorCode": 0,
		"data": map[string]interface{}{
			"metaIdc":      metaIdc,
			"metaEnv":      metaEnv,
			"metaHosts":    metaHosts,
			"metaClusters": metaClusters,
			"metaNodes":    metaNodes,
			"metaModule":   metaModule,
		},
	})
	return
}

type chartsData struct {
	Time   string `json:"time"`
	Number int64  `json:"number"`
}

func Charts(c *gin.Context) {

	var db = database.CK
	var dataList = []map[string]interface{}{}

	var extWhere string
	if c.Query("startTime") != "" && c.Query("endTime") != "" {
		extWhere = fmt.Sprintf("%s where event_time>='%s' and event_time<='%s' ", extWhere, c.Query("startTime"), c.Query("endTime"))
	} else {
		extWhere = fmt.Sprintf("%s where event_time>addHours(now(),-1) ", extWhere)
	}
	if c.Query("groupKeyword") != "" {
		extWhere = fmt.Sprintf("%s and event_group='%s' ", extWhere, c.Query("groupKeyword"))
	}
	if c.Query("typeKeyword") != "" {
		extWhere = fmt.Sprintf("%s and event_type='%s' ", extWhere, c.Query("typeKeyword"))
	}
	if c.Query("eventEntityKeyword") != "" {
		eventEntityKeywordStrList := "'" + c.Query("eventEntityKeyword") + "'"
		eventEntityKeywordStrList = strings.Replace(eventEntityKeywordStrList, ",", "','", -1)
		extWhere = fmt.Sprintf("%s and event_entity in (%s) ", extWhere, eventEntityKeywordStrList)
	}
	if c.Query("eventKeyKeyword") != "" {
		eventKeyKeywordStrList := "'" + c.Query("eventKeyKeyword") + "'"
		eventKeyKeywordStrList = strings.Replace(eventKeyKeywordStrList, ",", "','", -1)
		extWhere = fmt.Sprintf("%s and event_key in (%s) ", extWhere, eventKeyKeywordStrList)
	}
	// type Result struct {
	// 	EventType   string
	// 	EventKey    string
	// 	EventDetail string
	// }
	// var dataList Result
	sql := fmt.Sprintf("select toString(toStartOfInterval(event_time,interval 3 minute)) as time ,toString(count(*) as number)  from events %s group by time order by time asc ", extWhere)
	fmt.Println(sql)
	db.Raw(sql).Scan(&dataList)
	//db.Model(&model.Event{}).Find(&dataList)

	if len(dataList) == 0 {
		dataList = make([]map[string]interface{}, 0)
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"errorCode": 0,
		"data":      dataList,
	})
	return
}

type chartsFullData struct {
	// Name   string  `json:"name"`
	Time   string `json:"time"`
	Number string `json:"number"`
}

func ChartsFull(c *gin.Context) {

	var db = database.CK
	var dataList []model.Event
	if c.Query("typeKeyword") != "" {
		db = db.Where("event_type=? ", c.Query("typeKeyword"))
	}
	if c.Query("groupKeyword") != "" {
		db = db.Where("event_group=? ", c.Query("groupKeyword"))
	}
	if c.Query("startTime") != "" && c.Query("endTime") != "" {
		db = db.Where("event_time>=? and event_time<=?", c.Query("startTime"), c.Query("endTime"))
	} else {
		db = db.Where("event_time>addMinutes(now(),-10) ")
	}
	if c.Query("eventEntityKeyword") != "" {
		db = db.Where("event_entity in ? ", strings.Split(string(c.Query("eventEntityKeyword")), ","))
	}
	if c.Query("eventKeyKeyword") != "" {
		db = db.Where("event_key in ? ", strings.Split(string(c.Query("eventKeyKeyword")), ","))
	}
	db.Model(&model.Event{}).Find(&dataList)

	//转化为维度系列数据
	var tmp = make(map[string]map[string][]interface{})
	for _, v := range dataList {
		eventTime := v.EventTime.Format("2006-01-02 15:04:05")
		key := v.EventKey
		if v.EventTag != "" {
			key = v.EventKey + ": " + v.EventTag
		}
		subKey := strings.Join([]string{v.EventType, v.EventGroup, v.EventEntity}, " ")
		if tmp[key] == nil {
			tmp[key] = make(map[string][]interface{})
		}
		if tmp[key][subKey] == nil {
			tmp[key][subKey] = []interface{}{}
		}
		tmp[key][subKey] = append(tmp[key][subKey], map[string]interface{}{
			"time":   eventTime,
			"number": v.EventValue,
			"unit":   v.EventUnit,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"errorCode": 0,
		"data":      tmp,
	})
}

type Result struct {
	EventType   string
	EventKey    string
	EventDetail string
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

func EventDetail(c *gin.Context) {
	uuid := c.Query("uuid")
	if uuid == "" {
		c.JSON(200, gin.H{"success": false, "msg": "Params Error."})
		return
	}

	var ckdb = database.CK
	var result Result
	ckdb.Raw("select event_type,event_key,event_detail from events where event_uuid=? ", uuid).Scan(&result)
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
	var mydb = database.DB
	var eventDescription EventDescription
	mydb.Raw("select description from event_description where event_type=?  and event_key=? limit 1", result.EventType, result.EventKey).Scan(&eventDescription)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"msg":         "OK",
		"data":        resultJson,
		"total":       len(resultJson),
		"columns":     columns,
		"description": eventDescription.Description,
	})
	return

}

func TypeList(c *gin.Context) {
	var db = database.CK
	type result struct {
		EventType string `json:"event_type"`
	}
	var dataList []result
	db.Model(&model.Event{}).Select("event_type").Where("event_time>addHours(now(),-1)").Group("event_type").Find(&dataList)
	c.JSON(200, gin.H{"success": true, "data": dataList, "total": len(dataList)})
	return
}

func GroupList(c *gin.Context) {
	var db = database.CK
	type result struct {
		EventGroup string `json:"event_group"`
	}
	var dataList []result
	db.Model(&model.Event{}).Select("event_group").Where("event_time>addHours(now(),-1)").Where("event_type=?", c.Query("event_type")).Group("event_group").Find(&dataList)
	c.JSON(200, gin.H{"success": true, "data": dataList, "total": len(dataList)})
	return
}

func EntityList(c *gin.Context) {
	var db = database.CK
	type result struct {
		EventEntity string `json:"event_entity"`
	}
	var dataList []result
	db.Model(&model.Event{}).Select("event_entity").Where("event_time>addHours(now(),-1)").Where("event_type=?", c.Query("event_type")).Where("event_group=?", c.Query("event_group")).Group("event_entity").Find(&dataList)
	c.JSON(200, gin.H{"success": true, "data": dataList, "total": len(dataList)})
	return
}

func KeyList(c *gin.Context) {

	var db = database.CK
	type result struct {
		EventKey string `json:"event_key"`
	}
	var dataList []result
	db.Model(&model.Event{}).Select("event_key").Where("event_time>addHours(now(),-1)").Where("event_type=?", c.Query("event_type")).Where("event_group=?", c.Query("event_group")).Group("event_key").Find(&dataList)
	c.JSON(200, gin.H{"success": true, "data": dataList, "total": len(dataList)})
	return
}
