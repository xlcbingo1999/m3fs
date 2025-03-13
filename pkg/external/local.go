package external

import (
	"os"

	"github.com/open3fs/m3fs/pkg/errors"
)

// LocalInterface provides interface about local run functions.
type LocalInterface interface {
	MkdirTemp(string, string) (string, error)
	MkdirAll(string) error
	RemoveAll(string) error
	WriteFile(string, []byte, os.FileMode) error
}

type localExternal struct {
	externalBase
}

func (le *localExternal) init(em *Manager) {
	le.externalBase.init(em)
	em.Local = le
}

func (le *localExternal) MkdirTemp(dir, prefix string) (string, error) {
	name, err := os.MkdirTemp(dir, prefix)
	if err != nil {
		return "", errors.Trace(err)
	}
	return name, nil
}

func (le *localExternal) MkdirAll(dir string) error {
	return errors.Trace(os.MkdirAll(dir, 0755))
}

func (le *localExternal) WriteFile(path string, data []byte, perm os.FileMode) error {
	return errors.Trace(os.WriteFile(path, data, perm))
}

func (le *localExternal) RemoveAll(dir string) error {
	return errors.Trace(os.RemoveAll(dir))
}

func init() {
	registerNewExternalFunc(func() externalInterface {
		return new(localExternal)
	})
}
