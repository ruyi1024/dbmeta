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
	"fmt"
	"os"
	"strings"

	"github.com/ruyi1024/dbmeta/src/license"
)

const Version = "1.0.0-rc.1"

func runtimeVersion() string {
	// 优先读取仓库根目录 VERSION，读取失败时回退内置版本常量。
	raw, err := os.ReadFile("./VERSION")
	if err != nil {
		return Version
	}
	v := strings.TrimSpace(string(raw))
	if v == "" {
		return Version
	}
	return v
}

func help() {
	h := `
Usage: [OPTION] ...
Used to perform some operation commands on the dbmeta, if there is no command, start directly.

Mandatory arguments to long options are mandatory for short options too.
  -h        dispaly help info.
  -v        display this version and exit
  -c        specify the configuration file path, the default is './setting.yml'
  -l        display local machine id and license info.
`
	fmt.Println(h)
}

// ParseCLI 解析命令行；若应直接退出（如 -h/-v/-l），会调用 os.Exit 且不再返回。
func ParseCLI() (configPath string) {
	path := "./setting.yml"
	args := os.Args
	if len(args) >= 2 {
		switch args[1] {
		case "-h":
			help()
			os.Exit(0)
		case "-v":
			fmt.Println(runtimeVersion())
			os.Exit(0)
		case "-c":
			path = args[2]
		case "-l":
			license.Display()
			os.Exit(0)
		default:
			help()
			os.Exit(0)
		}
	}
	return path
}
