package helper

import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GBKToUTF8(buff []byte) ([]byte, error) {
	bytes, _, err := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), buff)
	return bytes, err
}

func MapKeys[T map[string]string](m T) []string {
	var tmpKeyList = make([]string, 0)
	for tmpKey, _ := range m {
		tmpKeyList = append(tmpKeyList, tmpKey)
	}
	return tmpKeyList
}
