package analyzer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	_ "github.com/gotomicro/ego-gen-api/generators/ts"
	_ "github.com/gotomicro/ego-gen-api/generators/umi"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Plugin  string
	Dir     string
	Output  string
	Depends []string
	OpenAPI OpenAPIConfig

	Generators []*GeneratorConfig
}

type OpenAPIConfig struct {
	Version string // OpenAPI version 3.0.0|3.0.3|3.1.0
}

type GeneratorConfig struct {
	Name   string
	Output string
}

type Entrypoint struct {
	k       *koanf.Koanf
	plugins []Plugin

	cfg Config
}

func NewEntrypoint(plugins ...Plugin) *Entrypoint {
	return &Entrypoint{plugins: plugins, k: koanf.New(".")}
}

const usageText = `Generate Doc:
	egogen --config config.yaml
or
	egogen --plugin gin --dir src/ --output docs/

Generate Frontend Code:
	egogen --config config.yaml gencode
or
	egogen --plugin gin --dir src/ --output docs/ gencode`

func (e *Entrypoint) Run(args []string) {
	app := cli.NewApp()
	app.Name = "egen"
	app.Usage = `Tool for generating OpenAPI documentation and Frontend Code by static-analysis`
	app.UsageText = usageText
	app.Description = `Tool for generating OpenAPI documentation and Frontend Code by static-analysis`
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "plugin",
		Aliases:     []string{"p", "plug"},
		Usage:       "specify plugin name",
		Destination: &e.cfg.Plugin,
	})
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "dir",
		Aliases:     []string{"d"},
		Usage:       "directory of your project which contains go.mod file",
		Destination: &e.cfg.Dir,
	})
	app.Flags = append(app.Flags, &cli.StringFlag{
		Name:        "output",
		Aliases:     []string{"o"},
		Usage:       "output directory of swagger.json",
		Destination: &e.cfg.Output,
	})
	app.Flags = append(app.Flags, &cli.StringSliceFlag{
		Name:    "depends",
		Aliases: []string{"dep"},
		Usage:   "depended module name",
		Action: func(context *cli.Context, depends []string) error {
			e.cfg.Depends = depends
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

		err = e.k.Unmarshal("", &e.cfg)
		if err != nil {
			return err
		}
	}

	if e.cfg.Plugin == "" {
		return fmt.Errorf("'plugin' is not set")
	}
	if e.cfg.Dir == "" {
		return fmt.Errorf("'dir' is not set")
	}
	if e.cfg.Output == "" {
		return fmt.Errorf("'output' is not set")
	}

	return nil
}

func (e *Entrypoint) loadConfig(cfg string) error {
	return e.k.Load(file.Provider(cfg), yaml.Parser())
}

func (e *Entrypoint) run(c *cli.Context) error {
	var plugin Plugin

	for _, p := range e.plugins {
		if p.Name() == e.cfg.Plugin {
			plugin = p
			break
		}
	}
	if plugin == nil {
		return fmt.Errorf("plugin %s not exists", e.cfg.Plugin)
	}

	stat, err := os.Stat(e.cfg.Dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", e.cfg.Dir)
	}

	err = os.MkdirAll(e.cfg.Output, os.ModePerm)
	if err != nil {
		return err
	}

	a := NewAnalyzer(e.k).Plugin(plugin).Depends(e.cfg.Depends...)
	doc := a.Process(e.cfg.Dir).Doc()
	if e.cfg.OpenAPI.Version != "" {
		doc.OpenAPI = e.cfg.OpenAPI.Version
	}

	// write documentation
	{
		docContent, err := json.MarshalIndent(doc, "", "    ")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(e.cfg.Output, "swagger.json"), docContent, fs.ModePerm)
		if err != nil {
			return err
		}
	}

	// execute generators
	for _, item := range e.cfg.Generators {
		err = newGeneratorExecutor(item, doc).execute()
		if err != nil {
			return err
		}
	}

	return nil
}
