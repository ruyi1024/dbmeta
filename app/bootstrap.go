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

package app

import (
	"context"
	"fmt"
	"strings"

	"dbmeta-core/log"
	"dbmeta-core/setting"
	"dbmeta-core/src/database"
	"dbmeta-core/src/license"
	"dbmeta-core/src/module"
)

// BootstrapOptions 控制启动阶段可选行为（如是否执行商业授权文件校验）。
type BootstrapOptions struct {
	// EnforceCommercialLicense 为 true 时，按 setting 与 policy 校验 LICENSE.ENC（供 dbmeta-enterprise 入口使用）。
	// 开源版 main 应传 false，避免默认依赖商业授权文件。
	EnforceCommercialLicense bool
}

// Bootstrap 初始化配置、日志、数据库、消息队列、可选 License 校验及扩展模块后台任务。
func Bootstrap(ctx context.Context, configPath string, opts BootstrapOptions) {
	err := setting.InitSetting(configPath)
	if err != nil {
		fmt.Println(err)
	}

	// 未配置 server.addr 时：企业版默认 :8088，开源版默认 :8086，便于本地同时跑两套而不抢端口。
	if opts.EnforceCommercialLicense && strings.TrimSpace(setting.Setting.Server.Addr) == "" {
		setting.Setting.Server.Addr = ":8088"
	}

	log.InitLogs()

	database.DB = database.InitDb()
	database.SQL = database.InitConnect()
	//database.RDS = database.InitRedis()
	//mq.NSQ = mq.InitNsq()

	if opts.EnforceCommercialLicense {
		if license.MustRunLicenseCheck(setting.Setting.License.DevSkip) {
			license.Check()
		} else {
			reason := license.SkipReason(setting.Setting.License.DevSkip)
			if reason != "" {
				log.Warn("license check skipped (" + reason + ")")
			}
		}
	}

	if err := module.StartBackgroundJobs(ctx); err != nil {
		log.Error(fmt.Sprintf("start extension background jobs error: %v", err))
	}
}
