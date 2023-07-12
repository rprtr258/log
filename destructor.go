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
	group             string
	preformattedAttrs []slog.Attr
	level             slog.Level
}

func NewDestructorHandler(h slog.Handler) destructorHandler {
	return destructorHandler{
		h:                 h,
		group:             "",
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

// TODO: slog.GroupValue
func formatTrivialField(grp, k string, v slog.Value) slog.Attr {
	return slog.Attr{
		Key:   grp + k,
		Value: v,
	}
}

func formatAttr(grp string, a slog.Attr) []slog.Attr {
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
		return []slog.Attr{formatTrivialField(grp, k, formatLeaf(b))}
	case slog.KindDuration:
		d := a.Value.Duration()
		if d == 0 {
			return nil
		}
		return []slog.Attr{formatTrivialField(grp, k, formatLeaf(d))}
	case slog.KindString:
		s := a.Value.String()
		if s == "" {
			return nil
		}
		return []slog.Attr{formatTrivialField(grp, k, formatLeaf(s))}
	case slog.KindFloat64:
		f := a.Value.Float64()
		if f == 0 {
			return nil
		}
		return []slog.Attr{formatTrivialField(grp, k, formatLeaf(f))}
	case slog.KindInt64:
		i := a.Value.Int64()
		if i == 0 {
			return nil
		}
		return []slog.Attr{formatTrivialField(grp, k, formatLeaf(i))}
	case slog.KindUint64:
		u := a.Value.Uint64()
		if u == 0 {
			return nil
		}
		return []slog.Attr{formatTrivialField(grp, k, formatLeaf(u))}
	case slog.KindGroup:
		// If a group has no Attrs (even if it has a non-empty key), ignore it.
		if len(a.Value.Group()) == 0 {
			return nil
		}

		groupPrefix := grp
		if a.Key != "" {
			groupPrefix = k + "/" + groupPrefix
		}

		res := []slog.Attr{}
		for _, aa := range a.Value.Group() {
			res = append(res, formatAttr(groupPrefix, aa)...)
		}
		return res
	case slog.KindLogValuer:
		panic("value is unresolved after resolve")
	case slog.KindTime, slog.KindAny:
		if isLeaf(a.Value.Any()) {
			return []slog.Attr{formatTrivialField(grp, k, formatLeaf(a.Value.Any()))}
		}

		v := a.Value.Any()
		reflValue := reflect.ValueOf(v)
		switch reflect.TypeOf(v).Kind() {
		case reflect.Pointer:
			if reflValue.IsZero() {
				return nil
			}

			res := formatAttr(grp, slog.Any(k, reflValue.Elem().Interface()))
			if len(res) == 0 {
				if err, ok := v.(error); ok {
					return []slog.Attr{formatTrivialField(grp, k, formatError(err))}
				}

				return nil
			}

			return res
		case reflect.Map:
			if reflValue.Len() == 0 {
				return nil
			}

			res := []slog.Attr{}
			for i := reflValue.MapRange(); i.Next(); {
				kk, vv := i.Key(), i.Value()

				res = append(res, formatAttr(grp, slog.Any(k+"."+fmt.Sprint(kk), vv.Interface()))...)
			}
			return res
		case reflect.Slice:
			if reflValue.Len() == 0 {
				return nil
			}

			res := []slog.Attr{}
			for i := 0; i < reflValue.Len(); i++ {
				res = append(res, formatAttr(grp, slog.Any(k+"."+strconv.Itoa(i), reflValue.Index(i).Interface()))...)
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

				res = append(res, formatAttr(grp, slog.Any(k+"."+field.Name, reflValue.Field(i).Interface()))...)
			}
			if len(res) == 0 {
				if err, ok := v.(error); ok {
					return []slog.Attr{formatTrivialField(grp, k, formatError(err))}
				}

				return nil
			}

			return res
		default:
			if err, ok := v.(error); ok {
				return []slog.Attr{formatTrivialField(grp, k, formatError(err))}
			}

			return []slog.Attr{formatTrivialField(grp, k, slog.StringValue(fmt.Sprintf("%[1]T(%#[1]v)", v)))}
		}
	default:
		panic("unknown value kind")
	}
}

func (l destructorHandler) Handle(ctx context.Context, record slog.Record) error {
	fieldsSlice := slices.Clip(l.preformattedAttrs)
	record.Attrs(func(a slog.Attr) bool {
		fieldsSlice = append(fieldsSlice, formatAttr("", a)...)
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
		newAttrs = append(newAttrs, formatAttr(l.group, slog.Any(a.Key, a.Value))...)
	}
	return destructorHandler{
		h:                 l.h,
		group:             l.group,
		preformattedAttrs: append(l.preformattedAttrs, newAttrs...),
		level:             l.level,
	}
}

func (l destructorHandler) WithGroup(name string) slog.Handler {
	return destructorHandler{
		h:                 l.h,
		group:             l.group + name + "/",
		preformattedAttrs: slices.Clip(l.preformattedAttrs),
		level:             l.level,
	}
}
