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

package grading

import (
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"github.com/ruyi1024/dbmeta/src/utils"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getLoginUsername(c *gin.Context) string {
	userinfo, _ := c.Get("loginUser")
	data, _ := json.Marshal(&userinfo)
	userMap := make(map[string]interface{})
	_ = json.Unmarshal(data, &userMap)
	if u, ok := userMap["username"].(string); ok {
		return u
	}
	return ""
}

func okList(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"list":     list,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// ListGrades 分级字典列表
func ListGrades(c *gin.Context) {
	var rows []model.DataSecurityGrade
	database.DB.Order("level_order ASC").Find(&rows)
	out := make([]map[string]interface{}, 0, len(rows))
	for _, g := range rows {
		out = append(out, map[string]interface{}{
			"id":          g.Id,
			"gradeCode":   g.GradeCode,
			"gradeName":   g.GradeName,
			"levelOrder":  g.LevelOrder,
			"description": g.Description,
			"standardRef": g.StandardRef,
			"enable":      g.Enable,
			"gmtCreated":  g.GmtCreated,
			"gmtUpdated":  g.GmtUpdated,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": out})
}

type gradeUpdateReq struct {
	Id          int64  `json:"id"`
	Enable      *int8  `json:"enable"`
	Description string `json:"description"`
	StandardRef string `json:"standardRef"`
}

// UpdateGrade 更新分级字典（启用状态、说明等，不修改编码）
func UpdateGrade(c *gin.Context) {
	var req gradeUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}
	updates := map[string]interface{}{
		"description":  req.Description,
		"standard_ref": req.StandardRef,
		"gmt_updated":  time.Now(),
	}
	if req.Enable != nil {
		updates["enable"] = *req.Enable
	}
	if err := database.DB.Model(&model.DataSecurityGrade{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

type assetReq struct {
	Id           int64  `json:"id"`
	DatasourceId int    `json:"datasourceId"`
	DatabaseName string `json:"databaseName"`
	TableName    string `json:"tableName"`
	ColumnName   string `json:"columnName"`
	GradeId      int64  `json:"gradeId"`
	AssignSource string `json:"assignSource"`
	Remark       string `json:"remark"`
}

// ListAssets 资产分级分页
func ListAssets(c *gin.Context) {
	current := c.Query("current")
	page := c.Query("page")
	if current != "" {
		page = current
	}
	if page == "" {
		page = "1"
	}
	pageNum := utils.StrToInt(page)
	if pageNum < 1 {
		pageNum = 1
	}
	pageSizeStr := c.Query("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	pageSize := utils.StrToInt(pageSizeStr)
	if pageSize < 1 {
		pageSize = 10
	}

	q := database.DB.Model(&model.DataAssetSecurityGrade{})
	if c.Query("datasourceId") != "" {
		q = q.Where("datasource_id = ?", utils.StrToInt(c.Query("datasourceId")))
	}
	if c.Query("databaseName") != "" {
		q = q.Where("database_name LIKE ?", "%"+c.Query("databaseName")+"%")
	}
	if c.Query("tableName") != "" {
		q = q.Where("table_name LIKE ?", "%"+c.Query("tableName")+"%")
	}
	if c.Query("gradeId") != "" {
		q = q.Where("grade_id = ?", utils.StrToInt64(c.Query("gradeId")))
	}

	var total int64
	q.Count(&total)

	var rows []model.DataAssetSecurityGrade
	offset := (pageNum - 1) * pageSize
	q.Order("gmt_updated DESC").Offset(offset).Limit(pageSize).Find(&rows)

	var grades []model.DataSecurityGrade
	database.DB.Find(&grades)
	nameByID := map[int64]string{}
	for _, g := range grades {
		nameByID[g.Id] = g.GradeName
	}

	list := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		list = append(list, map[string]interface{}{
			"id":           r.Id,
			"datasourceId": r.DatasourceId,
			"databaseName": r.DatabaseName,
			"tableName":    r.TableNameX,
			"columnName":   r.ColumnName,
			"gradeId":      r.GradeId,
			"gradeName":    nameByID[r.GradeId],
			"assignSource": r.AssignSource,
			"confidence":   r.Confidence,
			"remark":       r.Remark,
			"createdBy":    r.CreatedBy,
			"updatedBy":    r.UpdatedBy,
			"gmtCreated":   r.GmtCreated,
			"gmtUpdated":   r.GmtUpdated,
		})
	}
	okList(c, list, total, pageNum, pageSize)
}

func validateAsset(req *assetReq) error {
	if req.DatasourceId <= 0 || req.TableName == "" || req.GradeId <= 0 {
		return errors.New("数据源、表名、分级不能为空")
	}
	var cnt int64
	database.DB.Model(&model.Datasource{}).Where("id = ?", req.DatasourceId).Count(&cnt)
	if cnt == 0 {
		return errors.New("数据源不存在")
	}
	var gcnt int64
	database.DB.Model(&model.DataSecurityGrade{}).Where("id = ? AND enable = 1", req.GradeId).Count(&gcnt)
	if gcnt == 0 {
		return errors.New("分级无效或已停用")
	}
	return nil
}

// CreateAsset 新增资产分级
func CreateAsset(c *gin.Context) {
	var req assetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	if req.AssignSource == "" {
		req.AssignSource = "manual"
	}
	if err := validateAsset(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	u := getLoginUsername(c)
	now := time.Now()
	row := model.DataAssetSecurityGrade{
		DatasourceId: req.DatasourceId,
		DatabaseName: req.DatabaseName,
		TableNameX:   req.TableName,
		ColumnName:   req.ColumnName,
		GradeId:      req.GradeId,
		AssignSource: req.AssignSource,
		Remark:       req.Remark,
		CreatedBy:    u,
		UpdatedBy:    u,
		GmtCreated:   now,
		GmtUpdated:   now,
	}
	if err := database.DB.Create(&row).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "保存失败，可能已存在相同资产范围"})
		return
	}
	database.DB.Create(&model.DataAssetSecurityGradeLog{
		AssetId:    row.Id,
		GradeIdOld: nil,
		GradeIdNew: row.GradeId,
		Action:     "create",
		Operator:   u,
		GmtCreated: now,
	})
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": gin.H{"id": row.Id}})
}

// UpdateAsset 更新资产分级
func UpdateAsset(c *gin.Context) {
	var req assetReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}
	var row model.DataAssetSecurityGrade
	if err := database.DB.First(&row, req.Id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "记录不存在"})
		return
	}
	prevGrade := row.GradeId
	row.DatasourceId = req.DatasourceId
	row.DatabaseName = req.DatabaseName
	row.TableNameX = req.TableName
	row.ColumnName = req.ColumnName
	row.GradeId = req.GradeId
	if req.AssignSource != "" {
		row.AssignSource = req.AssignSource
	}
	row.Remark = req.Remark
	tmp := assetReq{
		DatasourceId: row.DatasourceId,
		DatabaseName: row.DatabaseName,
		TableName:    row.TableNameX,
		ColumnName:   row.ColumnName,
		GradeId:      row.GradeId,
	}
	if err := validateAsset(&tmp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	u := getLoginUsername(c)
	now := time.Now()
	row.UpdatedBy = u
	row.GmtUpdated = now
	if err := database.DB.Save(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}
	if prevGrade != row.GradeId {
		og := prevGrade
		database.DB.Create(&model.DataAssetSecurityGradeLog{
			AssetId:    row.Id,
			GradeIdOld: &og,
			GradeIdNew: row.GradeId,
			Action:     "update",
			Operator:   u,
			GmtCreated: now,
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

// DeleteAsset 删除资产分级
func DeleteAsset(c *gin.Context) {
	id := utils.StrToInt64(c.Param("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "id 无效"})
		return
	}
	var row model.DataAssetSecurityGrade
	if err := database.DB.First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}
	u := getLoginUsername(c)
	now := time.Now()
	gid := row.GradeId
	if err := database.DB.Delete(&model.DataAssetSecurityGrade{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败"})
		return
	}
	database.DB.Create(&model.DataAssetSecurityGradeLog{
		AssetId:    id,
		GradeIdOld: &gid,
		GradeIdNew: 0,
		Action:     "delete",
		Operator:   u,
		GmtCreated: now,
	})
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
}

// ListLogs 变更记录
func ListLogs(c *gin.Context) {
	current := c.Query("current")
	page := c.Query("page")
	if current != "" {
		page = current
	}
	if page == "" {
		page = "1"
	}
	pageNum := utils.StrToInt(page)
	if pageNum < 1 {
		pageNum = 1
	}
	pageSizeStr := c.Query("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	pageSize := utils.StrToInt(pageSizeStr)

	q := database.DB.Model(&model.DataAssetSecurityGradeLog{})
	if c.Query("assetId") != "" {
		q = q.Where("asset_id = ?", utils.StrToInt64(c.Query("assetId")))
	}
	var total int64
	q.Count(&total)
	var rows []model.DataAssetSecurityGradeLog
	offset := (pageNum - 1) * pageSize
	q.Order("gmt_created DESC").Offset(offset).Limit(pageSize).Find(&rows)

	var grades []model.DataSecurityGrade
	database.DB.Find(&grades)
	nameByID := map[int64]string{}
	for _, g := range grades {
		nameByID[g.Id] = g.GradeName
	}

	list := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		item := map[string]interface{}{
			"id":         r.Id,
			"assetId":    r.AssetId,
			"gradeIdNew": r.GradeIdNew,
			"action":     r.Action,
			"reason":     r.Reason,
			"operator":   r.Operator,
			"gmtCreated": r.GmtCreated,
		}
		if r.GradeIdOld != nil {
			item["gradeIdOld"] = *r.GradeIdOld
			item["gradeNameOld"] = nameByID[*r.GradeIdOld]
		} else {
			item["gradeIdOld"] = nil
			item["gradeNameOld"] = ""
		}
		if r.GradeIdNew > 0 {
			item["gradeNameNew"] = nameByID[r.GradeIdNew]
		} else {
			item["gradeNameNew"] = ""
		}
		list = append(list, item)
	}
	okList(c, list, total, pageNum, pageSize)
}
