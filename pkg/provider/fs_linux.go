// +build dragonfly linux solaris

package provider

import (
	"syscall"
	"time"
)

func getFileCreateTime(filePath string) (string, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(filePath, &st); err != nil {
		return "", err
	}

	return time.Unix(st.Ctim.Sec, 0).Format("02/01/2006, 15:04:05"), nil
}
