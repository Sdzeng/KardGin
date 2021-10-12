package helper

import (
	"fmt"
	"runtime"
)

func PrintErrorWithStack(errCode string, errMsg string) {

	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	fmt.Printf("\nerrcode[%s] errmsg:%s 堆栈 %s\n", errCode, errMsg, string(buf[:n]))
}
