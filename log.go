package log

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/rprtr258/fun"
)

var (
	_levelDebug = color.HiBlackString("DEBUG")
	_levelInfo  = "INFO"
	_levelWarn  = color.HiYellowString("WARN")
	_levelError = color.RedString("ERROR")
	_levelFatal = color.MagentaString("FATAL")
)

type F = map[string]any

type Logger struct {
	w      io.Writer
	prefix string
}

var _logger = newLogger()

func SetGlobalLogger(l Logger) {
	_logger = l
}

func newLogger() Logger { // TODO: set options
	return Logger{
		w:      os.Stderr,
		prefix: "",
	}
}

func (l Logger) Tag(tag string) Logger {
	return Logger{
		w:      l.w,
		prefix: l.prefix + "/" + tag,
	}
}

func Tag(tag string) Logger {
	return _logger.Tag(tag)
}

func (l Logger) log(level, message string, fields F) {
	prefix := fun.If(l.prefix != "", color.HiCyanString(l.prefix)+" ", "")
	if len(fields) == 0 {
		fmt.Fprintf(l.w, "[%s] %s%s\n", level, prefix, message)
		return
	}

	fieldsSlice := fun.ToSlice(fields, func(k string, v any) string {
		return color.BlueString(k) + "=" + color.GreenString("%#v", v)
	})
	sort.Strings(fieldsSlice)
	fieldsStr := strings.Join(fieldsSlice, " ")
	fmt.Fprintf(l.w, "[%s] %s%s %s\n", level, prefix, message, fieldsStr)
}

func (l Logger) Debugf(msg string, fields F) {
	l.log(_levelDebug, msg, fields)
}

func (l Logger) Debug(msg string) {
	l.Debugf(msg, nil)
}

func (l Logger) Infof(msg string, fields F) {
	l.log(_levelInfo, msg, fields)
}

func (l Logger) Info(msg string) {
	l.Infof(msg, nil)
}

func (l Logger) Warnf(msg string, fields F) {
	l.log(_levelWarn, msg, fields)
}

func (l Logger) Warn(msg string) {
	l.Warnf(msg, nil)
}

func (l Logger) Errorf(msg string, fields F) {
	l.log(_levelError, msg, fields)
}

func (l Logger) Error(msg string) {
	l.Errorf(msg, nil)
}

func (l Logger) Fatalf(msg string, fields F) {
	l.log(_levelFatal, msg, fields)
	os.Exit(1)
}

func (l Logger) Fatal(msg string) {
	l.Debugf(msg, nil)
}

func Debugf(msg string, fields F) {
	_logger.Debugf(msg, fields)
}

func Debug(msg string) {
	_logger.Debug(msg)
}

func Infof(msg string, fields F) {
	_logger.Infof(msg, fields)
}

func Info(msg string) {
	_logger.Info(msg)
}

func Warnf(msg string, fields F) {
	_logger.Warnf(msg, fields)
}

func Warn(msg string) {
	_logger.Warn(msg)
}

func Errorf(msg string, fields F) {
	_logger.Errorf(msg, fields)
}

func Error(msg string) {
	_logger.Error(msg)
}

func Fatalf(msg string, fields F) {
	_logger.Fatalf(msg, fields)
}

func Fatal(msg string) {
	_logger.Fatal(msg)
}
