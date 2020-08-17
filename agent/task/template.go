package task

import (
	"regexp"
	"strings"
)

var cmp = regexp.MustCompile("\\$\\{.*?\\}")

func getTemplateFunc(template string) func(msg map[string]interface{}) (text string) {
	return func(msg map[string]interface{}) (text string) {
		text = cmp.ReplaceAllStringFunc(template, func(old string) string {
			key := strings.Trim(old, "${}")
			new, ok := msg[key].(string)
			if !ok {
				return old
			}
			return new
		})
		if text[len(text)-1] != '\x0a' {
			text += "\n"
		}
		return text
	}
}
