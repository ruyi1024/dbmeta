package license

import (
	"os"
	"strings"
)

func skipByEnv() bool {
	v := strings.TrimSpace(os.Getenv("DBMETA_LICENSE_SKIP"))
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// MustRunLicenseCheck 返回 true 时需要执行 Check()。
// devSkip 来自配置文件 license.devSkip。
func MustRunLicenseCheck(devSkip bool) bool {
	if devSkip {
		return false
	}
	if skipByBuildTag() {
		return false
	}
	if skipByEnv() {
		return false
	}
	return true
}

// SkipReason 用于日志，说明为何跳过校验（无跳过时返回空字符串）。
func SkipReason(devSkip bool) string {
	if devSkip {
		return "license.devSkip=true"
	}
	if skipByBuildTag() {
		return "go build -tags opensource"
	}
	if skipByEnv() {
		return "env DBMETA_LICENSE_SKIP"
	}
	return ""
}
