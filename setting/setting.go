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

package setting

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type setting struct {
	Log        `yam:"log"`
	Server     ServerConfig `yaml:"server"`
	DataSource `yaml:"dataSource"`
	// Notice 仅来自数据库 settings 表（category=notice），不从 YAML 读取
	Notice  `yaml:"-"`
	Decrypt `yaml:"decrypt"`
	Token   `yaml:"token"`
	AI      `yaml:"ai"`
	License LicenseConfig `yaml:"license"`
}

// ServerConfig HTTP 服务监听。
type ServerConfig struct {
	// Addr 监听地址，如 ":8086" 或 "127.0.0.1:9090"。为空则默认 ":8086"。
	Addr string `yaml:"addr"`
}

// LicenseConfig 与 src/license 包配合：生产环境应执行完整校验。
type LicenseConfig struct {
	// DevSkip 为 true 时跳过 LICENSE.ENC 与机器码等校验，仅用于本地开发；生产环境务必为 false。
	DevSkip bool `yaml:"devSkip"`
}

type Log struct {
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
	Debug bool   `yaml:"debug"`
}

type DataSource struct {
	Host               string `yaml:"host"`
	Port               string `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	Database           string `yaml:"database"`
	RedisHost          string `yaml:"redisHost"`
	RedisPort          string `yaml:"redisPort"`
	RedisPassword      string `yaml:"redisPassword"`
	ClickhouseHost     string `yaml:"clickhouseHost"`
	ClickhousePort     string `yaml:"clickhousePort"`
	ClickhouseUser     string `yaml:"clickhouseUser"`
	ClickhousePassword string `yaml:"clickhousePassword"`
	ClickhouseDatabase string `yaml:"clickhouseDatabase"`
}

type Notice struct {
	MailHost             string `yaml:"mailHost"`
	MailPort             string `yaml:"mailPort"`
	MailUser             string `yaml:"mailUser"`
	MailPass             string `yaml:"mailPass"`
	MailFrom             string `yaml:"mailFrom"`
	AccessKeyId          string `yaml:"accessKeyId"`
	AccessKeySecret      string `yaml:"accessKeySecret"`
	SmsSignName          string `yaml:"smsSignName"`
	SmsTemplateCode      string `yaml:"smsTemplateCode"`
	PhoneTemplateCode    string `yaml:"phoneTemplateCode"`
	PhonePlayTimes       string `yaml:"phonePlayTimes"`
	WechatAppId          string `yaml:"wechatAppId"`
	WechatAppSecret      string `yaml:"wechatAppSecret"`
	WechatSendTemplateId string `yaml:"wechatSendTemplateId"`
}

type Decrypt struct {
	SignKey      string `yaml:"signKey"`
	DbPassKey    string `yaml:"dbPassKey"`
	Md5Iteration int
}

type Token struct {
	TokenKey     string `yaml:"key"`
	TokenName    string `yaml:"name"`
	Expired      string `yaml:"expired"`
	TokenExpired int64
}

type AI struct {
	DeepseekApiKey string      `yaml:"deepseekApiKey"`
	DeepseekApiUrl string      `yaml:"deepseekApiUrl"`
	DeepseekModel  string      `yaml:"deepseekModel"`
	Timeout        int         `yaml:"timeout"`
	DifyBaseUrl    string      `yaml:"difyBaseUrl"`
	DifyTimeout    int         `yaml:"difyTimeout"`
	Agents         []DifyAgent `yaml:"agents"`
}

// DifyAgent 智能体配置
type DifyAgent struct {
	ID              string      `yaml:"id"`
	Name            string      `yaml:"name"`
	Icon            string      `yaml:"icon"`
	Description     string      `yaml:"description"`
	ApiKey          string      `yaml:"apiKey"`
	Enabled         bool        `yaml:"enabled"`
	WelcomeMessage  string      `yaml:"welcomeMessage"`
	CustomQuestions []string    `yaml:"customQuestions"`
	GuideQuestions  []string    `yaml:"guideQuestions"`
	GuideFlows      []GuideFlow `yaml:"guideFlows"`
}

// GuideFlow 引导流程配置
type GuideFlow struct {
	ID    string      `yaml:"id"`
	Title string      `yaml:"title"`
	Steps []GuideStep `yaml:"steps"`
}

// GuideStep 引导步骤配置
type GuideStep struct {
	Type         string        `yaml:"type"`         // message, select, input, submit
	Content      string        `yaml:"content"`      // 消息内容
	Field        string        `yaml:"field"`        // 字段名
	Placeholder  string        `yaml:"placeholder"`  // 占位符
	Options      []GuideOption `yaml:"options"`      // 选项列表
	FinalMessage string        `yaml:"finalMessage"` // 最终提交消息
}

// GuideOption 选项配置
type GuideOption struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

var Setting = new(setting)

// ConfigBaseDir 为当前加载的配置文件所在目录的绝对路径，供 LICENSE.ENC 等相对配置文件解析；解析失败时为空。
var ConfigBaseDir string

// ConfigPath 为当前加载配置文件的绝对路径，供运行中更新配置后持久化回文件。
var ConfigPath string

func InitSetting(path string) (err error) {
	ConfigBaseDir = ""
	ConfigPath = ""
	if path != "" {
		if abs, err := filepath.Abs(path); err == nil {
			ConfigBaseDir = filepath.Dir(abs)
			ConfigPath = abs
		}
	}
	if f, err := os.Open(path); err == nil {
		defer func() {
			_ = f.Close()
		}()
		c, err := ioutil.ReadAll(f)
		if err != nil {
			return errors.Wrap(err, "init setting read file")
		}
		err = yaml.Unmarshal(c, &Setting)
		if err != nil {
			return errors.Wrap(err, "init setting unmarshal data")
		}
		//fmt.Println(fmt.Sprintf("config:%#v", Setting))
	}
	Setting.Md5Iteration = 1500
	// token expired
	if strings.HasSuffix(strings.ToLower(Setting.Expired), "h") {
		s := strings.Replace(Setting.Expired, "h", "", -1)
		h, _ := strconv.ParseInt(s, 10, 64)
		Setting.TokenExpired = h * 60 * 60
	}

	if strings.HasSuffix(strings.ToLower(Setting.Expired), "d") {
		s := strings.Replace(Setting.Expired, "d", "", -1)
		h, _ := strconv.ParseInt(s, 10, 64)
		Setting.TokenExpired = h * 60 * 60 * 24
	}
	return nil
}

func DataSourceInfo() DataSource {
	return Setting.DataSource
}

// NoticeInfo 返回当前通信配置（mail/aliyun/wechat），由数据库 settings 表加载后缓存在内存。
func NoticeInfo() Notice {
	return Setting.Notice
}

// SetNotice 由数据库加载层写入内存，供邮件/短信/微信等包读取。
func SetNotice(n Notice) {
	Setting.Notice = n
}

// ListenAddr 返回 Gin HTTP 监听地址（配置文件 server.addr）。
// 未配置时默认 :8086；企业版入口会在 Bootstrap 中将空值设为 :8088（见 app/bootstrap.go）。
func ListenAddr() string {
	s := strings.TrimSpace(Setting.Server.Addr)
	if s == "" {
		return ":8086"
	}
	return s
}
