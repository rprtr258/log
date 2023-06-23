package log

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
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

func formatTrivialField(k string, v string) string {
	return color.BlueString(k) + "=" + color.GreenString("%s", v)
}

func isLeaf(v any) bool {
	switch v := v.(type) {
	case time.Time:
		return true
	case interface {
		UnwrapFields() (string, map[string]any)
	}:
		return false
	default:
		if reflect.TypeOf(v) == nil {
			return true
		}

		switch reflect.TypeOf(v).Kind() {
		case reflect.Slice, reflect.Struct, reflect.Map, reflect.Pointer:
			return false
		default:
			return true
		}
	}
}

func formatLeaf(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(v)
	case time.Time:
		return v.Format(`2006.01.02 15:04:05 MST`)
	default:
		if reflect.TypeOf(v) == nil {
			return "<nil>"
		}

		switch reflect.TypeOf(v).Kind() {
		case reflect.String, reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128:
			return fmt.Sprint(v)
		default:
			return fmt.Sprintf("%#v", v)
		}
	}
}

// isShallow - returns true if given struct/list/map has one level of nestedness
func isShallow(v any) bool {
	switch v := v.(type) {
	case interface {
		UnwrapFields() (string, map[string]any)
	}:
		message, fields := v.UnwrapFields()
		res := message == ""
		for _, vv := range fields {
			res = res && isLeaf(vv)
		}
		return res && len(fields) != 1
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Pointer:
			pointerValue := reflect.ValueOf(v)
			if pointerValue.IsZero() { // nil pointer
				return false
			}

			return isShallow(pointerValue.Elem().Interface())
		case reflect.Map:
			mapValue := reflect.ValueOf(v)

			res := true
			for i := mapValue.MapRange(); i.Next(); {
				res = res && isLeaf(i.Value().Interface())
			}
			return res && mapValue.Len() != 1
		case reflect.Slice:
			sliceValue := reflect.ValueOf(v)

			res := true
			for i := 0; i < sliceValue.Len(); i++ {
				res = res && isLeaf(sliceValue.Index(i).Interface())
			}
			return res && sliceValue.Len() != 1
		case reflect.Struct:
			structType := reflect.TypeOf(v)
			structValue := reflect.ValueOf(v)

			res := true
			for i := 0; i < structType.NumField(); i++ {
				field := structType.Field(i)
				if !field.IsExported() {
					continue
				}

				res = res && isLeaf(structValue.Field(i).Interface())
			}
			return res && structValue.NumField() != 1
		default:
			return false
		}
	}
}

func formatShallowField(k string, v any) string {
	switch v := v.(type) {
	case map[string]any:
		var sb strings.Builder
		sb.WriteRune('{')
		itemWritten := false
		for kk, vv := range v {
			if itemWritten {
				sb.WriteString(", ")
			}

			itemWritten = true
			sb.WriteString(kk)
			sb.WriteString(": ")
			sb.WriteString(formatLeaf(vv))
		}
		sb.WriteRune('}')
		return formatTrivialField(k, sb.String())
	case interface {
		UnwrapFields() (string, map[string]any)
	}:
		_, fields := v.UnwrapFields()
		return formatShallowField(k, fields)
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Pointer:
			pointerValue := reflect.ValueOf(v)
			if pointerValue.IsZero() {
				return formatTrivialField(k, "<nil>")
			}

			return formatShallowField(k, pointerValue.Elem().Interface())
		case reflect.Slice:
			slice := reflect.ValueOf(v)
			var sb strings.Builder
			sb.WriteRune('[')
			for i := 0; i < slice.Len(); i++ {
				if i > 0 {
					sb.WriteString(", ")
				}

				sb.WriteString(formatLeaf(slice.Index(i).Interface()))
			}
			sb.WriteRune(']')
			return formatTrivialField(k, sb.String())
		case reflect.Struct:
			structType := reflect.TypeOf(v)
			structValue := reflect.ValueOf(v)

			var sb strings.Builder
			firstFieldWritten := false
			sb.WriteRune('{')
			for i := 0; i < structType.NumField(); i++ {
				field := structType.Field(i)
				if !field.IsExported() {
					continue
				}

				if firstFieldWritten {
					sb.WriteString(", ")
				}

				firstFieldWritten = true
				sb.WriteString(field.Name)
				sb.WriteString(": ")
				sb.WriteString(formatLeaf(structValue.Field(i).Interface()))
			}
			sb.WriteRune('}')
			return formatTrivialField(k, sb.String())
		default:
			panic(fmt.Sprintf("can't marshal %T as shallow type", v))
		}
	}
}

