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
	"context"
	"dbmeta-core/log"
	"dbmeta-core/setting"
	"dbmeta-core/src/database"
	"dbmeta-core/src/libary/clickhouse"
	"dbmeta-core/src/libary/mongodb"
	"dbmeta-core/src/libary/mssql"
	"dbmeta-core/src/libary/mysql"
	"dbmeta-core/src/libary/oracle"
	"dbmeta-core/src/libary/postgres"
	"dbmeta-core/src/libary/redis"
	"dbmeta-core/src/model"
	"dbmeta-core/src/utils"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func init() {
	go checker()
}

func checker() {
	time.Sleep(time.Second * time.Duration(30))
	var db = database.DB
	var record model.TaskOption
	db.Select("crontab").Where("task_key=?", "check_datasource").Take(&record)
	c := cron.New()
	c.AddFunc(record.Crontab, func() {
		db.Select("enable").Where("task_key=?", "check_datasource").Take(&record)
		if record.Enable == 1 {
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='check_datasource'").Updates(map[string]interface{}{"heartbeat_time": time.Now().Format("2006-01-02 15:04:05.999")})
			doDatasourceCheck()
			db.Model(model.TaskHeartbeat{}).Where("heartbeat_key='check_datasource'").Updates(map[string]interface{}{"heartbeat_end_time": time.Now().Format("2006-01-02 15:04:05.999")})
		}
	})
	c.Start()
}

func doDatasourceCheck() {
	logger := log.Logger
	logger.Info("开始执行数据源连接检查任务")

	// 创建任务日志记录器
	taskLogger := NewTaskLogger("check_datasource")
	if err := taskLogger.Start(); err != nil {
		logger.Error("创建任务日志失败", zap.Error(err))
		return
	}

	var db = database.DB
	var dataList []model.Datasource
	result := db.Where("enable=1").Order("type asc").Find(&dataList)
	if result.Error != nil {
		errorMsg := fmt.Sprintf("查询数据源失败: %v", result.Error)
		logger.Error(errorMsg)
		taskLogger.Failed(errorMsg)
		return
	}

	if len(dataList) == 0 {
		successMsg := "没有找到启用的数据源"
		logger.Info(successMsg)
		taskLogger.Success(successMsg)
		return
	}

	logger.Info("找到数据源", zap.Int("count", len(dataList)))
	taskLogger.UpdateResult(fmt.Sprintf("找到 %d 个启用的数据源", len(dataList)))

	successCount := 0
	failedCount := 0
	errorDetails := []string{}

	for i, datasource := range dataList {
		datasourceType := datasource.Type
		host := datasource.Host
		port := datasource.Port
		user := datasource.User
		pass := datasource.Pass
		dbid := datasource.Dbid

		logger.Info("处理数据源", zap.Int("index", i+1), zap.Int("total", len(dataList)),
			zap.String("type", datasourceType), zap.String("host", host), zap.String("port", port))

		var origPass string
		if pass != "" {
			var err error
			origPass, err = utils.AesPassDecode(pass, setting.Setting.DbPassKey)
			if err != nil {
				errorMsg := fmt.Sprintf("数据源 %s:%s 密码解密失败: %v", host, port, err)
				logger.Error(errorMsg)
				errorDetails = append(errorDetails, errorMsg)
				failedCount++
				continue
			}
		}

		checkResult := doDatasourceCheckTask(datasourceType, host, port, user, origPass, dbid, datasource.Env)
		if checkResult.Status == 0 {
			errorMsg := fmt.Sprintf("数据源 %s:%s 连接检查失败: %s", host, port, checkResult.StatusText)
			logger.Error(errorMsg)
			errorDetails = append(errorDetails, errorMsg)
			failedCount++
		} else {
			successCount++
		}

		// 更新进度
		progressMsg := fmt.Sprintf("已处理 %d/%d 个数据源 (成功: %d, 失败: %d)", i+1, len(dataList), successCount, failedCount)
		taskLogger.UpdateResult(progressMsg)
	}

	// 记录最终结果
	finalResult := fmt.Sprintf("任务完成 - 数据源总计: %d, 成功: %d, 失败: %d",
		len(dataList), successCount, failedCount)
	if len(errorDetails) > 0 {
		finalResult += fmt.Sprintf("。失败详情: %s", errorDetails[0])
		if len(errorDetails) > 1 {
			finalResult += fmt.Sprintf(" 等 %d 个错误", len(errorDetails))
		}
	}

	if failedCount == 0 {
		taskLogger.Success(finalResult)
	} else {
		taskLogger.Failed(finalResult)
	}

	logger.Info(finalResult)
}

type CheckResult struct {
	Status     int
	StatusText string
}

func doDatasourceCheckTask(datasourceType, host, port, user, pass, dbid, env string) CheckResult {
	status := 1
	statusText := "数据源连接正常"

	if datasourceType == "MySQL" || datasourceType == "TiDB" || datasourceType == "Doris" || datasourceType == "MariaDB" || datasourceType == "GreatSQL" || datasourceType == "OceanBase" {
		db, err := mysql.Connect(host, port, user, pass, dbid)
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Close()
		}
	} else if datasourceType == "Redis" {
		db, err := redis.Connect(host, port, pass)
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Close()
		}
	} else if datasourceType == "PostgreSQL" {
		db, err := postgres.Connect(host, port, user, pass, dbid)
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Close()
		}
	} else if datasourceType == "Oracle" {
		db, err := oracle.Connect(host, port, user, pass, dbid)
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Close()
		}
	} else if datasourceType == "SQLServer" {
		db, err := mssql.Connect(host, port, user, pass, dbid)
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Close()
		}
	} else if datasourceType == "MongoDB" {
		db, err := mongodb.Connect(host, port, user, pass, dbid)
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Disconnect(context.Background())
		}
	} else if datasourceType == "ClickHouse" {
		db, err := clickhouse.Connect(host, port, user, pass, dbid)
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Close()
		}
	}

	// 创建事件
	var events []model.Event
	event := model.Event{
		EventEntity: datasourceType,
		EventKey:    "datasourceCheck",
		EventValue:  float32(status),
		EventDetail: statusText,
		EventTime:   time.Now(),
	}
	events = append(events, event)

	// write events to mysql
	result := database.DB.Model(&model.Event{}).Create(events)
	if result.Error != nil {
		fmt.Println("Insert Event To MySQL Error: " + result.Error.Error())
		log.Logger.Error(fmt.Sprintf("Can't add events data to mysql: %s", result.Error.Error()))
		return CheckResult{Status: 0, StatusText: "写入事件到MySQL失败"}
	}

	return CheckResult{Status: status, StatusText: statusText}
}

// ExecuteDatasourceCheck 导出函数，用于手动执行任务
func ExecuteDatasourceCheck() {
	doDatasourceCheck()
}
