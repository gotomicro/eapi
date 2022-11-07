package analyzer

import (
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

type ModFile struct {
	*modfile.File
}

func LoadModFileFrom(packagePath string) (mod *ModFile, err error) {
	file, err := ReadGoMod(packagePath)
	if err != nil {
		return
	}

	mod = &ModFile{File: file}
	return
}

func (m *ModFile) GetDep(moduleName string) *module.Version {
	// check if is replaced
	for _, replace := range m.Replace {
		if replace.Old.Path == moduleName {
			return &replace.New
		}
	}

	for _, require := range m.Require {
		if require.Mod.Path == moduleName {
			return &require.Mod
		}
	}

	return nil
}
