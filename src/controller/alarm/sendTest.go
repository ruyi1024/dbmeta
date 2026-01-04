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
	"bytes"
	"dbmcloud/src/libary/aliyun"
	"dbmcloud/src/libary/mail"
	"dbmcloud/src/libary/utils"
	"dbmcloud/src/libary/wechat"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func DoSendEmailTest(c *gin.Context) {
	method := c.Request.Method

	if method == "POST" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		fmt.Print(params)
		emailList := params["email_list"].(string)
		mailTo := strings.Split(emailList, ";")
		mailTitle := "Lepus告警通知测试邮件"
		mailContent := "当您收到这份邮件，表明您的邮箱网关配置是正确的，可以正常发送告警邮件."
		if err := mail.Send(mailTo, mailTitle, mailContent); err != nil {
			c.JSON(200, gin.H{"success": false, "msg": "发送邮件失败:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "发送邮件成功",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"msg":     "method not allow.",
	})
	return
}

func DoSendSmsTest(c *gin.Context) {
	method := c.Request.Method

	if method == "POST" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		fmt.Print(params)
		smsList := params["sms_list"].(string)
		TemplateParam := "{\"entity\":\"MySQL-127.0.0.1:3306\",\"title\":\"[测试][QPS过高]\",\"rule\":\"qps(101)>100\",\"time\":\"2022-22-22 12:22:22\"}"
		if err := aliyun.SendSms(smsList, TemplateParam); err != nil {
			c.JSON(200, gin.H{"success": false, "msg": "发送短信失败:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "发送短信成功",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"msg":     "method not allow.",
	})
	return
}

func DoSendPhoneTest(c *gin.Context) {
	method := c.Request.Method

	if method == "POST" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		fmt.Print(params)
		phoneList := params["phone_list"].(string)
		TemplateParam := "{\"title\":\"数据库宕机\"}"
		if err := aliyun.CallPhone(phoneList, TemplateParam); err != nil {
			c.JSON(200, gin.H{"success": false, "msg": "拨打电话失败:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "拨打电话成功",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"msg":     "method not allow.",
	})
	return
}

func DoSendWechatTest(c *gin.Context) {
	method := c.Request.Method

	if method == "POST" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		//userStrList := "o0OjWwQTikvoazf8-OKHaxDMAV6c"
		userStrList := params["wechat_list"].(string)
		currentTime := utils.GetCurrentTime()
		templateData := fmt.Sprintf("{\"first\":{\"value\":\"[MySQL]数据库连接数异常\", \"color\":\"#0000CD\"},\"keyword1\":{\"value\":\"%s\", \"color\":\"#0000CD\"},\"keyword2\":{\"value\":\"192.168.10.100:3306\", \"color\":\"#0000CD\"},\"keyword3\":{\"value\":\"警告\", \"color\":\"#CC6633\"},\"keyword4\":{\"value\":\"ThreadConnected(101)>100\", \"color\":\"#0000CD\"},\"remark\":{\"value\":\"Lepus通知您尽快关注和处理。\", \"color\":\"#0000CD\"}}", currentTime)
		if err := wechat.Send(userStrList, templateData); err != nil {
			c.JSON(200, gin.H{"success": false, "msg": "发送微信失败:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     "发送微信成功",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"msg":     "method not allow.",
	})
	return
}

func DoSendWebhookTest(c *gin.Context) {
	method := c.Request.Method

	if method == "POST" {
		params := make(map[string]interface{})
		c.BindJSON(&params)
		webhookUrl := params["weburl"].(string)
		//post数据
		eventData := map[string]interface{}{
			"alarm_title":  "Lepus告警测试-数据库故障",
			"alarm_rule":   "!=",
			"alarm_value":  1,
			"event_time":   "2022-22-22 22:22:22",
			"event_type":   "MySQL",
			"event_group":  "Prod",
			"event_entity": "127.0.0.1:3306",
			"event_key":    "connect",
			"event_value":  0,
			"event_tag":    "",
		}
		client := &http.Client{Timeout: 3 * time.Second}
		jsonStr, _ := json.Marshal(eventData)
		resp, err := client.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonStr))
		if err != nil {
			c.JSON(200, gin.H{"success": false, "msg": "发送数据失败:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"msg":     fmt.Sprintln("发送数据成功，返回值：", resp.Request),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"msg":     "method not allow.",
	})
	return
}
