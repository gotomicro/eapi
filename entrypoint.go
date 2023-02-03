package eapi

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/gotomicro/eapi/generators/ts"
	_ "github.com/gotomicro/eapi/generators/umi"
	"github.com/gotomicro/eapi/spec"
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
	OpenAPI         string           `yaml:"openapi"` // OpenAPI version 3.0.0|3.0.3|3.1.0
	Info            *spec.Info       `yaml:"info"`    // Required
	SecuritySchemes *SecuritySchemes `yaml:"securitySchemes"`
}

type SecuritySchemes map[string]*spec.SecurityScheme

func (c OpenAPIConfig) applyToDoc(doc *spec.T) {
	if c.OpenAPI != "" {
		doc.OpenAPI = c.OpenAPI
	}
	if c.Info != nil {
		if c.Info.Version != "" {
			doc.Info.Version = c.Info.Version
		}
		if c.Info.Title != "" {
			doc.Info.Title = c.Info.Title
		}
		if c.Info.Description != "" {
			doc.Info.Description = c.Info.Description
		}
		if c.Info.TermsOfService != "" {
			doc.Info.TermsOfService = c.Info.TermsOfService
		}
	}
	if c.SecuritySchemes != nil {
		doc.Components.SecuritySchemes = make(map[string]*spec.SecuritySchemeRef)
		for name, scheme := range *c.SecuritySchemes {
			doc.Components.SecuritySchemes[name] = &spec.SecuritySchemeRef{Value: scheme}
		}
	}
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
	return &Entrypoint{
		k:       koanf.New("."),
		plugins: plugins,
		cfg: Config{
			Plugin: "gin",
			Dir:    ".",
			Output: "docs",
		},
	}
}

const usageText = `Generate Doc:
	eapi --config config.yaml
or
	eapi --plugin gin --dir src/ --output docs/

Generate Frontend Code:
	eapi --config config.yaml gencode
or
	eapi --plugin gin --dir src/ --output docs/ gencode`

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
		Usage:       "output directory of openapi.json",
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
	if cfg == "" {
		fileInfo, err := os.Stat("eapi.yaml")
		if err == nil && !fileInfo.IsDir() {
			cfg = "eapi.yaml"
		}
	}
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
		e.cfg.Dir = "."
	}
	if e.cfg.Output == "" {
		e.cfg.Output = "docs"
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
	e.cfg.OpenAPI.applyToDoc(doc)

	// write documentation
	{
		docContent, err := json.MarshalIndent(doc, "", "    ")
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(e.cfg.Output, "openapi.json"), docContent, fs.ModePerm)
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
