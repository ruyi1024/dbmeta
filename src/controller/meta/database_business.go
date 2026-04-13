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
	"gorm.io/gorm"
)

// DatabaseBusiness 数据库与业务信息关联 CRUD
func DatabaseBusiness(c *gin.Context) {
	switch c.Request.Method {
	case "GET":
		databaseBusinessList(c)
	case "POST":
		databaseBusinessCreate(c)
	case "PUT":
		databaseBusinessUpdate(c)
	default:
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Method not allowed"})
	}
}

func databaseBusinessList(c *gin.Context) {
	var db = database.DB
	if eq := strings.TrimSpace(c.Query("exact_database_name")); eq != "" {
		db = db.Where("database_name = ?", eq)
	} else if q := strings.TrimSpace(c.Query("database_name")); q != "" {
		db = db.Where("database_name LIKE ?", "%"+q+"%")
	}
	if q := strings.TrimSpace(c.Query("app_name")); q != "" {
		db = db.Where("app_name LIKE ?", "%"+q+"%")
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
	var dataList []model.MetaDatabaseBusiness
	result := db.Find(&dataList)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Query Error:" + result.Error.Error()})
		return
	}

	appSet := make(map[string]struct{})
	for _, r := range dataList {
		if r.AppName != "" {
			appSet[r.AppName] = struct{}{}
		}
	}
	appNames := make([]string, 0, len(appSet))
	for k := range appSet {
		appNames = append(appNames, k)
	}
	infoMap := make(map[string]model.MetaBusinessInfo)
	if len(appNames) > 0 {
		var infos []model.MetaBusinessInfo
		database.DB.Where("app_name IN ?", appNames).Find(&infos)
		for _, info := range infos {
			infoMap[info.AppName] = info
		}
	}

	type databaseBusinessListRow struct {
		model.MetaDatabaseBusiness
		AppDescription string `json:"app_description"`
		AppOwner       string `json:"app_owner"`
	}
	out := make([]databaseBusinessListRow, 0, len(dataList))
	for _, r := range dataList {
		info := infoMap[r.AppName]
		out = append(out, databaseBusinessListRow{
			MetaDatabaseBusiness: r,
			AppDescription:       info.AppDescription,
			AppOwner:             info.AppOwner,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
		"data":    out,
		"total":   len(out),
	})
}

func businessInfoExistsByAppName(appName string) bool {
	var n int64
	database.DB.Model(&model.MetaBusinessInfo{}).Where("app_name = ?", appName).Count(&n)
	return n > 0
}

func databaseBusinessCreate(c *gin.Context) {
	var record model.MetaDatabaseBusiness
	if err := c.BindJSON(&record); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Bind JSON Error: " + err.Error()})
		return
	}
	record.DatabaseName = strings.TrimSpace(record.DatabaseName)
	record.AppName = strings.TrimSpace(record.AppName)
	if record.DatabaseName == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据库名不能为空"})
		return
	}
	if record.AppName == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称不能为空"})
		return
	}
	if !businessInfoExistsByAppName(record.AppName) {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称不存在，请先在「业务信息」中维护"})
		return
	}
	result := database.DB.Create(&record)
	if result.Error != nil {
		msg := result.Error.Error()
		if strings.Contains(msg, "Duplicate") {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "该数据库名与应用名的组合已存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Create Error: " + msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "OK", "data": record})
}

func databaseBusinessUpdate(c *gin.Context) {
	var record model.MetaDatabaseBusiness
	if err := c.BindJSON(&record); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Bind JSON Error: " + err.Error()})
		return
	}
	if record.Id == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Invalid ID"})
		return
	}
	record.DatabaseName = strings.TrimSpace(record.DatabaseName)
	record.AppName = strings.TrimSpace(record.AppName)
	if record.DatabaseName == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据库名不能为空"})
		return
	}
	if record.AppName == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称不能为空"})
		return
	}
	if !businessInfoExistsByAppName(record.AppName) {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称不存在，请先在「业务信息」中维护"})
		return
	}
	updates := map[string]interface{}{
		"database_name": record.DatabaseName,
		"app_name":      record.AppName,
		"remark":        record.Remark,
	}
	result := database.DB.Model(&model.MetaDatabaseBusiness{}).Where("id = ?", record.Id).Updates(updates)
	if result.Error != nil {
		msg := result.Error.Error()
		if strings.Contains(msg, "Duplicate") {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "该数据库名与应用名的组合已存在"})
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

// DatabaseBusinessBatchSync POST /meta/database-business/batch-sync
// 按数据库名批量同步与业务信息的关联：提交的应用名为最终集合（可增删关联行）
func DatabaseBusinessBatchSync(c *gin.Context) {
	type batchReq struct {
		DatabaseName string   `json:"database_name"`
		AppNames     []string `json:"app_names"`
	}
	var req batchReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Bind JSON Error: " + err.Error()})
		return
	}
	dbn := strings.TrimSpace(req.DatabaseName)
	if dbn == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "数据库名不能为空"})
		return
	}
	seen := make(map[string]bool)
	var names []string
	for _, a := range req.AppNames {
		t := strings.TrimSpace(a)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		names = append(names, t)
	}
	for _, app := range names {
		if !businessInfoExistsByAppName(app) {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "应用名称不存在，请先在「业务信息」中维护: " + app})
			return
		}
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if len(names) == 0 {
			if err := tx.Where("database_name = ?", dbn).Delete(&model.MetaDatabaseBusiness{}).Error; err != nil {
				return err
			}
			return nil
		}
		if err := tx.Where("database_name = ? AND app_name NOT IN ?", dbn, names).Delete(&model.MetaDatabaseBusiness{}).Error; err != nil {
			return err
		}
		for _, app := range names {
			var cnt int64
			tx.Model(&model.MetaDatabaseBusiness{}).Where("database_name = ? AND app_name = ?", dbn, app).Count(&cnt)
			if cnt > 0 {
				continue
			}
			rec := model.MetaDatabaseBusiness{DatabaseName: dbn, AppName: app}
			if err := tx.Create(&rec).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "Duplicate") {
			c.JSON(http.StatusOK, gin.H{"success": false, "msg": "关联写入冲突，请重试"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "同步失败: " + msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "OK"})
}

// DatabaseBusinessDelete DELETE /meta/database-business/:id
func DatabaseBusinessDelete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "缺少 id"})
		return
	}
	result := database.DB.Where("id = ?", id).Delete(&model.MetaDatabaseBusiness{})
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
