package task

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"fmt"
	"time"
)

// TaskLogger 任务日志记录器
type TaskLogger struct {
	TaskKey   string
	LogID     int64
	StartTime time.Time
}

// NewTaskLogger 创建新的任务日志记录器
func NewTaskLogger(taskKey string) *TaskLogger {
	return &TaskLogger{
		TaskKey:   taskKey,
		StartTime: time.Now(),
	}
}

// Start 开始记录任务
func (tl *TaskLogger) Start() error {
	taskLog := model.TaskLog{
		TaskKey:   tl.TaskKey,
		StartTime: tl.StartTime,
		Status:    "running",
		Result:    "任务开始执行",
	}

	result := database.DB.Create(&taskLog)
	if result.Error != nil {
		return fmt.Errorf("创建任务日志失败: %v", result.Error)
	}

	tl.LogID = taskLog.Id
	return nil
}

// Complete 完成任务记录
func (tl *TaskLogger) Complete(status string, result string) error {
	if tl.LogID == 0 {
		return fmt.Errorf("任务日志ID未初始化")
	}

	completeTime := time.Now()
	duration := completeTime.Sub(tl.StartTime)

	updateData := map[string]interface{}{
		"complete_time": &completeTime,
		"status":        status,
		"result":        fmt.Sprintf("%s (执行时长: %v)", result, duration),
	}

	dbResult := database.DB.Model(&model.TaskLog{}).Where("id = ?", tl.LogID).Updates(updateData)
	if dbResult.Error != nil {
		return fmt.Errorf("更新任务日志失败: %v", dbResult.Error)
	}

	return nil
}

// Success 记录任务成功
func (tl *TaskLogger) Success(result string) error {
	return tl.Complete("success", result)
}

// Failed 记录任务失败
func (tl *TaskLogger) Failed(result string) error {
	return tl.Complete("failed", result)
}

// UpdateResult 更新任务结果（不改变状态）
func (tl *TaskLogger) UpdateResult(result string) error {
	if tl.LogID == 0 {
		return fmt.Errorf("任务日志ID未初始化")
	}

	dbResult := database.DB.Model(&model.TaskLog{}).Where("id = ?", tl.LogID).Update("result", result)
	if dbResult.Error != nil {
		return fmt.Errorf("更新任务结果失败: %v", dbResult.Error)
	}

	return nil
}

// GetTaskLogs 获取任务执行日志
func GetTaskLogs(taskKey string, limit int) ([]model.TaskLog, error) {
	var logs []model.TaskLog

	query := database.DB.Where("task_key = ?", taskKey).Order("gmt_created DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&logs)
	if result.Error != nil {
		return nil, fmt.Errorf("查询任务日志失败: %v", result.Error)
	}

	return logs, nil
}

// GetTaskLogsByDateRange 按日期范围获取任务日志
func GetTaskLogsByDateRange(taskKey string, startDate, endDate time.Time) ([]model.TaskLog, error) {
	var logs []model.TaskLog

	result := database.DB.Where("task_key = ? AND gmt_created BETWEEN ? AND ?",
		taskKey, startDate, endDate).Order("gmt_created DESC").Find(&logs)

	if result.Error != nil {
		return nil, fmt.Errorf("查询任务日志失败: %v", result.Error)
	}

	return logs, nil
}
