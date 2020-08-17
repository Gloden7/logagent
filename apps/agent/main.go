package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"logagent/agent"
	"logagent/util"
	"runtime/pprof"
	"strconv"
	"syscall"

	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pidFile := "./pid"
	command := "agent"

	pflag.Int("maxProcs", 1, "使用`maxProcs`个OS线程运行程序")
	pflag.Bool("deamon", false, "是否以守护进程方式运行")
	pflag.String("operation", "start", "操作(start,restart,stop)")
	pflag.String("config", "../../config/config.json", "配置文件路径")
	pflag.Bool("dev", false, "是否以开发模式运行")
	pflag.Bool("cpu", false, "生成 cpu Profiling文件")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetDefault("maxProcs", func() int {
		if numCPU := runtime.NumCPU() / 5; numCPU > 0 {
			return numCPU
		}
		return 1
	}())

	runtime.GOMAXPROCS(viper.GetInt("maxProcs"))

	deamon := viper.GetBool("deamon")

	if deamon {
		args := os.Args[1:]
		for i, v := range args {
			if v == "--deamon" || v == "--deamon=true" {
				args = append(args[:i], args[i+1:]...)
			}
		}

		cmd := exec.Command(os.Args[0], args...)
		cmd.Start()
		fmt.Fprintf(os.Stdout, "[PID] %d\n", cmd.Process.Pid)
		os.Exit(0)
	}

	start := func() {
		if content, err := ioutil.ReadFile(pidFile); err == nil {
			if content, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/comm", util.Bytes2str(content))); err == nil {
				if bytes.Contains(content, util.Str2bytes(command)) {
					fmt.Fprintln(os.Stdout, "[start] PID already exists")
					os.Exit(1)
				}
			}
		}

		confFile := viper.GetString("config")
		viper.SetConfigFile(confFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s", err))
		}

		ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0666)
		defer os.Remove(pidFile)

		if viper.GetBool("cpu") {
			if f, err := os.Create("./cpu.prof"); err == nil {
				pprof.StartCPUProfile(f)
				defer pprof.StopCPUProfile()
			}
		}

		a := agent.New(viper.GetViper())
		a.Main()
	}

	stop := func() {
		content, err := ioutil.ReadFile(pidFile)
		if err != nil {
			return
		}
		pid, err := strconv.Atoi(util.Bytes2str(content))
		if err != nil {
			return
		}
		err = syscall.Kill(pid, syscall.SIGKILL)
		if err != nil {
			fmt.Fprintf(os.Stdout, fmt.Sprintf("[stop] failed, err: %s\n", err))
			return
		}
		os.Remove(pidFile)
		fmt.Fprintf(os.Stdout, fmt.Sprintf("[stop] successfully\n"))
	}

	op := viper.GetString("operation")

	switch op {
	case "start":
		start()
	case "stop":
		stop()
	case "restart":
		stop()
		start()
	}
}
