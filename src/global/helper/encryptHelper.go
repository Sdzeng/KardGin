package helper

import (
	"crypto/md5"
	"encoding/hex"
)

// var (
// 	md5Hash hash.Hash
// )

// func init() {
// 	md5Hash = md5.New()
// }

// func StrMd5(str string) string {
// 	md5Hash.Write([]byte(str))
// 	return hex.EncodeToString(md5Hash.Sum(nil))
// }

func StrMd5(str string) string {
	// md5Hash.Write([]byte(str))
	md5sum := md5.Sum([]byte(str))
	return hex.EncodeToString(md5sum[:])
}
