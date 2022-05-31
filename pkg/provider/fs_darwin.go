// +build darwin freebsd netbsd openbsd

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

	return time.Unix(st.Ctimespec.Sec, 0).Format("02/01/2006, 15:04:05"), nil
}
