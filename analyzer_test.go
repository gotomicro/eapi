package eapi

import (
	"testing"
)

func TestAnalyzer_load(t *testing.T) {
	type args struct {
		pkgPath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "multi entrypoint",
			args: args{
				pkgPath: "./testdata/multi_entry",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Analyzer{}
			a.load(tt.args.pkgPath)
		})
	}
}
