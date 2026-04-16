//go:build !opensource

package license

func skipByBuildTag() bool {
	return false
}
