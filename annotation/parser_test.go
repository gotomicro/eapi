package annotation

import (
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name string
		code string
		want Annotation
	}{
		{
			name: "required",
			code: " @required",
			want: newSimpleAnnotation(Required),
		},
		{
			name: "REQUIRED",
			code: " @REQUIRED  ",
			want: newSimpleAnnotation(Required),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.code)
			if got := p.Parse(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
