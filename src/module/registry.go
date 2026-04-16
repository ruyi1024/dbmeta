package module

import (
	"context"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module 所有扩展模块都需要实现 Name。
type Module interface {
	Name() string
}

// RouteRegistrar 可选：注册额外路由。
type RouteRegistrar interface {
	RegisterRoutes(v1 *gin.RouterGroup)
}

// MigrationRegistrar 可选：注册额外数据库迁移。
type MigrationRegistrar interface {
	RegisterMigrations(db *gorm.DB) error
}

// BackgroundJobStarter 可选：启动额外后台任务。
type BackgroundJobStarter interface {
	StartBackgroundJobs(ctx context.Context) error
}

var (
	mu      sync.RWMutex
	modules = make(map[string]Module)
)

// Register 注册模块；同名模块会被覆盖，便于外部按需替换。
func Register(m Module) {
	if m == nil {
		return
	}
	name := m.Name()
	if name == "" {
		return
	}
	mu.Lock()
	modules[name] = m
	mu.Unlock()
}

// List 返回已注册模块快照。
func List() []Module {
	mu.RLock()
	defer mu.RUnlock()
	items := make([]Module, 0, len(modules))
	for _, m := range modules {
		items = append(items, m)
	}
	return items
}

// HasEnterprise 为 true 表示已加载 dbmeta-enterprise 等企业扩展模块（与仅开源 core 二进制区分）。
func HasEnterprise() bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := modules["enterprise"]
	return ok
}

// HasAudit 为 true 表示已加载审计中心插件（dbmeta-enterprise/audit，模块名 audit）。
func HasAudit() bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := modules["audit"]
	return ok
}

// HasSecurity 为 true 表示已加载数据安全插件（dbmeta-enterprise/security，模块名 security）。
func HasSecurity() bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := modules["security"]
	return ok
}

// RegisterRoutes 调用所有模块的路由注册（如果实现了该能力）。
func RegisterRoutes(v1 *gin.RouterGroup) {
	for _, m := range List() {
		if r, ok := m.(RouteRegistrar); ok {
			r.RegisterRoutes(v1)
		}
	}
}

// ApplyMigrations 调用所有模块的数据库迁移（如果实现了该能力）。
func ApplyMigrations(db *gorm.DB) error {
	for _, m := range List() {
		if r, ok := m.(MigrationRegistrar); ok {
			if err := r.RegisterMigrations(db); err != nil {
				return fmt.Errorf("module %s migrations failed: %w", m.Name(), err)
			}
		}
	}
	return nil
}

// StartBackgroundJobs 启动所有模块后台任务（如果实现了该能力）。
func StartBackgroundJobs(ctx context.Context) error {
	for _, m := range List() {
		if r, ok := m.(BackgroundJobStarter); ok {
			if err := r.StartBackgroundJobs(ctx); err != nil {
				return fmt.Errorf("module %s background jobs failed: %w", m.Name(), err)
			}
		}
	}
	return nil
}
