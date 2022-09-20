package storage

import (
	"bytes"
	"io/fs"
	"text/template"
)

type StringTemplate struct {
	templates *template.Template
}

func (st StringTemplate) Execute(name string, data any) (string, error) {
	var buffer bytes.Buffer
	if err := st.templates.ExecuteTemplate(&buffer, name, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func CreateTemplateFromFS(files fs.FS, paths ...string) (*StringTemplate, error) {
	t, err := template.ParseFS(files, paths...)
	if err != nil {
		return nil, err
	}
	return &StringTemplate{templates: t}, nil
}
