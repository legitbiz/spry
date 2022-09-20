package storage

import (
	"bytes"
	"io/fs"
	"text/template"
)

func CreateFromTemplate(name string, path string, data any) (string, error) {
	t := template.New(path)
	t, err := t.ParseFiles(path)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if err := t.ExecuteTemplate(&buffer, name, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

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

func CreateTemplateFromPaths(paths ...string) (*StringTemplate, error) {
	t, err := template.ParseFiles(paths...)
	if err != nil {
		return nil, err
	}
	return &StringTemplate{templates: t}, nil
}

func CreateTemplateFromFS(files fs.FS, paths ...string) (*StringTemplate, error) {
	t, err := template.ParseFS(files, paths...)
	if err != nil {
		return nil, err
	}
	return &StringTemplate{templates: t}, nil
}
