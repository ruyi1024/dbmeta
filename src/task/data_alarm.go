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

package task

import (
	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/src/controller/data"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	go dataAlarmCrontabTask()
}

// dataAlarmCrontabTask 启动数据告警定时任务
func dataAlarmCrontabTask() {
	time.Sleep(time.Second * time.Duration(30))

	// 创建定时器
	c := cron.New()

	// 每分钟检查一次是否有需要执行的任务
	c.AddFunc("* * * * *", func() {
		executeDataAlarms()
	})

	c.Start()
}

// executeDataAlarms 执行所有需要执行的数据告警
func executeDataAlarms() {
	logger := log.Logger

	// 获取所有启用的告警
	var alarms []model.DataAlarm
	result := database.DB.Where("status = 1").Find(&alarms)
	if result.Error != nil {
		logger.Error("查询数据告警失败", zap.Error(result.Error))
		return
	}

	if len(alarms) == 0 {
		return
	}

	// 获取当前时间
	now := time.Now()

	// 遍历所有告警，检查是否需要执行
	for _, alarm := range alarms {
		// 检查 cron 表达式是否匹配当前时间
		if shouldExecuteAlarm(&alarm, now) {
			logger.Info("执行数据告警", zap.Int("alarm_id", alarm.Id), zap.String("alarm_name", alarm.AlarmName))
			// 异步执行告警任务
			go data.ExecuteAlarmTask(&alarm)
		}
	}
}

// shouldExecuteAlarm 检查告警是否应该执行
func shouldExecuteAlarm(alarm *model.DataAlarm, now time.Time) bool {
	// 解析 cron 表达式
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(alarm.CronExpression)
	if err != nil {
		log.Logger.Error("解析cron表达式失败", zap.Error(err), zap.String("cron", alarm.CronExpression))
		return false
	}

	// 检查上次执行时间
	if alarm.LastRunTime != nil {
		// 如果上次执行时间距离现在不到1分钟，跳过（避免重复执行）
		if now.Sub(*alarm.LastRunTime) < time.Minute {
			return false
		}
	}

	// 检查 cron 表达式是否匹配当前时间
	nextRun := schedule.Next(now.Add(-time.Minute))
	return nextRun.Before(now) || nextRun.Equal(now)
}
