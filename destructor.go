package log

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

var _ slog.Handler = destructorHandler{}

type destructorHandler struct {
	h                 slog.Handler
	groups            []string
	preformattedAttrs []slog.Attr
	level             slog.Level
}

func NewDestructorHandler(h slog.Handler) destructorHandler {
	return destructorHandler{
		h:                 h,
		groups:            nil,
		preformattedAttrs: nil,
		level:             slog.LevelDebug,
	}
}

func (l destructorHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= l.level
}

func formatError(err error) slog.Value {
	return slog.StringValue(fmt.Sprintf("%T(%s)", err, err.Error()))
}

func formatLeaf(v any) slog.Value {
	switch v := v.(type) {
	case int:
		return slog.Int64Value(int64(v))
	case int8:
		return slog.Int64Value(int64(v))
	case int16:
		return slog.Int64Value(int64(v))
	case int32:
		return slog.Int64Value(int64(v))
	case int64:
		return slog.Int64Value(v)
	case uint:
		return slog.Uint64Value(uint64(v))
	case uint8:
		return slog.Uint64Value(uint64(v))
	case uint16:
		return slog.Uint64Value(uint64(v))
	case uint32:
		return slog.Uint64Value(uint64(v))
	case uint64:
		return slog.Uint64Value(v)
	case float32:
		return slog.Float64Value(float64(v))
	case float64:
		return slog.Float64Value(v)
	case bool:
		return slog.BoolValue(v)
	case time.Duration:
		return slog.DurationValue(v)
	case string:
		return slog.StringValue(v)
	case time.Time:
		return slog.TimeValue(v)
	case fmt.Stringer:
		return slog.StringValue(v.String())
	default:
		return slog.StringValue(fmt.Sprint(v))
	}
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

func formatAttr(a slog.Attr) []slog.Attr {
	// Attr's values should be resolved.
	k := a.Key
	a.Value = a.Value.Resolve()

	switch a.Value.Kind() {
	case slog.KindBool, slog.KindDuration, slog.KindString,
		slog.KindFloat64, slog.KindInt64, slog.KindUint64:
		return []slog.Attr{a}
	case slog.KindGroup:
		res := []any{}
		for _, aa := range a.Value.Group() {
			for _, aaa := range formatAttr(aa) {
				res = append(res, aaa)
			}
		}
		return []slog.Attr{slog.Group(a.Key, res...)}
	case slog.KindTime:
		return []slog.Attr{a}
	case slog.KindLogValuer:
		panic("value is unresolved after resolve")
	case slog.KindAny:
		if isLeaf(a.Value.Any()) {
			return []slog.Attr{{
				Key:   a.Key,
				Value: formatLeaf(a.Value.Any()),
			}}
		}

		v := a.Value.Any()
		reflValue := reflect.ValueOf(v)
		switch reflect.TypeOf(v).Kind() {
		case reflect.Pointer:
			if reflValue.IsZero() {
				return nil
			}

			res := formatAttr(slog.Any(k, reflValue.Elem().Interface()))
			if len(res) == 0 {
				if err, ok := v.(error); ok {
					return []slog.Attr{{
						Key:   k,
						Value: formatError(err),
					}}
				}

				return nil
			}

			return res
		case reflect.Map:
			res := []slog.Attr{}
			for i := reflValue.MapRange(); i.Next(); {
				kk, vv := i.Key(), i.Value()

				res = append(res, formatAttr(slog.Any(k+"."+fmt.Sprint(kk), vv.Interface()))...)
			}
			return res
		case reflect.Slice:
			res := []slog.Attr{}
			for i := 0; i < reflValue.Len(); i++ {
				res = append(res, formatAttr(slog.Any(k+"."+strconv.Itoa(i), reflValue.Index(i).Interface()))...)
			}
			return res
		case reflect.Struct:
			structType := reflect.TypeOf(v)

			res := []slog.Attr{}
			for i := 0; i < structType.NumField(); i++ {
				field := structType.Field(i)
				if !field.IsExported() {
					continue
				}

				res = append(res, formatAttr(slog.Any(k+"."+field.Name, reflValue.Field(i).Interface()))...)
			}
			if len(res) == 0 {
				if err, ok := v.(error); ok {
					return []slog.Attr{{Key: k, Value: formatError(err)}}
				}

				return nil
			}

			return res
		default:
			if err, ok := v.(error); ok {
				return []slog.Attr{{Key: k, Value: formatError(err)}}
			}

			return []slog.Attr{{Key: k, Value: slog.StringValue(fmt.Sprintf("%[1]T(%#[1]v)", v))}}
		}
	default:
		panic("unknown value kind")
	}
}

func (l destructorHandler) Handle(ctx context.Context, record slog.Record) error {
	fieldsSlice := slices.Clip(l.preformattedAttrs)
	record.Attrs(func(a slog.Attr) bool {
		fieldsSlice = append(fieldsSlice, formatAttr(a)...)
		return true
	})

	record2 := slog.Record{
		Time:    record.Time,
		Message: record.Message,
		Level:   record.Level,
		PC:      record.PC,
	}
	record2.AddAttrs(fieldsSlice...)

	return l.h.Handle(ctx, record2)
}

func (l destructorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := []slog.Attr{}
	for _, a := range attrs {
		for _, grp := range l.groups {
			a = slog.Group(grp, a)
		}
		newAttrs = append(newAttrs, formatAttr(a)...)
	}
	return destructorHandler{
		h:                 l.h,
		groups:            l.groups,
		preformattedAttrs: append(l.preformattedAttrs, newAttrs...),
		level:             l.level,
	}
}

func (l destructorHandler) WithGroup(name string) slog.Handler {
	return destructorHandler{
		h:                 l.h,
		groups:            append(l.groups, name),
		preformattedAttrs: slices.Clip(l.preformattedAttrs),
		level:             l.level,
	}
}
