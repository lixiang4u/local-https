package helper

import "encoding/json"

func JsonString(v interface{}) string {
	//buf, _ := json.MarshalIndent(v, "", "\t")
	buf, _ := json.Marshal(v)
	//log.Println("[JSON]", string(buf))
	return string(buf)
}
