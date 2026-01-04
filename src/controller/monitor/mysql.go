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
	_ "dbmcloud/log"
	"dbmcloud/src/database"
	"fmt"
	"net/http"
	"time"

	_ "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "go.uber.org/zap"
)

func MySQLStatus(c *gin.Context) {
	keyword := c.Query("keyword")
	sql := "select * from (select * from status_mysql where gmt_create > DATE_SUB(now(), INTERVAL 5 MINUTE ) order by id desc limit 1000) t group by host,port "
	if keyword != "" {
		sql = sql + fmt.Sprintf(" where host = '%s' or hostname = '%s'", keyword, keyword)
	}
	sql = sql + " order by id desc limit 1000"
	res, _ := database.QueryAll(sql)
	c.JSON(200, gin.H{"success": true, "data": res, "total": len(res)})
}

type Result struct {
	GmtCreate        time.Time
	Connections      int64
	ThreadsConnected int64
	ThreadsRunning   int64
	ThreadsWait      int64
	Queries          int64
	SlowQueries      int64
	BytesReceived    int64
	BytesSent        int64
	ComSelect        int64
	ComInsert        int64
	ComUpdate        int64
	ComDelete        int64
	ComCommit        int64
	ComRollback      int64
	InnodbRowsRead   int64
	InnodbRowsInsert int64
	InnodbRowsUpdate int64
	InnodbRowsDelete int64
}

func MySQLChart(c *gin.Context) {
	var db = database.DB
	params := make(map[string]interface{})
	c.BindJSON(&params)
	host := params["host"].(string)
	port := params["port"].(string)
	startTime := params["start_time"].(string)
	endTime := params["end_time"].(string)
	//ip := c.Query("ip")
	//port := c.Query("port")
	//startTime := c.Query("startTime")
	//endTime := c.Query("endTime")
	var result []Result
	res := db.Raw(fmt.Sprintf("select gmt_create,connections,threads_connected,threads_running,threads_wait,queries,slow_queries,bytes_received,bytes_sent,com_select,com_insert,com_update,com_delete,com_commit,com_rollback,innodb_rows_read,innodb_rows_inserted innodb_rows_inserte,innodb_rows_updated innodb_rows_update,innodb_rows_deleted innodb_rows_delete from status_mysql where host='%s' and port='%s' and gmt_create>='%s' and gmt_create<='%s'", host, port, startTime, endTime)).Scan(&result)
	if res.Error != nil {
		fmt.Print(res.Error)
	}
	var (
		allChartData         = make(map[string]interface{})
		connectionsChartList = make([]map[string]interface{}, 0)
		threadsChartList     = make([]map[string]interface{}, 0)
		queriesChartList     = make([]map[string]interface{}, 0)
		slowQueriesChartList = make([]map[string]interface{}, 0)
		bytesChartList       = make([]map[string]interface{}, 0)
		dmlChartList         = make([]map[string]interface{}, 0)
		trxChartList         = make([]map[string]interface{}, 0)
		innodbChartList      = make([]map[string]interface{}, 0)
	)
	for _, row := range result {
		var eventTime = row.GmtCreate.Format("2006-01-02 15:04:05")
		connectionsChartList = append(connectionsChartList, map[string]interface{}{"time": eventTime, "value": row.Connections, "category": "Connections"})
		threadsChartList = append(threadsChartList, map[string]interface{}{"time": eventTime, "value": row.ThreadsConnected, "category": "Connected"})
		threadsChartList = append(threadsChartList, map[string]interface{}{"time": eventTime, "value": row.ThreadsRunning, "category": "Running"})
		threadsChartList = append(threadsChartList, map[string]interface{}{"time": eventTime, "value": row.ThreadsWait, "category": "Wait"})
		queriesChartList = append(queriesChartList, map[string]interface{}{"time": eventTime, "value": row.Queries, "category": "Queries"})
		slowQueriesChartList = append(slowQueriesChartList, map[string]interface{}{"time": eventTime, "value": row.SlowQueries, "category": "slowQueries"})
		bytesChartList = append(bytesChartList, map[string]interface{}{"time": eventTime, "value": row.BytesReceived / 1024, "category": "BytesReceived"})
		bytesChartList = append(bytesChartList, map[string]interface{}{"time": eventTime, "value": row.BytesSent / 1024, "category": "BytesSent"})
		dmlChartList = append(dmlChartList, map[string]interface{}{"time": eventTime, "value": row.ComSelect, "category": "Select"})
		dmlChartList = append(dmlChartList, map[string]interface{}{"time": eventTime, "value": row.ComInsert, "category": "Insert"})
		dmlChartList = append(dmlChartList, map[string]interface{}{"time": eventTime, "value": row.ComUpdate, "category": "Update"})
		dmlChartList = append(dmlChartList, map[string]interface{}{"time": eventTime, "value": row.ComDelete, "category": "Delete"})
		trxChartList = append(trxChartList, map[string]interface{}{"time": eventTime, "value": row.ComCommit, "category": "Commit"})
		trxChartList = append(trxChartList, map[string]interface{}{"time": eventTime, "value": row.ComRollback, "category": "Rollback"})
		innodbChartList = append(innodbChartList, map[string]interface{}{"time": eventTime, "value": row.InnodbRowsRead, "category": "InnodbRowsRead"})
		innodbChartList = append(innodbChartList, map[string]interface{}{"time": eventTime, "value": row.InnodbRowsInsert, "category": "InnodbRowsInsert"})
		innodbChartList = append(innodbChartList, map[string]interface{}{"time": eventTime, "value": row.InnodbRowsUpdate, "category": "InnodbRowsUpdate"})
		innodbChartList = append(innodbChartList, map[string]interface{}{"time": eventTime, "value": row.InnodbRowsDelete, "category": "InnodbRowsDelete"})

	}
	allChartData["connectionsChartList"] = connectionsChartList
	allChartData["threadsChartList"] = threadsChartList
	allChartData["queriesChartList"] = queriesChartList
	allChartData["slowQueriesChartList"] = slowQueriesChartList
	allChartData["bytesChartList"] = bytesChartList
	allChartData["dmlChartList"] = dmlChartList
	allChartData["trxChartList"] = trxChartList
	allChartData["innodbChartList"] = innodbChartList

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "ok",
		"data":    allChartData,
	})

}
