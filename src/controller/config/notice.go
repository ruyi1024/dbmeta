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

package config

import (
	"net/http"

	"dbmeta-core/setting"
	"dbmeta-core/src/database"
	"dbmeta-core/src/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

const noticeCategory = "notice"

var noticeKeys = []string{
	"mailHost",
	"mailPort",
	"mailUser",
	"mailPass",
	"mailFrom",
	"accessKeyId",
	"accessKeySecret",
	"smsSignName",
	"smsTemplateCode",
	"phoneTemplateCode",
	"phonePlayTimes",
	"wechatAppId",
	"wechatAppSecret",
	"wechatSendTemplateId",
}

func upsertNoticeKV(items map[string]string) error {
	if len(items) == 0 {
		return nil
	}
	rows := make([]model.SettingKV, 0, len(items))
	for k, v := range items {
		rows = append(rows, model.SettingKV{
			Category:    noticeCategory,
			ConfigKey:   k,
			ConfigValue: v,
		})
	}
	return database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"category", "config_value", "gmt_updated"}),
	}).Create(&rows).Error
}

func loadNoticeFromDB() (setting.Notice, error) {
	if err := database.LoadNoticeIntoSetting(); err != nil {
		return setting.Notice{}, err
	}
	return setting.NoticeInfo(), nil
}

// NoticeGet 获取通信配置（mail / aliyun / wechat）。
func NoticeGet(c *gin.Context) {
	data, err := loadNoticeFromDB()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"msg":     "Query Error: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "OK",
		"data": map[string]interface{}{
			"mailHost":             data.MailHost,
			"mailPort":             data.MailPort,
			"mailUser":             data.MailUser,
			"mailPass":             data.MailPass,
			"mailFrom":             data.MailFrom,
			"accessKeyId":          data.AccessKeyId,
			"accessKeySecret":      data.AccessKeySecret,
			"smsSignName":          data.SmsSignName,
			"smsTemplateCode":      data.SmsTemplateCode,
			"phoneTemplateCode":    data.PhoneTemplateCode,
			"phonePlayTimes":       data.PhonePlayTimes,
			"wechatAppId":          data.WechatAppId,
			"wechatAppSecret":      data.WechatAppSecret,
			"wechatSendTemplateId": data.WechatSendTemplateId,
		},
	})
}

func NoticeMailPut(c *gin.Context) {
	var req struct {
		MailFrom string `json:"mailFrom"`
		MailHost string `json:"mailHost"`
		MailPass string `json:"mailPass"`
		MailPort string `json:"mailPort"`
		MailUser string `json:"mailUser"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Params Error: " + err.Error()})
		return
	}
	if err := upsertNoticeKV(map[string]string{
		"mailHost": req.MailHost,
		"mailPort": req.MailPort,
		"mailUser": req.MailUser,
		"mailPass": req.MailPass,
		"mailFrom": req.MailFrom,
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Update Error: " + err.Error()})
		return
	}
	_ = database.LoadNoticeIntoSetting()
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "OK"})
}

func NoticeAliyunPut(c *gin.Context) {
	var req struct {
		AccessKeyId       string `json:"accessKeyId"`
		AccessKeySecret   string `json:"accessKeySecret"`
		PhonePlayTimes    string `json:"phonePlayTimes"`
		PhoneTemplateCode string `json:"phoneTemplateCode"`
		SmsSignName       string `json:"smsSignName"`
		SmsTemplateCode   string `json:"smsTemplateCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Params Error: " + err.Error()})
		return
	}
	if err := upsertNoticeKV(map[string]string{
		"accessKeyId":       req.AccessKeyId,
		"accessKeySecret":   req.AccessKeySecret,
		"smsSignName":       req.SmsSignName,
		"smsTemplateCode":   req.SmsTemplateCode,
		"phoneTemplateCode": req.PhoneTemplateCode,
		"phonePlayTimes":    req.PhonePlayTimes,
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Update Error: " + err.Error()})
		return
	}
	_ = database.LoadNoticeIntoSetting()
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "OK"})
}

func NoticeWechatPut(c *gin.Context) {
	var req struct {
		WechatAppId          string `json:"wechatAppId"`
		WechatAppSecret      string `json:"wechatAppSecret"`
		WechatSendTemplateId string `json:"wechatSendTemplateId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Params Error: " + err.Error()})
		return
	}
	if err := upsertNoticeKV(map[string]string{
		"wechatAppId":          req.WechatAppId,
		"wechatAppSecret":      req.WechatAppSecret,
		"wechatSendTemplateId": req.WechatSendTemplateId,
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "Update Error: " + err.Error()})
		return
	}
	_ = database.LoadNoticeIntoSetting()
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "OK"})
}
