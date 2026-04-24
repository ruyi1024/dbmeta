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

package meta

import (
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BusinessInfo 业务信息 CRUD：GET/POST/PUT /meta/business-info/list
func BusinessInfo(c *gin.Context) {
	method := c.Request.Method
	switch method {
	case "GET":
		businessInfoList(c)
	case "POST":
		businessInfoCreate(c)
	case "PUT":
		businessInfoUpdate(c)
	default:
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Method not allowed"})
	}
}

func businessInfoList(c *gin.Context) {
	var db = database.DB
	if q := strings.TrimSpace(c.Query("app_name")); q != "" {
		db = db.Where("app_name LIKE ?", "%"+q+"%")
	}
	if q := strings.TrimSpace(c.Query("app_owner")); q != "" {
		db = db.Where("app_owner LIKE ?", "%"+q+"%")
	}
	sorterMap := make(map[string]string)
	_ = json.Unmarshal([]byte(c.Query("sorter")), &sorterMap)
	if len(sorterMap) == 0 {
		db = db.Order("gmt_created DESC")
	} else {
		for sortField, sortOrder := range sorterMap {
			if sortField != "" && sortOrder != "" {
				db = db.Order(fmt.Sprintf("%s %s", sortField, strings.Replace(sortOrder, "end", "", 1)))
			}
		}
	}
	var dataList []model.MetaBusinessInfo
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
}

func businessInfoCreate(c *gin.Context) {
	var record model.MetaBusinessInfo
	if err := c.BindJSON(&record); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Bind JSON Error: " + err.Error()})
		return
	}
	record.AppName = strings.TrimSpace(record.AppName)
	if record.AppName == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称不能为空"})
		return
	}
	result := database.DB.Create(&record)
	if result.Error != nil {
		msg := result.Error.Error()
		if strings.Contains(msg, "Duplicate") {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称已存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Create Error: " + msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "OK", "data": record})
}

func businessInfoUpdate(c *gin.Context) {
	var record model.MetaBusinessInfo
	if err := c.BindJSON(&record); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Bind JSON Error: " + err.Error()})
		return
	}
	if record.Id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Invalid ID"})
		return
	}
	record.AppName = strings.TrimSpace(record.AppName)
	if record.AppName == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称不能为空"})
		return
	}
	updates := map[string]interface{}{
		"app_name":        record.AppName,
		"app_description": record.AppDescription,
		"app_owner":       record.AppOwner,
		"app_owner_email": record.AppOwnerEmail,
		"app_owner_phone": record.AppOwnerPhone,
		"remark":          record.Remark,
	}
	result := database.DB.Model(&model.MetaBusinessInfo{}).Where("id = ?", record.Id).Updates(updates)
	if result.Error != nil {
		msg := result.Error.Error()
		if strings.Contains(msg, "Duplicate") {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称已存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Update Error: " + msg})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "未找到记录或数据未变化"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "Update successful"})
}

// BusinessInfoDelete DELETE /meta/business-info/:id
func BusinessInfoDelete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "缺少 id"})
		return
	}
	result := database.DB.Where("id = ?", id).Delete(&model.MetaBusinessInfo{})
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Delete Error: " + result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "未找到记录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "OK"})
}
