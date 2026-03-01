package assets

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed assets/*
var embeddedFiles embed.FS

// SafeStringAsset Returns asset data as string
// panic if not found or any err is detected
func SafeStringAsset(name string) string {
	data, err := embeddedFiles.ReadFile("assets/" + name)
	if err != nil {
		panic(err)
	}

	return string(data)
}

// TemplateAsset Executes one template by applying it to a data structure.
// panic if not found or any err is detected
func TemplateAsset(name string, data interface{}) string {
	tObj, err := template.New(name).Parse(SafeStringAsset(name))
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	err = tObj.Execute(&tpl, data)
	if err != nil {
		panic(err)
	}

	return tpl.String()
}
