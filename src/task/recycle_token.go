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

package task

import (
	"dbmcloud/log"
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	go recycleToken()
}

func recycleToken() {
	/*
		time.Sleep(time.Second * time.Duration(rand.Intn(60)))
		timer := time.NewTimer(120 * time.Second)
		defer timer.Stop()
		for {
			<-timer.C
			database.DB.Model(model.TaskHeartbeat{}).Where("heartbeat_key='recycle_token'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			database.DB.Delete(model.Token{}, "expired <= ?", time.Now().Format("2006-01-02 15:04:05"))
			database.DB.Model(model.TaskHeartbeat{}).Where("heartbeat_key='recycle_token'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})

		}
	*/

	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	result := db.Select("crontab").Where("task_key=?", "recycle_token").Take(&record)
	if result.Error != nil {
		log.Logger.Error(result.Error.Error())
		return

	}
	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "recycle_token").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='recycle_token'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doRecycleTokenTask()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='recycle_token'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

func doRecycleTokenTask() {
	logger := log.Logger
	logger.Info("开始执行Token清理任务")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("recycle_token")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	// 查询过期的Token数量
	var expiredCount int64
	expireTime := time.Now().Format("2006-01-02 15:04:05")
	
	countResult := database.DB.Model(&model.Token{}).Where("expired <= ?", expireTime).Count(&expiredCount)
	if countResult.Error != nil {
		errorMsg := fmt.Sprintf("查询过期Token数量失败: %v", countResult.Error)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	logger.Info("找到过期Token", zap.Int64("count", expiredCount))
	taskLogger.UpdateResult(fmt.Sprintf("找到 %d 个过期Token", expiredCount))

	if expiredCount == 0 {
		successMsg := "没有过期的Token需要清理"
		logger.Info(successMsg)
		taskLogger.Success(successMsg)
		return
	}

	// 删除过期Token
	deleteResult := database.DB.Delete(model.Token{}, "expired <= ?", expireTime)
	if deleteResult.Error != nil {
		errorMsg := fmt.Sprintf("删除过期Token失败: %v", deleteResult.Error)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	// 记录最终结果
	finalResult := fmt.Sprintf("Token清理完成 - 预期删除: %d, 实际删除: %d", expiredCount, deleteResult.RowsAffected)
	
	if deleteResult.RowsAffected != expiredCount {
		finalResult += fmt.Sprintf(" (数量不匹配，可能有并发操作)")
	}

	taskLogger.Success(finalResult)
	logger.Info(finalResult)
}
