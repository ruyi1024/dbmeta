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
	"html/template"
	"net/http"

	"dbmeta-core/log"
	"dbmeta-core/router"
	"dbmeta-core/setting"
	"dbmeta-core/webassets"

	"github.com/gin-gonic/gin"
)

// RunHTTPServer 启动 Gin HTTP 服务（含内置前端静态资源）。
func RunHTTPServer() error {
	r := router.Router()
	r.Use(log.HandleLogger(log.Logger), log.HandleRecovery(log.Logger, true))

	r.SetHTMLTemplate(template.Must(template.New("").ParseFS(webassets.IndexHTML, "index.html")))
	r.StaticFS("/public/", http.FS(webassets.Static))

	r.GET("/logo.png", func(c *gin.Context) {
		c.Request.URL.Path = "/public/static/logo.png"
		r.HandleContext(c)
	})
	r.GET("/avatar.jpg", func(c *gin.Context) {
		c.Request.URL.Path = "/public/static/avatar.jpg"
		r.HandleContext(c)
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", "")
	})
	return r.Run(setting.ListenAddr())
}
