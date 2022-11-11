package analyzer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

type Entrypoint struct {
	plugins []Plugin

	plugin  string
	dir     string
	output  string
	depends []string
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
		Name:        "plugin",
		Aliases:     []string{"p", "plug"},
		Usage:       "specify plugin name",
		Destination: &e.plugin,
	})
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "dir",
		Aliases:     []string{"d"},
		Usage:       "directory of your project which contains go.mod file",
		Destination: &e.dir,
	})
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "output",
		Aliases:     []string{"o"},
		Usage:       "output directory of swagger.json",
		Destination: &e.output,
	})
	app.Flags = append(app.Flags, &cli.StringSliceFlag{
		Name:    "depends",
		Aliases: []string{"dep"},
		Usage:   "depended module name",
		Action: func(context *cli.Context, depends []string) error {
			e.depends = depends
			return nil
		},
	})
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:     "config",
		Aliases:  []string{"c"},
		Usage:    "configuration file",
		Required: false,
	})

	app.Before = e.before
	app.Action = e.run

	err := app.Run(args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func (e *Entrypoint) before(c *cli.Context) error {
	cfg := c.String("config")
	if cfg != "" {
		err := e.loadConfig(cfg)
		if err != nil {
			return err
		}

		if e.plugin == "" {
			e.plugin = viper.GetString("plugin")
		}
		if e.dir == "" {
			e.dir = viper.GetString("dir")
		}
		if e.output == "" {
			e.output = viper.GetString("output")
		}
		if len(e.depends) == 0 {
			e.depends = viper.GetStringSlice("depends")
		}
	}

	if e.plugin == "" {
		return fmt.Errorf("'plugin' is not set")
	}
	if e.dir == "" {
		return fmt.Errorf("'dir' is not set")
	}
	if e.output == "" {
		return fmt.Errorf("'output' is not set")
	}

	return nil
}

func (e *Entrypoint) loadConfig(cfg string) error {
	viper.SetConfigFile(cfg)
	return viper.ReadInConfig()
}

func (e *Entrypoint) run(c *cli.Context) error {
	var plugin Plugin

	for _, p := range e.plugins {
		if p.Name() == e.plugin {
			plugin = p
			break
		}
	}
	if plugin == nil {
		return fmt.Errorf("plugin %s not exists", e.plugin)
	}

	stat, err := os.Stat(e.dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", e.dir)
	}

	err = os.MkdirAll(e.output, os.ModePerm)
	if err != nil {
		return err
	}

	a := NewAnalyzer().Plugin(plugin).Depends(e.depends...)
	doc := a.Process(e.dir).Doc()
	docContent, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(e.output, "swagger.json"), docContent, fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
