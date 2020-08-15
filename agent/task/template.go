package task

import (
	"bytes"
	"fmt"
	"logagent/util"
	"regexp"
)

var cmp = regexp.MustCompile("\\$\\{.*?\\}")

func templates(template []byte, msg map[string]interface{}) []byte {
	content := cmp.ReplaceAllFunc(template, func(old []byte) []byte {
		key := util.Bytes2str(bytes.Trim(old, "${}"))
		new, ok := msg[key]
		if !ok {
			return old
		}

		switch v := new.(type) {
		case []byte:
			return v
		default:
			return util.Str2bytes(fmt.Sprint(v))
		}
	})
	if len(content) > 0 && content[len(content)-1] != '\x0a' {
		content = append(content, '\x0a')
	}
	return content
}
