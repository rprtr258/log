package log

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/kr/pretty"
	"github.com/rprtr258/xerr"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

type ProcID uint64

func TestFormatLeaf(t *testing.T) {
	for name, test := range map[string]struct {
		v    any
		want string
	}{
		"ProcID": {
			v:    ProcID(123),
			want: "123",
		},
		"str": {
			v:    "aboba",
			want: `"aboba"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, formatLeaf(test.v))
		})
	}
}

func TestIsLeaf(t *testing.T) {
	for name, test := range map[string]struct {
		v    any
		want bool
	}{
		"map[string]any": {
			v: map[string]any{
				"key": "value",
			},
			want: false,
		},
		"pointer to map[string]any": {
			v: &map[string]any{
				"key":  "value",
				"ints": []int{1, 2, 3},
			},
			want: false,
		},
		"ProcID": {
			v:    ProcID(123),
			want: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, isLeaf(test.v))
		})
	}
}

func TestIsShallow(t *testing.T) {
	for name, test := range map[string]struct {
		v    any
		want bool
	}{
		"slice of map[string]any": {
			v: []map[string]any{
				{
					"key": "value",
				},
			},
			want: false,
		},
		"map[string]any": {
			v: map[string]any{
				"key": "value",
				"two": "three",
			},
			want: true,
		},
		"depply nested xerr": {
			v: xerr.Combine(
				xerr.Combine(
					xerr.NewWM(&exec.Error{
						Name: "jjjjjjjjj",
						Err:  errors.New("executable file not found in $PATH"),
					}, "look for executable path"),
				),
			),
			want: false,
		},
		"pointer to map[string]any": {
			v: &map[string]any{
				"key":  "value",
				"ints": []int{1, 2, 3},
			},
			want: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, isShallow(test.v), "v is %s", pretty.Sprint(test.v))
		})
	}
}

func TestFormatAttr(t *testing.T) {
	for name, test := range map[string]struct {
		attr slog.Attr
		want []string
	}{
		"str": {
			attr: slog.String("a", "b"),
			want: []string{`a="b"`},
		},
		"nil pointer to struct": {
			attr: slog.Any("a", struct {
				p *int
			}{
				p: nil,
			}),
			want: nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, formatAttr("", test.attr))
		})
	}
}
