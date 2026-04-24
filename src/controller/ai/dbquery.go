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

package ai

import (
	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DbQueryRequest 智能查数请求
type DbQueryRequest struct {
	Question       string `json:"question" binding:"required"`
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
	ModelId        int    `json:"model_id"`        // 可选，指定使用的AI模型ID
	DatasourceId   int    `json:"datasource_id"`   // 可选，指定数据源ID
	DatabaseName   string `json:"database_name"`   // 可选，指定数据库名
	DatasourceType string `json:"datasource_type"` // 可选，指定数据源类型
	Host           string `json:"host"`            // 可选，指定主机
	Port           string `json:"port"`            // 可选，指定端口
	TableName      string `json:"table_name"`      // 可选，指定表名
}

// DbQueryResponse 智能查数响应
type DbQueryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    *struct {
		SQLQuery    string                   `json:"sql_query"`
		QueryResult []map[string]interface{} `json:"query_result"`
		Total       int                      `json:"total"`
		Page        int                      `json:"page"`
		PageSize    int                      `json:"page_size"`
	} `json:"data,omitempty"`
}

// DbQuery 智能查数接口
func DbQuery(c *gin.Context) {
	var req DbQueryRequest
	if err := c.BindJSON(&req); err != nil {
		log.Logger.Error("解析请求参数失败", zap.Error(err))
		c.JSON(http.StatusOK, DbQueryResponse{
			Success: false,
			Message: "参数解析失败: " + err.Error(),
		})
		return
	}

	// 验证必填参数
	if req.Question == "" {
		c.JSON(http.StatusOK, DbQueryResponse{
			Success: false,
			Message: "查询问题不能为空",
		})
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	// 限制最大页大小
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	log.Logger.Info("收到智能查数请求",
		zap.String("question", req.Question),
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize),
		zap.Int("model_id", req.ModelId),
		zap.String("database_name", req.DatabaseName),
		zap.String("datasource_type", req.DatasourceType),
		zap.String("host", req.Host),
		zap.String("port", req.Port))

	// 调用服务处理查询
	serviceReq := &service.DbQueryRequest{
		Question:       req.Question,
		Page:           req.Page,
		PageSize:       req.PageSize,
		ModelId:        req.ModelId,
		DatasourceId:   req.DatasourceId,
		DatabaseName:   req.DatabaseName,
		DatasourceType: req.DatasourceType,
		Host:           req.Host,
		Port:           req.Port,
		TableName:      req.TableName,
	}

	result, err := service.ProcessDbQuery(serviceReq)
	if err != nil {
		log.Logger.Error("处理智能查数失败", zap.Error(err))
		c.JSON(http.StatusOK, DbQueryResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 返回成功结果
	c.JSON(http.StatusOK, DbQueryResponse{
		Success: true,
		Data: &struct {
			SQLQuery    string                   `json:"sql_query"`
			QueryResult []map[string]interface{} `json:"query_result"`
			Total       int                      `json:"total"`
			Page        int                      `json:"page"`
			PageSize    int                      `json:"page_size"`
		}{
			SQLQuery:    result.SQLQuery,
			QueryResult: result.QueryResult,
			Total:       result.Total,
			Page:        result.Page,
			PageSize:    result.PageSize,
		},
	})
}
