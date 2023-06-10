package log

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

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
	fields F
}

var _logger = newLogger()

func SetGlobalLogger(l Logger) {
	_logger = l
}

func newLogger() Logger { // TODO: set options
	return Logger{
		w:      os.Stderr,
		prefix: "",
		fields: nil,
	}
}

func (l Logger) Tag(tag string) Logger {
	return Logger{
		w:      l.w,
		prefix: l.prefix + "/" + tag,
		fields: l.fields,
	}
}

func Tag(tag string) Logger {
	return _logger.Tag(tag)
}

func (l Logger) With(fields F) Logger {
	newFields := make(F, len(fields)+len(l.fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return Logger{
		w:      l.w,
		prefix: l.prefix,
		fields: newFields,
	}
}

func With(fields F) Logger {
	return _logger.With(fields)
}

func formatValue(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(v)
	case time.Time:
		return v.Format("2006.01.02 15:04:05 MST")
	case fmt.Stringer:
		return v.String()
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Slice:
			slice := reflect.ValueOf(v)

			var sb strings.Builder
			sb.WriteRune('[')
			for i := 0; i < slice.Len(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}

				sb.WriteString(formatValue(slice.Index(i).Interface()))
			}
			sb.WriteRune(']')
			return sb.String()
		case reflect.Struct:
			structType := reflect.TypeOf(v)
			structValue := reflect.ValueOf(v)

			var sb strings.Builder
			sb.WriteRune('{')
			firstFieldPrinted := false
			for i := 0; i < structType.NumField(); i++ {
				field := structType.Field(i)
				if !field.IsExported() {
					continue
				}
				if firstFieldPrinted {
					sb.WriteString(", ")
				}
				firstFieldPrinted = true
				sb.WriteString(fmt.Sprintf(
					"%s=%s",
					field.Name,
					formatValue(structValue.Field(i).Interface()),
				))
			}
			sb.WriteRune('}')
			return sb.String()
		default:
			return fmt.Sprintf("%#v", v)
		}
	}
}

func formatField(k string, v any) string {
	return color.BlueString(k) + "=" + color.GreenString("%s", formatValue(v))
}

func (l Logger) log(level, message string, fields F) {
	prefix := fun.If(l.prefix != "", color.HiCyanString(l.prefix)+" ", "")
	if len(fields) == 0 && len(l.fields) == 0 {
		fmt.Fprintf(l.w, "[%s] %s%s\n", level, prefix, message)
		return
	}

	loggerFieldsSlice := fun.ToSlice(l.fields, formatField)
	fieldsSlice := fun.ToSlice(fields, formatField)
	fieldsSlice = append(fieldsSlice, loggerFieldsSlice...)

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
	l.Fatalf(msg, nil)
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
