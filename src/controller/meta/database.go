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
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DatabaseList(c *gin.Context) {
	var db = database.DB
	method := c.Request.Method
	if method == "GET" {
		// 默认过滤已删除的记录
		if c.Query("is_deleted") == "" {
			db = db.Where("is_deleted = ?", 0)
		} else {
			db = db.Where("is_deleted = ?", c.Query("is_deleted"))
		}
		if c.Query("datasource_type") != "" {
			db = db.Where("datasource_type=?", c.Query("datasource_type"))
		}
		if c.Query("host") != "" {
			db = db.Where("host=?", c.Query("host"))
		}
		if c.Query("port") != "" {
			db = db.Where("port=?", c.Query("port"))
		}
		if c.Query("database_name") != "" {
			db = db.Where("database_name like ? ", "%"+c.Query("database_name")+"%")
		}
		sorterMap := make(map[string]string)
		sorterData := c.Query("sorter")
		json.Unmarshal([]byte(sorterData), &sorterMap)
		for sortField, sortOrder := range sorterMap {
			if sortField != "" && sortOrder != "" {
				db = db.Order(fmt.Sprintf("%s %s", sortField, strings.Replace(sortOrder, "end", "", 1)))
			}
		}

		var dataList []model.MetaDatabase
		result := db.Find(&dataList)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Query Error:" + result.Error.Error()})
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

	if method == "PUT" {
		var record model.MetaDatabase
		if err := c.BindJSON(&record); err != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Bind JSON Error: " + err.Error()})
			return
		}

		if record.Id == 0 {
			c.JSON(200, gin.H{"success": false, "msg": "Invalid ID: ID cannot be zero"})
			return
		}

		updates := map[string]interface{}{
			"alias_name":      record.AliasName,
			"ops_owner":       record.OpsOwner,
			"ops_owner_phone": record.OpsOwnerPhone,
			"is_deleted":      record.IsDeleted,
		}
		result := database.DB.Model(&model.MetaDatabase{}).Where("id = ?", record.Id).Updates(updates)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Update Error: " + result.Error.Error()})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(200, gin.H{"success": false, "msg": "No record found with id"})
			return
		}

		c.JSON(200, gin.H{"success": true, "msg": "Update successful"})
		return
	}
}
