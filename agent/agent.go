package agent

import (
	"logagent/agent/task"
	"logagent/logging"
	"logagent/util"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

// Agent agent结构体
type Agent struct {
	cpuLimit  float64
	waitGroup util.WaitGroupWrapper
	tasks     map[uuid.UUID]*task.Task
	logger    *zap.SugaredLogger
	conf      []*task.Conf
}

// New 创建一个Agent对象
func New(conf *Conf) *Agent {
	if conf.GoNum > 0 {
		runtime.GOMAXPROCS(conf.GoNum)
	}
	logger := logging.New(conf.Logs)
	return &Agent{
		cpuLimit: conf.CPULimit,
		logger:   logger,
		conf:     conf.Tasks,
		tasks:    make(map[uuid.UUID]*task.Task),
	}
}

// Main 主函数
func (a *Agent) Main() {
	deviceID := getDeviceID()
	isLimit := a.cpuLimit > 0
	for _, conf := range a.conf {
		t := task.New(a.logger, conf, isLimit, deviceID)
		a.addTask(t)
	}
	if isLimit {
		go a.watchCPUUsage()
	}

	for k, t := range a.tasks {
		a.waitGroup.Wrap(t.Run)
		a.logger.Infof("task %s start", k)
	}

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-c
		a.close()
	}()

	a.waitGroup.Wait()
}

func (a *Agent) addTask(t *task.Task) {
	key := uuid.NewV4()
	a.tasks[key] = t
}

func (a *Agent) close() error {
	for k, t := range a.tasks {
		err := t.Close()
		if err != nil {
			return err
		}
		a.logger.Infof("task %s stop", k)
	}
	return nil
}

func (a *Agent) watchCPUUsage() {
	for {
		cpuUsage := getCPUUsage()
		if cpuUsage > a.cpuLimit {
			samplingRate := (100 - cpuUsage) / 100
			for _, t := range a.tasks {
				t.SetSamplingRate(samplingRate)
			}
		}
	}
}

func getDeviceID() string {
	cmd := exec.Command("dmidecode", "-s", "system-uuid")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return util.Bytes2str(out)
}
