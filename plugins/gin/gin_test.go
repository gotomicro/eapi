package gin

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	analyzer "github.com/gotomicro/ego-gen-api"
	"github.com/spf13/viper"
)

func currentDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func TestGinPlugin(t *testing.T) {
	viper.SetConfigFile(filepath.Join(currentDir(), "testdata/server/config.gin.yaml"))
	err := viper.ReadInConfig()
	if err != nil {
		t.Error(err)
	}

	path := filepath.Join(currentDir(), "testdata/server")
	plugin := NewPlugin()
	a := analyzer.NewAnalyzer().Plugin(plugin)
	a.Depends("github.com/gin-gonic/gin", "encoding/json")
	a.Process(path)
	doc := a.Doc()
	docContent, _ := json.MarshalIndent(doc, "", "  ")
	t.Logf("%s", string(docContent))
}
