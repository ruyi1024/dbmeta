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

package alarm

import (
	"dbmcloud/src/database"
	"dbmcloud/src/model"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ChannelList(c *gin.Context) {
	var db = database.DB
	method := c.Request.Method
	if method == "GET" {
		if c.Query("enable") != "" {
			db = db.Where("enable=?", c.Query("enable"))
		}
		if c.Query("mail_enable") != "" {
			db = db.Where("mail_enable=?", c.Query("mail_enable"))
		}
		if c.Query("sms_enable") != "" {
			db = db.Where("sms_enable=?", c.Query("sms_enable"))
		}
		if c.Query("phone_enable") != "" {
			db = db.Where("phone_enable=?", c.Query("phone_enable"))
		}
		if c.Query("wechat_enable") != "" {
			db = db.Where("wechat_enable=?", c.Query("wechat_enable"))
		}
		if c.Query("name") != "" {
			db = db.Where("name like ? ", "%"+c.Query("name")+"%")
		}
		if c.Query("mail_list") != "" {
			db = db.Where("mail_list like ? ", "%"+c.Query("mail_list")+"%")
		}
		if c.Query("sms_list") != "" {
			db = db.Where("sms_list like ? ", "%"+c.Query("sms_list")+"%")
		}
		if c.Query("phone_list") != "" {
			db = db.Where("phone_list like ? ", "%"+c.Query("phone_list")+"%")
		}
		if c.Query("wechat_list") != "" {
			db = db.Where("wechat_list like ? ", "%"+c.Query("wechat_list")+"%")
		}
		sorterMap := make(map[string]string)
		sorterData := c.Query("sorter")
		json.Unmarshal([]byte(sorterData), &sorterMap)
		for sortField, sortOrder := range sorterMap {
			if sortField != "" && sortOrder != "" {
				db = db.Order(fmt.Sprintf("%s %s", sortField, strings.Replace(sortOrder, "end", "", 1)))
			}
		}

		var dataList []model.AlarmChannel
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
	if method == "POST" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		fmt.Print(params)
		data := model.AlarmChannel{Name: params["name"].(string), Description: params["description"].(string), MailEnable: int(params["mail_enable"].(float64)), SmsEnable: int(params["sms_enable"].(float64)), PhoneEnable: int(params["phone_enable"].(float64)), WechatEnable: int(params["wechat_enable"].(float64)), WebhookEnable: int(params["webhook_enable"].(float64)), Enable: int(params["enable"].(float64)), MailList: params["mail_list"].(string), SmsList: params["sms_list"].(string), PhoneList: params["phone_list"].(string), WechatList: params["wechat_list"].(string), WebhookUrl: params["webhook_url"].(string)}
		result := db.Create(&data)
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Insert Error:" + result.Error.Error()})
			return
		}

	}
	if method == "PUT" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		result := db.Model(&model.AlarmChannel{}).Where("id=?", params["id"].(float64)).Updates(map[string]interface{}{"Name": params["name"].(string), "Description": params["description"].(string), "Enable": params["enable"].(float64), "MailEnable": params["mail_enable"].(float64), "PhoneEnable": params["phone_enable"].(float64), "SmsEnable": params["sms_enable"].(float64), "WechatEnable": params["wechat_enable"].(float64), "WebhookEnable": params["webhook_enable"].(float64), "MailList": params["mail_list"].(string), "SmsList": params["sms_list"].(string), "PhoneList": params["phone_list"].(string), "WechatList": params["wechat_list"].(string), "WebhookUrl": params["webhook_url"].(string)})
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Update Error:" + result.Error.Error()})
			return
		}
	}
	if method == "DELETE" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		result := db.Where("id = ?", params["id"].(float64)).Delete(&model.AlarmChannel{})
		if result.Error != nil {
			c.JSON(200, gin.H{"success": false, "msg": "Delete Error:" + result.Error.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
	})
	return
}
