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
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func LevelList(c *gin.Context) {
	var db = database.DB
	method := c.Request.Method
	if method == "GET" {
		if c.Query("enable") != "" {
			db = db.Where("enable=?", c.Query("enable"))
		}
		if c.Query("level_name") != "" {
			db = db.Where("name like ? ", "%"+c.Query("name")+"%")
		}
		sorterMap := make(map[string]string)
		sorterData := c.Query("sorter")
		json.Unmarshal([]byte(sorterData), &sorterMap)
		for sortField, sortOrder := range sorterMap {
			if sortField != "" && sortOrder != "" {
				db = db.Order(fmt.Sprintf("%s %s", sortField, strings.Replace(sortOrder, "end", "", 1)))
			}
		}

		var dataList []model.AlarmLevel
		result := db.Find(&dataList)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Query Error:" + result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "OK",
			"data":    dataList,
			"total":   len(dataList),
		})
		return

	}
	if method == "POST" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		fmt.Print(params)
		data := model.AlarmLevel{LevelName: params["level_name"].(string), Description: params["description"].(string), Enable: int(params["enable"].(float64))}
		result := db.Create(&data)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Insert Error:" + result.Error.Error()})
			return
		}

	}
	if method == "PUT" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		result := db.Model(&model.AlarmLevel{}).Where("id=?", params["id"].(float64)).Updates(map[string]interface{}{"LevelName": params["level_name"].(string), "Description": params["description"].(string), "Enable": int(params["enable"].(float64))})
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Update Error:" + result.Error.Error()})
			return
		}
	}
	if method == "DELETE" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		result := db.Where("id = ?", params["id"].(float64)).Delete(&model.AlarmLevel{})
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Delete Error:" + result.Error.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
	})
	return
}
