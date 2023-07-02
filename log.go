package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

var (
	_levelDebug = color.HiBlackString("DEBUG")
	_levelInfo  = color.HiWhiteString("INFO")
	_levelWarn  = color.HiYellowString("WARN")
	_levelError = color.RedString("ERROR")
	_levelFatal = color.MagentaString("FATAL")
)

var _ slog.Handler = Logger{}

type Logger struct {
	// TODO: add mutex
	w                 io.Writer
	group             string
	preformattedAttrs []string
	level             slog.Level
}

func New() Logger {
	return Logger{
		w:                 os.Stderr,
		group:             "",
		preformattedAttrs: nil,
		level:             slog.LevelDebug,
	}
}

func (l Logger) Enabled(_ context.Context, level slog.Level) bool {
	return level >= l.level
}

func (l Logger) Handle(_ context.Context, record slog.Record) error {
	var level string
	switch record.Level {
	case slog.LevelDebug:
		level = _levelDebug
	case slog.LevelInfo:
		level = _levelInfo
	case slog.LevelWarn:
		level = _levelWarn
	case slog.LevelError:
		level = _levelError
	default:
		level = _levelFatal
	}

	fieldsSlice := slices.Clip(l.preformattedAttrs)
	record.Attrs(func(a slog.Attr) bool {
		fieldsSlice = append(fieldsSlice, formatAttr("", a)...)
		return true
	})

	sort.Strings(fieldsSlice)
	var fieldsStr string
	if len(fieldsSlice) > 0 {
		fieldsStr = "\n\t" + strings.Join(fieldsSlice, "\n\t")
	}

	_, err := fmt.Fprintf(l.w, "[%s] %s%s\n", level, record.Message, fieldsStr)
	return err
}

func (l Logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := []string{}
	for _, a := range attrs {
		newAttrs = append(newAttrs, formatAttr(l.group, slog.Any(a.Key, a.Value))...)
	}
	return Logger{
		w:                 l.w,
		group:             l.group,
		preformattedAttrs: append(l.preformattedAttrs, newAttrs...),
		level:             l.level,
	}
}

func (l Logger) WithGroup(name string) slog.Handler {
	return Logger{
		w:                 l.w,
		group:             l.group + name + "/",
		preformattedAttrs: slices.Clip(l.preformattedAttrs),
		level:             l.level,
	}
}

func formatTrivialField(grp, k, v string) string {
	return color.HiCyanString(grp) + color.BlueString(k) + "=" + color.GreenString("%s", v)
}

func isLeaf(v any) bool {
	switch vv := v.(type) {
	case time.Time:
		return true
	default:
		if reflect.TypeOf(vv) == nil {
			return true
		}

		switch reflect.TypeOf(vv).Kind() {
		case reflect.Slice, reflect.Struct, reflect.Map, reflect.Pointer:
			return false
		default:
			return true
		}
	}
}

func formatLeaf(v any) string {
	switch v := v.(type) {
	case bool, time.Duration,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, complex64, complex128:
		return fmt.Sprint(v)
	case string:
		return fmt.Sprintf("%q", v)
	case time.Time:
		return v.Format(`"2006.01.02 15:04:05 MST"`)
	case fmt.Stringer:
		return fmt.Sprint(v)
	default:
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

// isShallow - returns true if given struct/list/map has one level of nesting
func isShallow(v any) bool {
	reflValue := reflect.ValueOf(v)
	switch reflect.TypeOf(v).Kind() {
	case reflect.Pointer:
		if reflValue.IsZero() { // nil pointer
			return false
		}

		return isShallow(reflValue.Elem().Interface())
	case reflect.Map:
		res := true
		for i := reflValue.MapRange(); i.Next(); {
			res = res && isLeaf(i.Value().Interface())
		}
		return res && reflValue.Len() != 1
	case reflect.Slice:
		res := true
		for i := 0; i < reflValue.Len(); i++ {
			res = res && isLeaf(reflValue.Index(i).Interface())
		}
		return res && reflValue.Len() != 1
	case reflect.Struct:
		structType := reflect.TypeOf(v)

		res := true
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			if !field.IsExported() {
				continue
			}

			res = res && isLeaf(reflValue.Field(i).Interface())
		}
		return res && reflValue.NumField() != 1
	default:
		return false
	}
}

func formatShallow(grp string, v slog.Attr) string {
	k := v.Key
	reflValue := reflect.ValueOf(v.Value.Any())
	switch reflect.TypeOf(v.Value.Any()).Kind() {
	case reflect.Pointer:
		if reflValue.IsZero() {
			return ""
		}

		return formatShallow(grp, slog.Any(k, reflValue.Elem().Interface()))
	case reflect.Map:
		if reflValue.Len() == 0 {
			return ""
		}

		var sb strings.Builder
		sb.WriteRune('{')
		itemWritten := false
		for i := reflValue.MapRange(); i.Next(); {
			kk, vv := i.Key(), i.Value()

			if itemWritten {
				sb.WriteString(", ")
			}

			itemWritten = true
			sb.WriteString(fmt.Sprint(kk.Interface()))
			sb.WriteString(": ")
			sb.WriteString(formatLeaf(vv.Interface()))
		}
		sb.WriteRune('}')
		return formatTrivialField(grp, k, sb.String())
	case reflect.Slice:
		if reflValue.Len() == 0 {
			return ""
		}

		var sb strings.Builder
		sb.WriteRune('[')
		for i := 0; i < reflValue.Len(); i++ {
			if i > 0 {
				sb.WriteString(", ")
			}

			sb.WriteString(formatLeaf(reflValue.Index(i).Interface()))
		}
		sb.WriteRune(']')
		return formatTrivialField(grp, k, sb.String())
	case reflect.Struct:
		structType := reflect.TypeOf(v.Value.Any())

		if structType.NumField() == 0 {
			return ""
		}

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
			sb.WriteString(formatLeaf(reflValue.Field(i).Interface()))
		}
		sb.WriteRune('}')
		return formatTrivialField(grp, k, sb.String())
	default:
		panic(fmt.Sprintf("can't marshal %T as shallow type", v.Value.Any()))
	}
}

