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

package query

import (
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"
	"fmt"
)

func CheckPrivilege(username string, datasourceType string, datasource string, databaseName string, tableName string, sqlType string, queryNumber int) (bool, string) {

	var db = database.DB
	var dataList []model.Privilege
	//create/alter判断是否有库权限
	if sqlType == "create" || sqlType == "alter" {
		db = db.Where("username=?", username)
		db = db.Where("datasource_type=?", datasourceType)
		db = db.Where("datasource=?", datasource)
		db = db.Where("grant_type=?", "database")
		db = db.Where("database_name=?", databaseName)
		result := db.First(&dataList)
		if result.Error != nil {
			return false, "没有库权限操作，请先授权后操作."
		}
	}
	//select/insert/update/select先判断表级别权限，再判断库级别权限
	if sqlType == "select" || sqlType == "insert" || sqlType == "update" || sqlType == "delete" {
		//判断是否有表权限
		var db = database.DB
		db = db.Where("username=?", username)
		db = db.Where("datasource_type=?", datasourceType)
		db = db.Where("datasource=?", datasource)
		db = db.Where("grant_type=?", "table")
		db = db.Where("database_name=?", databaseName)
		db = db.Where("table_name=?", tableName)
		result := db.First(&dataList)
		if result.Error != nil {
			//判断是否有库权限
			var db = database.DB
			db = db.Where("username=?", username)
			db = db.Where("datasource_type=?", datasourceType)
			db = db.Where("datasource=?", datasource)
			db = db.Where("grant_type=?", "database")
			db = db.Where("table_name=?", "")
			result := db.First(&dataList)
			if result.Error != nil {
				return false, "没有库表操作权限，请先授权后操作."
			}

		}
	}

	record := dataList[0]
	doSelect := record.DoSelect
	doInsert := record.DoInsert
	doUpdate := record.DoUpdate
	doDelete := record.DoDelete
	doCreate := record.DoCreate
	doAlter := record.DoAlter
	maxSelect := record.MaxSelect
	maxUpdate := record.MaxUpdate
	maxDelete := record.MaxDelete

	if sqlType == "select" && doSelect == 0 {
		return false, fmt.Sprintf("没有表%s.%s查询权限，请先授权后操作.", databaseName, tableName)
	}
	if sqlType == "insert" && doInsert == 0 {
		return false, fmt.Sprintf("没有表%s.%s写入权限，请先授权后操作.", databaseName, tableName)
	}
	if sqlType == "update" && doUpdate == 0 {
		return false, fmt.Sprintf("没有表%s.%s更新权限，请先授权后操作.", databaseName, tableName)
	}
	if sqlType == "delete" && doDelete == 0 {
		return false, fmt.Sprintf("没有表%s.%s删除权限，请先授权后操作.", databaseName, tableName)
	}
	if sqlType == "create" && doCreate == 0 {
		return false, fmt.Sprintf("没有表%s.%s创建权限，请先授权后操作.", databaseName, tableName)
	}
	if sqlType == "alter" && doAlter == 0 {
		return false, fmt.Sprintf("没有表%s.%s变更权限，请先授权后操作.", databaseName, tableName)
	}
	if sqlType == "select" && queryNumber > maxSelect {
		return false, fmt.Sprintf("你查询数据表%s.%s的上限为%s.", databaseName, tableName, maxSelect)
	}
	if sqlType == "update" && queryNumber > maxUpdate {
		return false, fmt.Sprintf("你修改数据表%s.%s的上限为%s.", databaseName, tableName, maxUpdate)
	}
	if sqlType == "delete" && queryNumber > maxDelete {
		return false, fmt.Sprintf("你删除数据表%s.%s的上限为%s.", databaseName, tableName, maxDelete)
	}

	return true, "OK"

}
