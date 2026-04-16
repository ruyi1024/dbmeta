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

package task

import (
	"database/sql"
	"fmt"

	"dbmeta-core/src/database"
	"dbmeta-core/src/model"
)

// connectToDatabase 连接业务库（原 task/analysis.go，数据质量等任务复用）。
func connectToDatabase(datasource model.Datasource, password string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch datasource.Type {
	case "MySQL", "TiDB", "Doris", "MariaDB", "GreatSQL", "OceanBase":
		db, err = database.Connect(
			database.WithDriver("mysql"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithDatabase(datasource.Dbid))
	case "PostgreSQL":
		db, err = database.Connect(
			database.WithDriver("postgres"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithDatabase(datasource.Dbid))
	case "ClickHouse":
		db, err = database.Connect(
			database.WithDriver("clickhouse"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithDatabase(datasource.Dbid))
	case "Oracle":
		db, err = database.Connect(
			database.WithDriver("godror"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithSid(datasource.Dbid))
	case "SQLServer":
		db, err = database.Connect(
			database.WithDriver("mssql"),
			database.WithHost(datasource.Host),
			database.WithPort(datasource.Port),
			database.WithUsername(datasource.User),
			database.WithPassword(password),
			database.WithDatabase(datasource.Dbid))
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", datasource.Type)
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}
