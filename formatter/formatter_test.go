package formatter

import "testing"

func TestFormat(t *testing.T) {
	options := &Options{
		IndentWidth: 4,
	}

	type args struct {
		doc Doc
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case1",
			args: args{
				doc: Group(
					Content("export function hello() {"),
					LineBreak(),
					Indent(Join(
						LineBreak(),
						NewDocContent("console.log(\"hello, world\");"),
						NewDocContent("return;"),
					)),
					LineBreak(),
					Content("}"),
				),
			},
			want: `export function hello() {
    console.log("hello, world");
    return;
}`,
		},
		{
			name: "nested-indent",
			args: args{
				doc: Group(
					Content("export function hello() {"),
					Indent(Group(
						LineBreak(),
						Join(
							LineBreak(),
							Content("{"),
							Indent(NewDocContent("console.log(\"hello, world\");")),
							Content("}"),
							NewDocContent("return;"),
						),
					)),
					LineBreak(),
					Content("}"),
				),
			},
			want: `export function hello() {
    {
        console.log("hello, world");
    }
    return;
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Format(tt.args.doc, options); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}
