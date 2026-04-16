package license

import (
	"fmt"
	"log"
	"net"
)

var (
	checkImpl   = defaultCheck
	displayImpl = defaultDisplay
)

// RegisterRuntimeHooks 由企业版在 init 中注册完整商业授权实现。
func RegisterRuntimeHooks(check func(), display func()) {
	if check != nil {
		checkImpl = check
	}
	if display != nil {
		displayImpl = display
	}
}

// LocalMachineID 返回当前机器首个可用网卡 MAC（用于 -l 输出及企业版授权校验）。
func LocalMachineID() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Panic("Get local MAC failed")
	}
	for _, inter := range interfaces {
		mac := inter.HardwareAddr
		if mac.String() != "" {
			return mac.String()
		}
	}
	return ""
}

// Check 调用已注册的商业版授权校验实现。
func Check() {
	checkImpl()
}

// Display 显示本机授权信息（-l）。
func Display() {
	if skipByBuildTag() || skipByEnv() {
		fmt.Println("Machine Id:", LocalMachineID())
		fmt.Println("License file display skipped (opensource build or DBMETA_LICENSE_SKIP).")
		return
	}
	displayImpl()
}

func defaultCheck() {
	log.Panic("commercial license checker is not registered")
}

func defaultDisplay() {
	fmt.Println("Machine Id:", LocalMachineID())
	fmt.Println("License details are available in enterprise build only.")
}
