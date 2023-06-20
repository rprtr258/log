package log

import (
	"testing"

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
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, isShallow(test.v))
		})
	}
}
