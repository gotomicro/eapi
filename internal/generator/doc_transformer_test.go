package generator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_transformValue(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "hello",
			args: args{
				v: "hello",
			},
			want: "hello",
		},
		{
			name: "map + slice",
			args: args{
				v: map[string]interface{}{
					"hello": map[string]interface{}{
						"world": []string{"ok"},
					},
				},
			},
			want: map[interface{}]interface{}{
				"hello": map[interface{}]interface{}{
					"world": []interface{}{"ok"},
				},
			},
		},
		{
			name: "struct",
			args: args{
				v: &struct {
					A struct {
						B []string
					}
				}{
					A: struct {
						B []string
					}{
						B: []string{"hello", "world"},
					},
				},
			},
			want: map[string]interface{}{
				"A": map[string]interface{}{
					"B": []interface{}{"hello", "world"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.args.v)
			got := transformValue(v)
			assert.Equal(t, tt.want, got, "transformValue() = %v, want %v", got, tt.want)
		})
	}
}