func formatAttr(grp string, a slog.Attr) []string {
	t, isTime := a.Value.Any().(time.Time)
	if isTime && t.IsZero() || // If r.Time is the zero time, ignore the time.
		a.Equal(slog.Attr{}) ||
		a.Value.Any() == nil { // If an Attr's key and value are both the zero value, ignore the Attr.
		return nil
	}

	// Attr's values should be resolved.
	k := a.Key
	a.Value = a.Value.Resolve()

	switch a.Value.Kind() {
	case slog.KindBool:
		b := a.Value.Bool()
		if !b {
			return nil
		}
		return []string{formatTrivialField(grp, k, formatLeaf(b))}
	case slog.KindDuration:
		d := a.Value.Duration()
		if d == 0 {
			return nil
		}
		return []string{formatTrivialField(grp, k, formatLeaf(d))}
	case slog.KindString:
		s := a.Value.String()
		if s == "" {
			return nil
		}
		return []string{formatTrivialField(grp, k, formatLeaf(s))}
	case slog.KindFloat64:
		f := a.Value.Float64()
		if f == 0 {
			return nil
		}
		return []string{formatTrivialField(grp, k, formatLeaf(f))}
	case slog.KindInt64:
		i := a.Value.Int64()
		if i == 0 {
			return nil
		}
		return []string{formatTrivialField(grp, k, formatLeaf(i))}
	case slog.KindUint64:
		u := a.Value.Uint64()
		if u == 0 {
			return nil
		}
		return []string{formatTrivialField(grp, k, formatLeaf(u))}
	case slog.KindGroup:
		// If a group has no Attrs (even if it has a non-empty key), ignore it.
		if len(a.Value.Group()) == 0 {
			return nil
		}

		groupPrefix := grp
		if a.Key != "" {
			groupPrefix = k + "/" + groupPrefix
		}

		res := []string{}
		for _, aa := range a.Value.Group() {
			res = append(res, formatAttr(groupPrefix, aa)...)
		}
		return res
	case slog.KindLogValuer:
		panic("value is unresolved after resolve")
	case slog.KindTime, slog.KindAny:
		if isLeaf(a.Value.Any()) {
			return []string{formatTrivialField(grp, k, formatLeaf(a.Value.Any()))}
		}

		if isShallow(a.Value.Any()) {
			res := formatShallow(grp, a)
			if res == "" {
				return nil
			}

			return []string{res}
		}

		v := a.Value.Any()
		reflValue := reflect.ValueOf(v)
		switch reflect.TypeOf(v).Kind() {
		case reflect.Pointer:
			if reflValue.IsZero() {
				return nil
			}

			return formatAttr(grp, slog.Any(k, reflValue.Elem().Interface()))
		case reflect.Map:
			if reflValue.Len() == 0 {
				return nil
			}

			res := []string{}
			for i := reflValue.MapRange(); i.Next(); {
				kk, vv := i.Key(), i.Value()

				res = append(res, formatAttr(grp, slog.Any(k+"."+fmt.Sprint(kk), vv.Interface()))...)
			}
			return res
		case reflect.Slice:
			if reflValue.Len() == 0 {
				return nil
			}

			res := []string{}
			for i := 0; i < reflValue.Len(); i++ {
				res = append(res, formatAttr(grp, slog.Any(k+"."+strconv.Itoa(i), reflValue.Index(i).Interface()))...)
			}
			return res
		case reflect.Struct:
			structType := reflect.TypeOf(v)

			res := []string{}
			for i := 0; i < structType.NumField(); i++ {
				field := structType.Field(i)
				if !field.IsExported() {
					continue
				}

				res = append(res, formatAttr(grp, slog.Any(k+"."+field.Name, reflValue.Field(i).Interface()))...)
			}
			if len(res) == 0 {
				return nil
			}

			return res
		default:
			if stringer, ok := v.(error); ok {
				return []string{formatTrivialField(grp, k, stringer.Error())}
			}

			return []string{formatTrivialField(grp, k, fmt.Sprintf("%[1]T(%#[1]v)", v))}
		}
	default:
		panic("unknown value kind")
	}
}
