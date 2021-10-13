package helper

import (
	"fmt"
	"runtime"
)

func PrintError(errCode string, errMsg string, stack bool) {
	if stack {
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		fmt.Printf("\nerrcode[%s] errmsg:%s \n堆栈==> %s", errCode, errMsg, string(buf[:n]))
	} else {
		fmt.Printf("\nerrcode[%s] errmsg:%s", errCode, errMsg)

	}
}
