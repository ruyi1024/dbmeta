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

package service

import (
	"github.com/ruyi1024/dbmeta/log"
	"github.com/ruyi1024/dbmeta/src/database"
	"github.com/ruyi1024/dbmeta/src/model"

	"go.uber.org/zap"
)

const datasourceStatusTextMaxRunes = 500

// TruncateDatasourceStatusText 截断为数据源表 status_text 可存储长度（按 rune，避免截断多字节字符）。
func TruncateDatasourceStatusText(s string) string {
	r := []rune(s)
	if len(r) <= datasourceStatusTextMaxRunes {
		return s
	}
	return string(r[:datasourceStatusTextMaxRunes])
}

// UpdateDatasourceConnectionStatus 将连接检查结果写回 datasource 表（供计划任务与测试连接等调用）。
func UpdateDatasourceConnectionStatus(id int, status int32, statusText string) {
	if id <= 0 {
		return
	}
	res := database.DB.Model(&model.Datasource{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      status,
		"status_text": TruncateDatasourceStatusText(statusText),
	})
	if res.Error != nil {
		log.Logger.Error("更新数据源连接状态失败", zap.Int("id", id), zap.Error(res.Error))
	}
}
