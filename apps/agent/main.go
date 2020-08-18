package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"logagent/agent"
	"logagent/util"
	"os/exec"
	"strconv"
	"syscall"

	"os"
)

func main() {
	pidFile := "./pid"
	command := "main"

	var deamon bool
	var op string
	var confFile string

	flag.StringVar(&confFile, "f", "../../config/config.json", "指定`logagent`配置文件")
	flag.BoolVar(&deamon, "deamon", false, "以守护进程方式运行")
	flag.StringVar(&op, "opt", "start", "指定`logagent`操作start,restart,stop")
	flag.Parse()

	start := func() {
		if deamon {
			args := os.Args[1:]
			for i, v := range args {
				if v == "-deamon" || v == "--deamon" || v == "-deamon=true" || v == "-deamon==true" {
					args = append(args[:i], args[i+1:]...)
				}
			}
			cmd := exec.Command(os.Args[0], args...)
			cmd.Start()
			fmt.Fprintf(os.Stdout, "[PID] %d\n", cmd.Process.Pid)
			os.Exit(0)
		}

		if content, err := ioutil.ReadFile(pidFile); err == nil {
			if content, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/comm", util.Bytes2str(content))); err == nil {
				if bytes.Contains(content, util.Str2bytes(command)) {
					fmt.Fprintln(os.Stdout, "[start] PID already exists")
					os.Exit(1)
				}
			}
		}

		conf := agent.InitConfig(confFile)

		ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0666)
		defer os.Remove(pidFile)

		a := agent.New(conf)
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
