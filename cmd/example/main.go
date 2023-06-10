package main

import (
	"github.com/rprtr258/log"
)

func main() {
	fields := log.F{
		"int":  1,
		"str":  "aboba",
		"list": []string{"a", "b", "c"},
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
}
