package agent

import (
	"encoding/json"
	"io/ioutil"
	"logagent/agent/task"
	"logagent/logging"
)

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
		panic("invalid config file")
	}

	conf := &Conf{}
	err = json.Unmarshal(content, conf)

	if err != nil {
		panic("invalid config file")
	}

	return conf
}
