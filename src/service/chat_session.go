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
*/

package service

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"fmt"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateSession 创建新会话
func CreateSession(userName string) (*model.ChatSession, error) {
	sessionId := uuid.New().String()
	title := fmt.Sprintf("新对话 %s", time.Now().Format("01-02 15:04"))

	session := &model.ChatSession{
		SessionId: sessionId,
		UserName:  userName,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := database.DB.Create(session)
	if result.Error != nil {
		return nil, fmt.Errorf("创建会话失败: %v", result.Error)
	}

	return session, nil
}

// GetSession 获取会话
func GetSession(sessionId string) (*model.ChatSession, error) {
	var session model.ChatSession
	result := database.DB.Where("session_id = ?", sessionId).First(&session)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("会话不存在")
		}
		return nil, fmt.Errorf("查询会话失败: %v", result.Error)
	}

	return &session, nil
}

// ListSessions 列出用户的所有会话
func ListSessions(userName string) ([]model.ChatSession, error) {
	var sessions []model.ChatSession
	result := database.DB.Where("user_name = ?", userName).
		Order("updated_at DESC").
		Find(&sessions)
	if result.Error != nil {
		return nil, fmt.Errorf("查询会话列表失败: %v", result.Error)
	}

	return sessions, nil
}

// DeleteSession 删除会话
func DeleteSession(sessionId, userName string) error {
	// 先删除会话的所有消息
	result := database.DB.Where("session_id = ?", sessionId).Delete(&model.ChatMessage{})
	if result.Error != nil {
		zap.L().Error("删除会话消息失败", zap.Error(result.Error))
	}

	// 删除会话
	result = database.DB.Where("session_id = ? AND user_name = ?", sessionId, userName).
		Delete(&model.ChatSession{})
	if result.Error != nil {
		return fmt.Errorf("删除会话失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("会话不存在或无权限删除")
	}

	return nil
}

// UpdateSessionTitle 更新会话标题
func UpdateSessionTitle(sessionId, title string) error {
	result := database.DB.Model(&model.ChatSession{}).
		Where("session_id = ?", sessionId).
		Updates(map[string]interface{}{
			"title":      title,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("更新会话标题失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("会话不存在")
	}

	return nil
}

// UpdateSessionTime 更新会话时间（用于每次对话后更新）
func UpdateSessionTime(sessionId string) error {
	result := database.DB.Model(&model.ChatSession{}).
		Where("session_id = ?", sessionId).
		Update("updated_at", time.Now())
	if result.Error != nil {
		return fmt.Errorf("更新会话时间失败: %v", result.Error)
	}

	return nil
}

// SaveMessage 保存消息
func SaveMessage(sessionId, role, content, sqlQuery string, queryResult []map[string]interface{}) error {
	// 将[]map[string]interface{}转换为model.QueryResultArray
	// model.QueryResultArray是[]map[string]interface{}的类型别名，可以直接赋值
	var jsonResult model.QueryResultArray
	if queryResult != nil {
		// 通过unsafe转换或循环复制
		jsonResult = *(*model.QueryResultArray)(unsafe.Pointer(&queryResult))
	}

	message := &model.ChatMessage{
		SessionId:   sessionId,
		Role:        role,
		Content:     content,
		SqlQuery:    sqlQuery,
		QueryResult: jsonResult,
		CreatedAt:   time.Now(),
	}

	result := database.DB.Create(message)
	if result.Error != nil {
		return fmt.Errorf("保存消息失败: %v", result.Error)
	}

	// 更新会话时间
	UpdateSessionTime(sessionId)

	return nil
}

// GetSessionMessages 获取会话消息历史
func GetSessionMessages(sessionId string, limit int) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	query := database.DB.Where("session_id = ?", sessionId).
		Order("gmt_created ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&messages)
	if result.Error != nil {
		return nil, fmt.Errorf("查询消息历史失败: %v", result.Error)
	}

	return messages, nil
}
