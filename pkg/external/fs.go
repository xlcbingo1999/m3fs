package external

import (
	"os"

	"github.com/open3fs/m3fs/pkg/errors"
)

// FSInterface provides interface about local fs, this is not implemented for remote runner.
type FSInterface interface {
	MkdirTemp(string, string) (string, error)
	MkdirAll(string) error
	RemoveAll(string) error
	WriteFile(string, []byte, os.FileMode) error
}

type fsExternal struct {
	externalBase

	returnUnimplemented bool
}

func (fe *fsExternal) init(em *Manager) {
	fe.externalBase.init(em)
	em.FS = fe
	if _, ok := em.Runner.(*RemoteRunner); ok {
		fe.returnUnimplemented = true
	}
}

func (fe *fsExternal) MkdirTemp(dir, prefix string) (string, error) {
	if fe.returnUnimplemented {
		return "", errors.New("unimplemented")
	}
	name, err := os.MkdirTemp(dir, prefix)
	if err != nil {
		return "", errors.Trace(err)
	}
	return name, nil
}

func (fe *fsExternal) MkdirAll(dir string) error {
	if fe.returnUnimplemented {
		return errors.New("unimplemented")
	}
	return errors.Trace(os.MkdirAll(dir, 0755))
}

func (fe *fsExternal) WriteFile(path string, data []byte, perm os.FileMode) error {
	if fe.returnUnimplemented {
		return errors.New("unimplemented")
	}
	return errors.Trace(os.WriteFile(path, data, perm))
}

func (fe *fsExternal) RemoveAll(dir string) error {
	if fe.returnUnimplemented {
		return errors.New("unimplemented")
	}
	return errors.Trace(os.RemoveAll(dir))
}

func init() {
	registerNewExternalFunc(func() externalInterface {
		return new(fsExternal)
	})
}
