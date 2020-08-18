package agent

import (
	"io/ioutil"
	"logagent/agent/task"
	"logagent/logging"
	"logagent/util"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

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

	d := a.cpuLimit > 0
	for _, conf := range a.conf {
		t := task.New(a.logger, conf, d)
		a.addTask(t)
	}
	if d {
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

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					return
				}
				total += val
				if i == 4 {
					idle = val
				}
			}
			return
		}
	}
	return
}

func getCPUUsage() float64 {
	idle0, total0 := getCPUSample()
	time.Sleep(1 * time.Second)
	idle1, total1 := getCPUSample()

	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	cpuUsage := 100 * (totalTicks - idleTicks) / totalTicks
	return cpuUsage
}
