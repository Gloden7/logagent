package task

import (
	"regexp"
	"strings"
)

var cmp = regexp.MustCompile("\\$\\{.*?\\}")

func getTemplateFunc(template string) func(msg message) (text string) {
	if len(template) == 0 {
		return func(msg message) string {
			if text, ok := msg["message"].(string); ok {
				return text
			}
			return ""
		}
	}
	return func(msg message) (text string) {
		text = cmp.ReplaceAllStringFunc(template, func(old string) string {
			key := strings.Trim(old, "${}")
			new, ok := msg[key].(string)
			if !ok {
				return old
			}
			return new
		})

		return text
	}
}
