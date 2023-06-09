package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/rprtr258/xerr"
	"golang.org/x/exp/slog"

	"github.com/rprtr258/log"
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
	StartTime  time.Time
	Status     StatusType
	RunningFor time.Duration
	Pid        int
	CPU        uint64
	Memory     uint64
	ExitCode   int
}

func run(l slog.Handler) {
	slog.SetDefault(slog.New(l))

	st := Status{
		StartTime:  time.Now(),
		Status:     StatusRunning,
		RunningFor: time.Minute*3 + time.Second*42,
		Pid:        123,
		CPU:        300,
		Memory:     1000,
		ExitCode:   0,
	}

	fields := []any{
		"int", 1,
		"str", "aboba",
		"list", []string{"a", "b", "c"},
		"ts", time.Now(),
		"status", st,
	}

	slog.Debug("debug msg")
	slog.Debug("debug msg with fields", fields...)
	slog.Info("info msg")
	slog.Info("info msg with fields", fields...)
	slog.Warn("warn msg")
	slog.Warn("warn msg with fields", fields...)
	slog.Error("error msg")
	slog.Error("error msg with fields", fields...)

	l1 := slog.Group("tag1", fields...)
	slog.LogAttrs(nil, slog.LevelDebug, "debug-logattrs msg with fields", l1) //nolint
	slog.Debug("debug msg with fields", l1)
	slog.Info("info msg with fields", l1)
	slog.Warn("warn msg with fields", l1)
	slog.Error("error msg with fields", l1)

	l2 := slog.Group("tag2", l1)
	slog.Debug("debug msg with fields2", l2)
	slog.Info("info msg with fields2", l2)
	slog.Warn("warn msg with fields2", l2)
	slog.Error("error msg with fields2", l2)

	l3 := slog.With("kupi", "doru")
	l3.Debug("debug msg")
	l3.Debug("debug msg with fields", fields...)
	l3.Info("info msg")
	l3.Info("info msg with fields", fields...)
	l3.Warn("warn msg")
	l3.Warn("warn msg with fields", fields...)
	l3.Error("error msg")
	l3.Error("error msg with fields", fields...)

	err1 := xerr.NewM("xerr with fields", xerr.Fields{
		"int":    1,
		"str":    "aboba",
		"list":   []string{"a", "b", "c"},
		"ts":     time.Now(),
		"status": st,
	})
	slog.Error("error happened", "err", err1)
	err2 := xerr.New(xerr.Fields{
		"int": 1,
		"str": "aboba",
		"ts":  time.Now(),
	})
	slog.Error("plain error happened", "err", err2)
	slog.Error("combined error happened",
		"err", xerr.NewM("aboba", xerr.Errors{err1, err2}),
	)
	slog.Error("deeply embedded error happened",
		"err", xerr.Combine(
			xerr.Combine(
				xerr.NewWM(&exec.Error{
					Name: "jjjjjjjjj",
					Err:  errors.New("executable file not found in $PATH"),
				}, "look for executable path"),
			),
		),
	)
	slog.Info("pointer to info", "data", &map[string]any{
		"StartTime": time.Now(),
		"Status":    StatusRunning,
		"Pid":       123,
		"CPU":       300,
		"Memory":    []int{1000, 2000, 3000},
		"ExitCode":  0,
		"map[string]string": map[string]string{
			"a": "b",
			"c": "d",
		},
	})
}

func main() {
	for _, l := range []struct {
		name string
		h    slog.Handler
	}{
		{
			name: "json",
			h: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				AddSource: true,
			}),
		},
		{
			name: "destruct(json)",
			h: log.NewDestructorHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				AddSource: true,
			})),
		},
		{
			name: "text",
			h: slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				AddSource: true,
			}),
		},
		{
			name: "destruct(text)",
			h: log.NewDestructorHandler(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				AddSource: true,
			})),
		},
		{
			name: "destruct(pretty)",
			h:    log.NewDestructorHandler(log.NewPrettyHandler(os.Stderr)),
		},
		{
			name: "pretty",
			h:    log.NewPrettyHandler(os.Stderr),
		},
	} {
		fmt.Println(l.name)
		run(l.h)
		fmt.Println()
		fmt.Println()
		fmt.Println()
	}
}
