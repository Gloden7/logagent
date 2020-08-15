package agent

import (
	"io/ioutil"
	"logagent/agent/task"
	"logagent/logging"
	"logagent/util"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Agent agent
type Agent struct {
	waitGroup util.WaitGroupWrapper
	tasks     map[uuid.UUID]*task.Task
	viper     *viper.Viper
	logger    *zap.SugaredLogger
}

// New create a new agent
func New(viper *viper.Viper) *Agent {

	logsConf := logging.Conf{}
	if err := viper.UnmarshalKey("logs", &logsConf); err != nil {
		panic(err)
	}
	logger := logging.New(logsConf, viper.GetBool("dev"))

	return &Agent{
		viper:  viper,
		logger: logger,
		tasks:  make(map[uuid.UUID]*task.Task),
	}
}

// Main agent main
func (a *Agent) Main() {

	confs := []*task.Conf{}
	if err := viper.UnmarshalKey("tasks", &confs); err != nil {
		a.logger.Panic(err)
	}

	cpulimit := viper.GetFloat64("cpuLimit")

	if cpulimit > 0 {
		for _, conf := range confs {
			t := task.New(a.logger, conf, true)
			a.addTask(t)
		}
		go a.watchCPUUsage(cpulimit)
	} else {
		for _, conf := range confs {
			t := task.New(a.logger, conf, false)
			a.addTask(t)
		}
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

func (a *Agent) watchCPUUsage(cpuli float64) {
	for {
		cpuUsage := getCPUUsage()
		if cpuUsage > cpuli {
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
