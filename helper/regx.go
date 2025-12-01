package helper

import "regexp"

func SimpleRegEx(plainText, regex string) string {
	//regEx := regexp.MustCompile(`(\d+)`)
	regEx := regexp.MustCompile(regex)
	tmpList := regEx.FindStringSubmatch(plainText)
	if len(tmpList) < 2 {
		return ""
	}
	return tmpList[1]
}

func SimpleRegExList(plainText, regex string) []string {
	regEx := regexp.MustCompile(regex)
	tmpList := regEx.FindStringSubmatch(plainText)
	if len(tmpList) < 2 {
		return nil
	}
	return tmpList
}
