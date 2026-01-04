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
	"dbmcloud/setting"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProcessListRequest 请求参数
type ProcessListRequest struct {
	DatasourceId int `json:"datasource_id" binding:"required"`
}

// ProcessListResponse 响应数据
type ProcessListResponse struct {
	Id      int    `json:"id"`
	User    string `json:"user"`
	Host    string `json:"host"`
	Db      string `json:"db"`
	Command string `json:"command"`
	Time    int    `json:"time"`
	State   string `json:"state"`
	Info    string `json:"info"`
}

// GetProcessList 获取进程列表
func GetProcessList(c *gin.Context) {
	var req ProcessListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "参数错误: " + err.Error()})
		return
	}

	// 获取数据源信息
	var datasource model.Datasource
	result := database.DB.Where("id = ?", req.DatasourceId).First(&datasource)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "数据源不存在"})
		return
	}

	// 解密密码
	origPass, err := utils.AesPassDecode(datasource.Pass, setting.Setting.DbPassKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "密码解密失败"})
		return
	}

	// 连接数据库
	dbCon, err := database.Connect(
		database.WithDriver("mysql"),
		database.WithHost(datasource.Host),
		database.WithPort(datasource.Port),
		database.WithUsername(datasource.User),
		database.WithPassword(origPass),
		database.WithDatabase("information_schema"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "连接数据库失败: " + err.Error()})
		return
	}
	defer dbCon.Close()

	// 查询processlist
	querySql := "SELECT id, user, host, db, command, time, state, info FROM information_schema.processlist ORDER BY time DESC"
	processList, err := database.QueryRemote(dbCon, querySql)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "查询失败: " + err.Error()})
		return
	}

	// 转换数据格式
	var response []ProcessListResponse
	for _, item := range processList {
		response = append(response, ProcessListResponse{
			Id:      utils.StrToInt(formatProcessInterface(item["id"])),
			User:    formatProcessInterface(item["user"]),
			Host:    formatProcessInterface(item["host"]),
			Db:      formatProcessInterface(item["db"]),
			Command: formatProcessInterface(item["command"]),
			Time:    utils.StrToInt(formatProcessInterface(item["time"])),
			State:   formatProcessInterface(item["state"]),
			Info:    formatProcessInterface(item["info"]),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"count":   len(response),
	})
}

func formatProcessInterface(inter interface{}) string {
	if inter != nil {
		return inter.(string)
	}
	return ""
}
