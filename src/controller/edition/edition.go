package edition

import (
	"github.com/ruyi1024/dbmeta/src/module"

	"github.com/gin-gonic/gin"
)

// GetEdition 返回当前进程是否包含商业扩展（enterprise / audit / security 等插件），供前端控制菜单与路由。
func GetEdition(c *gin.Context) {
	commercial := module.HasEnterprise() || module.HasAudit() || module.HasSecurity()
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"commercial": commercial,
		},
	})
}
