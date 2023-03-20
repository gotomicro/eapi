package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	analyzer "github.com/gotomicro/eapi"
	"github.com/gotomicro/eapi/plugins/echo"
	"github.com/gotomicro/eapi/plugins/gin"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/stretchr/testify/assert"
)

var plugins = map[string]analyzer.Plugin{}

func init() {
	var pluginList = []analyzer.Plugin{gin.NewPlugin(), echo.NewPlugin()}
	for _, plugin := range pluginList {
		plugins[plugin.Name()] = plugin
	}
}

func TestAnalyzer(t *testing.T) {
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
		{
			name: "echo",
			args: args{
				pkgPath: "./testdata/echo",
			},
		},
		{
			name: "gin",
			args: args{
				pkgPath: "./testdata/gin",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := koanf.New(".")
			err := k.Load(file.Provider(filepath.Join(tt.args.pkgPath, "eapi.yaml")), yaml.Parser())
			assert.NoError(t, err)

			var config analyzer.Config
			err = k.Unmarshal("", &config)
			assert.NoError(t, err)

			plugin, ok := plugins[config.Plugin]
			assert.Truef(t, ok, "plugin %s not exists", config.Plugin)

			a := analyzer.NewAnalyzer(k).Plugin(plugin).Depends(config.Depends...).Process(tt.args.pkgPath)
			expectedDoc, err := os.ReadFile(filepath.Join(tt.args.pkgPath, "docs/openapi.json"))
			assert.NoError(t, err)

			doc := a.Doc().Specialize()
			config.OpenAPI.ApplyToDoc(doc)
			actualDoc, err := json.MarshalIndent(doc, "", "  ")
			assert.NoError(t, err)
			assert.JSONEq(t, string(expectedDoc), string(actualDoc))
		})
	}
}
