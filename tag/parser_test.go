package tag

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name    string
		args    args
		wantRes map[string]string
	}{
		{
			name: "name",
			args: args{
				tag: `json:"key"`,
			},
			wantRes: map[string]string{
				"json": "key",
			},
		},
		{
			name: "escaped value",
			args: args{
				tag: `k:"key\"xxx"`,
			},
			wantRes: map[string]string{
				"k": "key\\\"xxx",
			},
		},
		{
			name: "multiple KV",
			args: args{
				tag: `a:"1" b:"2"`,
			},
			wantRes: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := Parse(tt.args.tag); !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("Parse() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
