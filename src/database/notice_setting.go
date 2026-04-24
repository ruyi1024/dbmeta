package database

import (
	"github.com/ruyi1024/dbmeta/setting"
	"github.com/ruyi1024/dbmeta/src/model"

	"gorm.io/gorm"
)

const noticeCategory = "notice"

var noticeConfigKeys = []string{
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

func noticeFromRows(rows []model.SettingKV) setting.Notice {
	var n setting.Notice
	for _, row := range rows {
		switch row.ConfigKey {
		case "mailHost":
			n.MailHost = row.ConfigValue
		case "mailPort":
			n.MailPort = row.ConfigValue
		case "mailUser":
			n.MailUser = row.ConfigValue
		case "mailPass":
			n.MailPass = row.ConfigValue
		case "mailFrom":
			n.MailFrom = row.ConfigValue
		case "accessKeyId":
			n.AccessKeyId = row.ConfigValue
		case "accessKeySecret":
			n.AccessKeySecret = row.ConfigValue
		case "smsSignName":
			n.SmsSignName = row.ConfigValue
		case "smsTemplateCode":
			n.SmsTemplateCode = row.ConfigValue
		case "phoneTemplateCode":
			n.PhoneTemplateCode = row.ConfigValue
		case "phonePlayTimes":
			n.PhonePlayTimes = row.ConfigValue
		case "wechatAppId":
			n.WechatAppId = row.ConfigValue
		case "wechatAppSecret":
			n.WechatAppSecret = row.ConfigValue
		case "wechatSendTemplateId":
			n.WechatSendTemplateId = row.ConfigValue
		}
	}
	return n
}

func loadNoticeIntoSetting(db *gorm.DB) error {
	var rows []model.SettingKV
	if err := db.Where("category = ? AND config_key IN ?", noticeCategory, noticeConfigKeys).Find(&rows).Error; err != nil {
		return err
	}
	setting.SetNotice(noticeFromRows(rows))
	return nil
}

// LoadNoticeIntoSetting 从 settings 表刷新通信配置到内存（启动后或 API 更新后调用）。
func LoadNoticeIntoSetting() error {
	if DB == nil {
		return nil
	}
	return loadNoticeIntoSetting(DB)
}
