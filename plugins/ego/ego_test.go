package ego

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	analyzer "github.com/gotomicro/ego-gen-api"
)

func currentDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func TestEgoPlugin(t *testing.T) {
	path := filepath.Join(currentDir(), "testdata/server")
	a := analyzer.NewAnalyzer().Plugin(&Plugin{})
	a.Process(path)
	doc := a.Doc()
	docContent, _ := json.MarshalIndent(doc, "", "  ")
	t.Logf("%s", string(docContent))
}
