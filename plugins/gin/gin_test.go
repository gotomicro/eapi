package gin

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	analyzer "github.com/gotomicro/ego-gen-api"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
)

func currentDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func TestGinPlugin(t *testing.T) {
	k := koanf.New(".")
	err := k.Load(file.Provider(filepath.Join(currentDir(), "testdata/server/config.gin.yaml")), yaml.Parser())
	if err != nil {
		t.Error(err)
	}

	path := filepath.Join(currentDir(), "testdata/server")
	plugin := NewPlugin()
	a := analyzer.NewAnalyzer(k).Plugin(plugin)
	a.Depends("github.com/gin-gonic/gin", "encoding/json")
	a.Process(path)
	doc := a.Doc()
	docContent, _ := json.MarshalIndent(doc, "", "  ")
	t.Logf("%s", string(docContent))
}
