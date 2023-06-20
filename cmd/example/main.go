package main

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/rprtr258/log"
	"github.com/rprtr258/xerr"
)

// example enum
type StatusType int

const (
	StatusInvalid StatusType = iota
	StatusStarting
	StatusRunning
	StatusStopped
)

func (ps StatusType) String() string {
	switch ps {
	case StatusInvalid:
		return "invalid"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopped:
		return "stopped"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", ps)
	}
}

// example struct
type Status struct {
	StartTime time.Time
	Status    StatusType
	Pid       int
	CPU       uint64
	Memory    uint64
	ExitCode  int
}

func main() {
	fields := log.F{
		"int":  1,
		"str":  "aboba",
		"list": []string{"a", "b", "c"},
		"ts":   time.Now(),
		"status": Status{
			StartTime: time.Now(),
			Status:    StatusRunning,
			Pid:       123,
			CPU:       300,
			Memory:    1000,
			ExitCode:  0,
		},
	}

	log.Debug("debug msg")
	log.Debugf("debug msg with fields", fields)
	log.Info("info msg")
	log.Infof("info msg with fields", fields)
	log.Warn("warn msg")
	log.Warnf("warn msg with fields", fields)
	log.Error("error msg")
	log.Errorf("error msg with fields", fields)

	l1 := log.Tag("tag1")
	l1.Debug("debug msg")
	l1.Debugf("debug msg with fields", fields)
	l1.Info("info msg")
	l1.Infof("info msg with fields", fields)
	l1.Warn("warn msg")
	l1.Warnf("warn msg with fields", fields)
	l1.Error("error msg")
	l1.Errorf("error msg with fields", fields)

	l2 := l1.Tag("tag2")
	l2.Debug("debug msg")
	l2.Debugf("debug msg with fields", fields)
	l2.Info("info msg")
	l2.Infof("info msg with fields", fields)
	l2.Warn("warn msg")
	l2.Warnf("warn msg with fields", fields)
	l2.Error("error msg")
	l2.Errorf("error msg with fields", fields)

	l3 := log.With(log.F{"kupi": "doru"})
	l3.Debug("debug msg")
	l3.Debugf("debug msg with fields", fields)
	l3.Info("info msg")
	l3.Infof("info msg with fields", fields)
	l3.Warn("warn msg")
	l3.Warnf("warn msg with fields", fields)
	l3.Error("error msg")
	l3.Errorf("error msg with fields", fields)

	err1 := xerr.NewM("xerr with fields", xerr.Fields{
		"int":  1,
		"str":  "aboba",
		"list": []string{"a", "b", "c"},
		"ts":   time.Now(),
		"status": Status{
			StartTime: time.Now(),
			Status:    StatusRunning,
			Pid:       123,
			CPU:       300,
			Memory:    1000,
			ExitCode:  0,
		},
	})
	log.Errorf("error happened", log.F{"err": err1})
	err2 := xerr.New(xerr.Fields{
		"int": 1,
		"str": "aboba",
		"ts":  time.Now(),
	})
	log.Errorf("plain error happened", log.F{"err": err2})
	log.Errorf("combined error happened", log.F{
		"err": xerr.NewM("aboba", xerr.Errors{err1, err2}),
	})
	log.Errorf("deeply embedded error happened", log.F{
		"err": xerr.Combine(
			xerr.Combine(
				xerr.NewWM(&exec.Error{
					Name: "jjjjjjjjj",
					Err:  errors.New("executable file not found in $PATH"),
				}, "look for executable path"),
			),
		),
	})
	log.Infof("pointer to info", log.F{"data": &map[string]any{
		"StartTime": time.Now(),
		"Status":    StatusRunning,
		"Pid":       123,
		"CPU":       300,
		"Memory":    []int{1000, 2000, 3000},
		"ExitCode":  0,
	}})
}
