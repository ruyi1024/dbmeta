package privilege

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"dbmcloud/src/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

/*
判断字符是否在数组里面的方法
*/
func in(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true
		}
	}
	return false
}

/*
获取当前用户
*/
func getLoginUsername(c *gin.Context) string {
	//var c *gin.Context
	userinfo, _ := c.Get("loginUser")  //获取用户cookie
	data, _ := json.Marshal(&userinfo) //userinfo返回结果是struct model.Users,需要转换成map
	userMap := make(map[string]interface{})
	json.Unmarshal(data, &userMap)
	loginUsername := userMap["username"].(string)
	return loginUsername
}

/*
执行授权方法
*/
func DoGrant(c *gin.Context) {
	//获取请求参数
	params := make(map[string]string)
	c.BindJSON(&params)
	if len(params) == 0 {
		c.JSON(200, gin.H{"success": false, "msg": "Params Error"})
		return
	}
	username := params["username"]
	datasourceType := params["type"]
	datasource := params["datasource"]
	grantType := params["grant_type"]
	databaseName := params["database"]
	tables := params["tables"]
	privileges := params["privileges"]
	maxSelect := params["max_select"]
	maxUpdate := params["max_update"]
	maxDelete := params["max_delete"]
	reason := params["reason"]
	expireDay := params["expire_day"]
	enable := params["enable"]

	//计算过期日期
	currentTime := time.Now()
	expireTime := currentTime.AddDate(0, 0, utils.StrToInt(expireDay))
	expireDate := fmt.Sprintf("%d-%02d-%02d", expireTime.Year(), expireTime.Month(), expireTime.Day())

	//获取登录用户
	loginUsername := getLoginUsername(c)

	//权限字符串列表拆分为数组，并判断是否拥有权限
	var (
		doInsert int8 = 0
		doUpdate int8 = 0
		doDelete int8 = 0
		doSelect int8 = 0
		doCreate int8 = 0
		doAlter  int8 = 0
	)
	privilegeArray := strings.Split(privileges, ";")
	if in("select", privilegeArray) {
		doSelect = 1
	}
	if in("insert", privilegeArray) {
		doInsert = 1
	}
	if in("update", privilegeArray) {
		doUpdate = 1
	}
	if in("delete", privilegeArray) {
		doDelete = 1
	}
	if in("create", privilegeArray) {
		doCreate = 1
	}
	if in("alter", privilegeArray) {
		doAlter = 1
	}

	//循环表写入或者更新权限
	var db = database.DB
	tableArray := strings.Split(tables, ";")
	for _, tableName := range tableArray {
		var dataList []model.Privilege
		db.Where("username=?", username).Where("datasource=?", datasource).Where("grant_type=?", grantType).Where("database_name=?", databaseName).Where("table_name=?", tableName).Find(&dataList)
		if (len(dataList)) == 0 {
			var record model.Privilege
			record.Username = username
			record.DatasourceType = datasourceType
			record.Datasource = datasource
			record.GrantType = grantType
			record.DatabaseName = databaseName
			record.TableName = tableName
			record.DoSelect = doSelect
			record.DoInsert = doInsert
			record.DoUpdate = doUpdate
			record.DoDelete = doDelete
			record.DoCreate = doCreate
			record.DoAlter = doAlter
			record.MaxSelect = utils.StrToInt(maxSelect)
			record.MaxUpdate = utils.StrToInt(maxUpdate)
			record.MaxDelete = utils.StrToInt(maxDelete)
			t, _ := time.Parse("2006-01-02", expireDate)
			record.ExpireDate = t
			record.Enable = utils.StrToInt(enable)
			record.UserCreated = loginUsername
			record.Reason = reason
			result := db.Create(&record)
			if result.Error != nil {
				c.JSON(200, gin.H{"success": false, "msg": "Insert Error:" + result.Error.Error()})
			}

		} else {
			var record model.Privilege
			record.DoSelect = doSelect
			record.DoInsert = doInsert
			record.DoUpdate = doUpdate
			record.DoDelete = doDelete
			record.DoCreate = doCreate
			record.DoAlter = doAlter
			record.MaxSelect = utils.StrToInt(maxSelect)
			record.MaxUpdate = utils.StrToInt(maxUpdate)
			record.MaxDelete = utils.StrToInt(maxDelete)
			t, _ := time.Parse("2006-01-02", expireDate)
			record.ExpireDate = t
			record.Enable = utils.StrToInt(enable)
			record.UserUpdated = loginUsername
			record.Reason = reason
			//gin里面如果更新为0则字段不会更新，可以使用select指定更新字段解决
			result := db.Model(&record).Select("do_select", "do_insert", "do_update", "do_delete", "do_create", "do_alter", "max_select", "max_update", "max_delete", "reason", "expire_date", "enable", "user_updated").Omit("id").Where("username=?", username).Where("datasource=?", datasource).Where("grant_type=?", grantType).Where("database_name=?", databaseName).Where("table_name=?", tableName).Updates(&record)
			if result.Error != nil {
				c.JSON(200, gin.H{"success": false, "msg": "Update Error:" + result.Error.Error()})
			}
		}
	}
	c.JSON(200, gin.H{"success": true})
	return
}

/*
查询权限列表
*/
func List(c *gin.Context) {
	var db = database.DB
	method := c.Request.Method
	if method == "GET" {
		if c.Query("username") != "" {
			db = db.Where("username=?", c.Query("username"))
		}
		if c.Query("datasource_type") != "" {
			db = db.Where("datasource_type=?", c.Query("datasource_type"))
		}
		if c.Query("grant_type") != "" {
			db = db.Where("grant_type=?", c.Query("grant_type"))
		}
		if c.Query("database_name") != "" {
			db = db.Where("database_name like ? ", "%"+c.Query("database_name")+"%")
		}
		if c.Query("table_name") != "" {
			db = db.Where("table_name like ? ", "%"+c.Query("table_name")+"%")
		}
		sorterMap := make(map[string]string)
		sorterData := c.Query("sorter")
		json.Unmarshal([]byte(sorterData), &sorterMap)
		for sortField, sortOrder := range sorterMap {
			if sortField != "" && sortOrder != "" {
				db = db.Order(fmt.Sprintf("%s %s", sortField, strings.Replace(sortOrder, "end", "", 1)))
			}
		}

		var dataList []model.Privilege
		result := db.Find(&dataList)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Query Error:" + result.Error.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "OK",
			"data":    dataList,
			"total":   len(dataList),
		})
		return

	}
}
