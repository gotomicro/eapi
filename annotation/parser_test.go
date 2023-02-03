package annotation

import (
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		want    Annotation
		wantErr bool
	}{
		{
			name: "consume",
			code: "@consume application/json",
			want: &ConsumeAnnotation{ContentType: "application/json"},
		},
		{
			name: "produce",
			code: "@produce application/json",
			want: &ProduceAnnotation{ContentType: "application/json"},
		},
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
		{
			name: "security",
			code: " @security oauth2 pet:read pet:write",
			want: newSecurityAnnotation("oauth2", []string{"pet:read", "pet:write"}),
		},
		{
			name:    "security error",
			code:    "@security",
			wantErr: true,
			want:    (*SecurityAnnotation)(nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.code)
			got, err := p.Parse()
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
					return
				} else {
					t.Logf("error: %v", err)
				}
			} else if tt.wantErr {
				t.Errorf("want error. but got nil")
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
