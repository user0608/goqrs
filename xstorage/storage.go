package xstorage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

type StorageService interface {
	Find(location string) (io.ReadSeeker, error)
	Save(location string, r io.Reader) error
	Delete(location string) error
}

type template struct {
	basePath string
}

func NewTemplateStorage(basePath string) (StorageService, error) {
	basePath = path.Clean(basePath)
	if err := os.MkdirAll(basePath, os.ModeDir|0760); err != nil {
		return nil, err
	}
	t := &template{basePath: basePath}
	return t, nil
}

func (s *template) path(fp string) string {
	return s.basePath + "/" + fp
}
func (s *template) Find(location string) (io.ReadSeeker, error) {
	contentfile, err := os.ReadFile(s.path(location))
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(contentfile), nil
}
func (s *template) Save(location string, r io.Reader) error {
	file, err := os.OpenFile(s.path(location), os.O_CREATE|os.O_WRONLY, 0760)
	if err != nil {
		var patherr *fs.PathError
		if errors.As(err, &patherr) {
			fmt.Println("OK")
			if err := os.MkdirAll(s.path(path.Dir(location)), os.ModeDir|0760); err != nil {
				return err
			}
			file, err = os.OpenFile(s.path(location), os.O_CREATE|os.O_WRONLY, 0760)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer file.Close()
	if _, err := io.Copy(file, r); err != nil {
		return err
	}
	return nil
}
func (s *template) Delete(location string) error {
	if err := os.Remove(s.path(location)); err != nil {
		return err
	}
	return nil
}
