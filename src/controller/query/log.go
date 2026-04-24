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
*/

package query

import (
	"github.com/ruyi1024/dbmeta/src/module"
)

// WriteLog 在 SQL 执行路径中调用；仅当企业版注册了 module.WriteQueryLog 时写入 query_log。
func WriteLog(username string, datasourceType string, datasource string, queryType string, sqlType string, databaseName string, status string, times int64, content string, doResult string) {
	if module.WriteQueryLog != nil {
		module.WriteQueryLog(username, datasourceType, datasource, queryType, sqlType, databaseName, status, times, content, doResult)
	}
}
