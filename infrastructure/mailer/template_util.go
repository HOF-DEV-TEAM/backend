package mailer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

type Template struct{}

var functions = template.FuncMap{}

// TODO: add template caching..
func (t *Template) Create(buf io.Writer, fileName string, data any) error {
	dir, err := os.Getwd()
	page := fmt.Sprintf("%s/%s", filepath.Join(dir, "templates"), fileName)

	ts, err := template.New(fileName).Funcs(functions).ParseFiles(page)

	if err != nil {
		return err
	}

	ts, err = ts.ParseGlob("./templates/*.layout.tmpl")

	if err != nil {
		return err
	}

	if err = ts.Execute(buf, data); err != nil {
		return err
	}
	return nil
}
