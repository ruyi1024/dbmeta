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

package dm

import (
	"database/sql"
	"fmt"
	"strings"

	_ "gitee.com/chunanyong/dm"
)

func Connect(host, port, username, password, schema string) (*sql.DB, error) {
	dsn := fmt.Sprintf("dm://%s:%s@%s:%s", username, password, host, port)
	if schema != "" {
		dsn = fmt.Sprintf("%s?schema=%s", dsn, schema)
	}
	db, err := sql.Open("dm", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func QueryAll(db *sql.DB, sqlText string) ([]map[string]interface{}, error) {
	rows, err := db.Query(sqlText)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var list []map[string]interface{}
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			continue
		}
		entry := make(map[string]interface{})
		for i, col := range columns {
			col = strings.ToLower(col)
			v := values[i]
			if b, ok := v.([]byte); ok {
				entry[col] = string(b)
			} else {
				entry[col] = v
			}
		}
		list = append(list, entry)
	}
	return list, nil
}
