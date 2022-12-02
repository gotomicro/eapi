package eapi

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotomicro/eapi/generators"
	"github.com/gotomicro/eapi/spec"
)

type generatorExecutor struct {
	cfg *GeneratorConfig
	doc *spec.T
}

func newGeneratorExecutor(cfg *GeneratorConfig, doc *spec.T) *generatorExecutor {
	return &generatorExecutor{cfg: cfg, doc: doc}
}

func (r *generatorExecutor) execute() (err error) {
	item := r.cfg
	tpl, ok := generators.Generators[item.Name]
	if !ok {
		return fmt.Errorf("generator '%s' not exists", item.Name)
	}
	for _, t := range tpl.Items {
		err = r.executeItem(t)
		if err != nil {
			return
		}
	}

	return
}

func (r *generatorExecutor) executeItem(t *generators.Item) error {
	outputFile := filepath.Join(r.cfg.Output, t.FileName)
	dir := filepath.Dir(outputFile)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(outputFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	content := t.Print(r.doc)
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
