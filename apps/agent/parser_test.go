package main

// 注意这个地方与上面注释的地方不能有空行，并且不能使用括号如import ("C" "fmt")
import (
	"encoding/json"
	"logagent/util"
	"regexp"
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

func jsonify(log string) {
	data := make(map[string]interface{})
	json.Unmarshal(util.Str2bytes(log), &data)
	// fmt.Println(data)
}

func jsoniterify(log string) {
	data := make(map[string]interface{})
	jsoniter.Unmarshal(util.Str2bytes(log), &data)
	// fmt.Println(data)
}

func split(log string) {
	strings.Split(log, "-")
	// data := strings.Split(log, "-")
	// fmt.Println(data)
}

func regex(log string, cmp *regexp.Regexp) {
	cmp.FindStringSubmatch(log)
	// data := cmp.FindStringSubmatch(log)
	// fmt.Println(data)
}

func regexByte(log []byte, cmp *regexp.Regexp) {
	cmp.FindSubmatch(log)
	// data := cmp.FindSubmatch(log)
	// fmt.Println(data)
}

func BenchmarkJsonify(b *testing.B) {
	log := "{\"timestamp\":\"2020-08-19 22:11:59,515\", \"level\": \"ERROR\", \"message\": \"No matched file: /waf/system_service/genconf/log/^genconf.log\"}"
	jsonify(log)
}

func BenchmarkJsonifyiter(b *testing.B) {
	log := "{\"timestamp\":\"2020-08-19 22:11:59,515\", \"level\": \"ERROR\", \"message\": \"No matched file: /waf/system_service/genconf/log/^genconf.log\"}"
	jsoniterify(log)
}

func BenchmarkSplit(b *testing.B) {
	log := "2020-08-19 22:11:59,515 - ERROR - No matched file: /waf/system_service/genconf/log/^genconf\\.log"
	b.ResetTimer()
	split(log)
}

func BenchmarkRegexp(b *testing.B) {
	log := "2020-08-19 22:11:59,515 - ERROR - No matched file: /waf/system_service/genconf/log/^genconf\\.log"
	cmp := regexp.MustCompile("^(?P<timestamp>.*?) - (?P<levelname>.*?) - (?P<message>.*?)$")
	b.ResetTimer()
	regex(log, cmp)
}

func BenchmarkRegexpByte(b *testing.B) {
	log := []byte("2020-08-19 22:11:59,515 - ERROR - No matched file: /waf/system_service/genconf/log/^genconf\\.log")
	cmp := regexp.MustCompile("^(?P<timestamp>.*?) - (?P<levelname>.*?) - (?P<message>.*?)$")
	b.ResetTimer()
	regexByte(log, cmp)
}
