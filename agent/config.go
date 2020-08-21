package agent

import (
	"io/ioutil"
	"log"
	"logagent/agent/task"
	"logagent/logging"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Conf agent配置结构体
type Conf struct {
	CPULimit float64
	GoNum    int
	Logs     logging.Conf
	Tasks    []*task.Conf
}

// InitConfig 初始化一个配置对象
func InitConfig(confFile string) *Conf {
	content, err := ioutil.ReadFile(confFile)

	if err != nil {
		log.Fatal("invalid config file")
	}

	conf := &Conf{}
	err = json.Unmarshal(content, conf)

	if err != nil {
		log.Fatal("invalid config file")
	}

	return conf
}