// TODO: sort by depth
func formatField(k string, v any) []string {
	if v == nil || reflect.ValueOf(v).IsZero() {
		return nil
	}

	if isLeaf(v) {
		return []string{formatTrivialField(k, formatLeaf(v))}
	}

	if isShallow(v) {
		return []string{formatShallowField(k, v)}
	}

	switch v := v.(type) {
	case string, bool, time.Time, // fmt.Stringer,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return []string{formatTrivialField(k, formatLeaf(v))}
	case interface {
		UnwrapFields() (string, map[string]any)
	}:
		res := []string{}
		message, fields := v.UnwrapFields()
		if message != "" {
			res = append(res, formatTrivialField(k, message))
		}
		for kk, vv := range fields {
			res = append(res, formatField(k+"."+kk, vv)...)
		}
		return res
	default:
		switch reflect.TypeOf(v).Kind() {
		case reflect.Pointer:
			pointerValue := reflect.ValueOf(v)
			if pointerValue.IsZero() {
				return []string{formatTrivialField(k, "<nil>")}
			}

			if ff := formatField(k, pointerValue.Elem().Interface()); len(ff) != 0 {
				return ff
			}
		case reflect.Map:
			structValue := reflect.ValueOf(v)

			res := []string{}
			for i := structValue.MapRange(); i.Next(); {
				kk, vv := i.Key(), i.Value()

				res = append(res, formatField(k+"."+fmt.Sprint(kk), vv.Interface())...)
			}
			return res
		case reflect.Slice:
			slice := reflect.ValueOf(v)

			res := []string{}
			for i := 0; i < slice.Len(); i++ {
				res = append(res, formatField(k+"."+strconv.Itoa(i), slice.Index(i).Interface())...)
			}
			return res
		case reflect.Struct:
			structType := reflect.TypeOf(v)
			structValue := reflect.ValueOf(v)

			res := []string{}
			for i := 0; i < structType.NumField(); i++ {
				field := structType.Field(i)
				if !field.IsExported() {
					continue
				}

				res = append(res, formatField(k+"."+field.Name, structValue.Field(i).Interface())...)
			}
			return res
		}

		if stringer, ok := v.(fmt.Stringer); ok {
			return []string{formatTrivialField(k, stringer.String())}
		}

		if stringer, ok := v.(error); ok {
			return []string{formatTrivialField(k, stringer.Error())}
		}

		return []string{formatTrivialField(k, fmt.Sprintf("%[1]T(%#[1]v)", v))}
	}
}

func concat[T any](a ...[]T) []T {
	resultLen := 0
	for _, v := range a {
		resultLen += len(v)
	}

	res := make([]T, 0, resultLen)
	for _, v := range a {
		res = append(res, v...)
	}
	return res
}

func (l Logger) log(level, message string, fields F) {
	prefix := fun.If(l.prefix != "", color.HiCyanString(l.prefix)+" ", "")
	if len(fields) == 0 && len(l.fields) == 0 {
		fmt.Fprintf(l.w, "[%s] %s%s\n", level, prefix, message)
		return
	}

	fieldsSlice := concat(
		concat(fun.ToSlice(fields, formatField)...),
		concat(fun.ToSlice(l.fields, formatField)...),
	)

	sort.Strings(fieldsSlice)
	fieldsStr := "\n\t" + strings.Join(fieldsSlice, "\n\t")

	fmt.Fprintf(l.w, "[%s] %s%s%s\n", level, prefix, message, fieldsStr)
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
