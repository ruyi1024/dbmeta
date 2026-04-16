package module

import "sync"

var commercialTaskHandlers sync.Map // task_key -> func()

// RegisterCommercialTaskHandler 由企业版模块注册仅商业版存在的计划任务执行函数（供 /task/option/execute 等调用）。
func RegisterCommercialTaskHandler(taskKey string, fn func()) {
	if taskKey == "" || fn == nil {
		return
	}
	commercialTaskHandlers.Store(taskKey, fn)
}

// RunCommercialTask 若已注册则执行并返回 true。
func RunCommercialTask(taskKey string) bool {
	v, ok := commercialTaskHandlers.Load(taskKey)
	if !ok {
		return false
	}
	fn, ok := v.(func())
	if !ok {
		return false
	}
	fn()
	return true
}

// HasCommercialEdition 是否加载了任一企业扩展模块（用于 commercial_only 任务展示与执行判断）。
func HasCommercialEdition() bool {
	return HasEnterprise() || HasAudit() || HasSecurity()
}
