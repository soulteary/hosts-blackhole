package crypto

import (
	"crypto/md5" //#nosec
	"fmt"
)

func Md5(str string) string {
	data := []byte(str)
	/* #nosec */
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func ETag(data []byte) string {
	/* #nosec */
	return fmt.Sprintf("W/%x", md5.Sum(data))
}
