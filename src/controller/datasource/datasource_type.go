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

package datasource

import (
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ruyi1024/dbmeta/webassets"
)

const defaultDatasourceTypeLogo = "/static/dblogo/logo-mysql.png"

func datasourceTypeDefaultLogo(name string) string {
	switch strings.TrimSpace(name) {
	case "MySQL":
		return "/static/dblogo/logo-mysql.png"
	case "MariaDB":
		return "/static/dblogo/logo-mariadb.png"
	case "GreatSQL":
		return "/static/dblogo/logo-greatsql.png"
	case "TiDB":
		return "/static/dblogo/logo-tidb.png"
	case "Doris":
		return "/static/dblogo/logo-doris.png"
	case "OceanBase":
		return "/static/dblogo/logo-oceanbase.png"
	case "ClickHouse":
		return "/static/dblogo/logo-clickhouse.png"
	case "Oracle":
		return "/static/dblogo/logo-oracle.png"
	case "PostgreSQL":
		return "/static/dblogo/logo-postgresql.png"
	case "SQLServer":
		return "/static/dblogo/logo-sqlserver.png"
	case "MongoDB":
		return "/static/dblogo/logo-mongodb.png"
	case "Redis":
		return "/static/dblogo/logo-redis.png"
	case "达梦数据库":
		return "/static/dblogo/logo-dm.png"
	default:
		return defaultDatasourceTypeLogo
	}
}

func TypeLogo(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("path"))
	if raw == "" {
		raw = defaultDatasourceTypeLogo
	}
	if strings.Contains(raw, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid logo path"})
		return
	}
	clean := strings.TrimPrefix(raw, "/")
	if !strings.HasPrefix(clean, "static/") {
		clean = "static/" + strings.TrimPrefix(clean, "static/")
	}
	c.FileFromFS(clean, http.FS(webassets.Static))
}

func TypeList(c *gin.Context) {
	var db = database.DB
	method := c.Request.Method
	if method == "GET" {
		var dataList []model.DatasourceType
		if c.Query("enable") != "" {
			db = db.Where("enable=?", c.Query("enable"))
		}
		db.Order("sort asc")
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
		var record model.DatasourceType
		c.BindJSON(&record)
		if strings.TrimSpace(record.Logo) == "" {
			record.Logo = datasourceTypeDefaultLogo(record.Name)
		}
		result := database.DB.Create(&record)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Insert Error: " + result.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
		return

	}

	if method == "PUT" {
		var record model.DatasourceType
		c.BindJSON(&record)
		var oldRecord model.DatasourceType
		if err := database.DB.Where("id = ?", record.Id).First(&oldRecord).Error; err == nil {
			if strings.TrimSpace(record.Logo) == "" {
				if strings.TrimSpace(oldRecord.Logo) != "" {
					record.Logo = oldRecord.Logo
				} else {
					record.Logo = datasourceTypeDefaultLogo(record.Name)
				}
			}
		} else if strings.TrimSpace(record.Logo) == "" {
			record.Logo = datasourceTypeDefaultLogo(record.Name)
		}
		result := database.DB.Model(&record).Omit("id").Where("id = ?", record.Id).Updates(record)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Update Error: " + result.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
		return
	}

	if method == "DELETE" {
		var record model.DatasourceType
		c.BindJSON(&record)
		result := database.DB.Model(&model.DatasourceType{}).Where("id = ?", record.Id).Delete(record)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Delete Error:" + result.Error.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true})
		return
	}
}
