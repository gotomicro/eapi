package analyzer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

type Entrypoint struct {
	plugins []Plugin
}

func NewEntrypoint(plugins ...Plugin) *Entrypoint {
	return &Entrypoint{plugins: plugins}
}

func (e *Entrypoint) Run(args []string) {
	app := cli.NewApp()
	app.Name = "egen"
	app.Usage = "Tool for generating open API documentation by static-analysis"
	app.Description = `Tool for generating open API documentation by static-analysis`
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:     "plugin",
		Aliases:  []string{"p", "plug"},
		Usage:    "specify plugin name",
		Required: true,
	})
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:     "dir",
		Aliases:  []string{"d"},
		Usage:    "directory of your project which contains go.mod file",
		Required: true,
	})
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:     "output",
		Aliases:  []string{"o"},
		Usage:    "output directory of swagger.json",
		Required: true,
	})
	app.Flags = append(app.Flags, &cli.StringSliceFlag{
		Name:     "depends",
		Aliases:  []string{"dep"},
		Usage:    "depended module name",
		Required: false,
	})
	app.Action = e.run

	err := app.Run(args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func (e *Entrypoint) run(c *cli.Context) error {
	var (
		plugin     Plugin
		pluginName = c.String("plugin")
		dir        = c.String("dir")
		outputDir  = c.String("output")
		depends    = c.StringSlice("depends")
	)

	for _, p := range e.plugins {
		if p.Name() == pluginName {
			plugin = p
			break
		}
	}
	if plugin == nil {
		return fmt.Errorf("plugin %s not exists", pluginName)
	}

	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return err
	}

	a := NewAnalyzer().Plugin(plugin).Depends(depends...)
	doc := a.Process(dir).Doc()
	docContent, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(outputDir, "swagger.json"), docContent, fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
