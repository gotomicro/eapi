package jsvm

import (
	"testing"

	"github.com/gotomicro/eapi/spec"
)

func TestVM_Print(t *testing.T) {
	type args struct {
		code string
		doc  *spec.T
	}
	tests := []struct {
		name    string
		args    args
		wantRes string
		wantErr bool
	}{
		{
			name: "print test",
			args: args{
				code: `
const { docBuilders: {join, indent, hardline}, printDocToString } = require("eapi");
module.exports.print = function (doc) {
	console.log(JSON.stringify(doc));
	return printDocToString([
		"function hello() {",
		indent([
			hardline,
			"return 0;",
		]),
		hardline,
		"}"
	], {tabWidth: 2}).formatted;
};
`,
				doc: &spec.T{
					OpenAPI: "3.1.0",
				},
			},
			wantRes: `function hello() {
  return 0;
}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			gotRes, err := v.Run(tt.args.code, tt.args.doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRes != tt.wantRes {
				t.Errorf("Run() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
