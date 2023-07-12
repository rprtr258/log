package log

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

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

var _ slog.Handler = prettyHandler{}

type prettyHandler struct {
	// TODO: add mutex
	w                 io.Writer
	group             string
	preformattedAttrs []string
	level             slog.Level
}

func NewPrettyHandler(w io.Writer) prettyHandler {
	return prettyHandler{
		w:                 w,
		group:             "",
		preformattedAttrs: nil,
		level:             slog.LevelDebug,
	}
}

func (l prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= l.level
}

func prettyFormatAttr(grp string, a slog.Attr) []string {
	var valueStr string
	switch v := a.Value; v.Kind() {
	case slog.KindAny:
		if stringer, ok := v.Any().(fmt.Stringer); ok {
			valueStr = stringer.String()
		} else {
			valueStr = fmt.Sprint(v)
		}
	case slog.KindBool:
		valueStr = fmt.Sprint(v.Bool())
	case slog.KindDuration:
		valueStr = fmt.Sprint(v.Duration())
	case slog.KindFloat64:
		valueStr = fmt.Sprint(v.Float64())
	case slog.KindInt64:
		valueStr = fmt.Sprint(v.Int64())
	case slog.KindUint64:
		valueStr = fmt.Sprint(v.Uint64())
	case slog.KindString:
		valueStr = fmt.Sprintf("%q", v.String())
	case slog.KindTime:
		valueStr = fmt.Sprint(v.Time().Format(`"2006.01.02 15:04:05 MST"`))
	case slog.KindGroup:
		res := make([]string, 0, len(v.Group()))
		for _, a1 := range v.Group() {
			res = append(res, prettyFormatAttr(grp+a.Key+"/", a1)...)
		}
		return res
	case slog.KindLogValuer:
		return prettyFormatAttr(grp, slog.Any(a.Key, a.Value.Resolve()))
	default:
		panic(fmt.Sprintf("unsupported kind: %#v", v.Kind()))
	}

	return []string{color.HiCyanString(grp) +
		color.BlueString(a.Key) +
		"=" +
		color.GreenString("%s", valueStr)}
}

func (l prettyHandler) Handle(_ context.Context, record slog.Record) error {
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
		fieldsSlice = append(fieldsSlice, prettyFormatAttr("", a)...)
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

func (l prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := []string{}
	for _, a := range attrs {
		for _, attr := range formatAttr(l.group, slog.Any(a.Key, a.Value)) {
			newAttrs = append(newAttrs, prettyFormatAttr("", attr)...)
		}
	}
	return prettyHandler{
		w:                 l.w,
		group:             l.group,
		preformattedAttrs: append(l.preformattedAttrs, newAttrs...),
		level:             l.level,
	}
}

func (l prettyHandler) WithGroup(name string) slog.Handler {
	return prettyHandler{
		w:                 l.w,
		group:             l.group + name + "/",
		preformattedAttrs: slices.Clip(l.preformattedAttrs),
		level:             l.level,
	}
}
