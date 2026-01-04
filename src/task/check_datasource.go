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
	"dbmcloud/setting"
	"dbmcloud/src/database"
	"dbmcloud/src/libary/clickhouse"
	"dbmcloud/src/libary/mongodb"
	"dbmcloud/src/libary/mssql"
	"dbmcloud/src/libary/mysql"
	"dbmcloud/src/libary/oracle"
	"dbmcloud/src/libary/postgres"
	"dbmcloud/src/libary/redis"
	"dbmcloud/src/libary/tool"
	"dbmcloud/src/model"
	"dbmcloud/src/mq"
	"dbmcloud/src/utils"
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
		env := datasource.Env

		logger.Info("检查数据源", zap.Int("index", i+1), zap.Int("total", len(dataList)),
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
		checkStatus := checkConnectionTask(datasourceType, env, host, port, user, origPass, dbid)
		if checkStatus.Status == 1 {
			successCount++
		} else {
			failedCount++
			errorDetails = append(errorDetails, checkStatus.StatusText)
		}

		// 更新进度
		progressMsg := fmt.Sprintf("已检查 %d/%d 个数据源 (成功: %d, 失败: %d)", i+1, len(dataList), successCount, failedCount)
		taskLogger.UpdateResult(progressMsg)
	}

	// 记录最终结果
	finalResult := fmt.Sprintf("任务完成 - 数据源总计: %d, 连接成功: %d, 连接失败: %d", len(dataList), successCount, failedCount)
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
	Status     int32
	StatusText string
}

func checkConnectionTask(datasourceType, env, host, port, user, pass, dbid string) CheckResult {

	var status int32 = 1
	var statusText = "数据源服务连接正常."
	if datasourceType == "MySQL" || datasourceType == "TiDB" || datasourceType == "Doris" || datasourceType == "MariaDB" || datasourceType == "GreatSQL" || datasourceType == "OceanBase" {
		db, err := mysql.Connect(host, port, user, pass, "")
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Close()
		}

	} else if datasourceType == "ClickHouse" {
		db, err := clickhouse.Connect(host, port, user, pass, "")
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			defer db.Close()
		}
	} else if datasourceType == "PostgreSQL" {
		db, err := postgres.Connect(host, port, user, pass, "postgres")
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
		db, err := mssql.Connect(host, port, user, pass, "")
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
	} else if datasourceType == "MongoDB" {
		_, err := mongodb.Connect(host, port, user, pass, "local")
		if err != nil {
			status = 0
			statusText = fmt.Sprintf("数据源通信失败: Can't connect server on %s:%s, %s", host, port, err)
			log.Logger.Error(fmt.Sprintf("Datasource check: Can't connect server on %s:%s, %s", host, port, err))
		} else {
			//defer db.Close()
		}
	} else {
		return CheckResult{Status: 0, StatusText: "不支持的数据源类型"}
	}

	var db = database.DB
	var record model.Datasource
	record.Status = status
	record.StatusText = statusText
	db.Model(&record).Select("status", "status_text").Omit("id").Where("host=?", host).Where("port=?", port).Updates(&record)

	//数据源监测事件
	eventEntity := fmt.Sprintf("%s:%s", host, port)
	eventType := datasourceType
	eventGroup := env
	detail := make([]map[string]interface{}, 0)
	detail = append(detail, map[string]interface{}{"Error": statusText})
	events := make([]map[string]interface{}, 0)
	event := map[string]interface{}{
		"event_uuid":   tool.GetUUID(),
		"event_time":   tool.GetNowTime(),
		"event_type":   eventType,
		"event_group":  eventGroup,
		"event_entity": eventEntity,
		"event_key":    "datasourceCheck",
		"event_value":  utils.IntToDecimal(int(status)),
		"event_tag":    "",
		"event_unit":   "",
		"event_detail": utils.MapToStr(detail),
	}
	events = append(events, event)

	// write events to mysql
	result := database.DB.Model(&model.Event{}).Create(events)
	if result.Error != nil {
		fmt.Println("Insert Event To MySQL Error: " + result.Error.Error())
		log.Logger.Error(fmt.Sprintf("Can't add events data to mysql: %s", result.Error.Error()))
		return CheckResult{Status: 0, StatusText: "写入事件到MySQL失败"}
	}

	//send event to nsq
	//fmt.Println(events)
	for _, event := range events {
		mq.Send(event)
	}

	return CheckResult{Status: status, StatusText: statusText}
}
