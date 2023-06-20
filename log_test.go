package log

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/kr/pretty"
	"github.com/rprtr258/xerr"
	"github.com/stretchr/testify/assert"
)

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
			},
			want: true,
		},
		"new": {
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
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, isShallow(test.v), "v is %s", pretty.Sprint(test.v))
		})
	}
}
