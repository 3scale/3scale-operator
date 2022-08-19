package assets

import (
	"bytes"
	"text/template"
)

// Important: Run "make" to regenerate code after modifying/adding/removing any asset
// adding -nometadata to not preserve size, mode, and modtime info. It should not be necessary as the content will be embedded in custom resources.
//go:generate go-bindata --prefix assets -pkg $GOPACKAGE -nometadata -o bindata.go assets/...

// SafeStringAsset Returns asset data as string
// panic if not found or any err is detected
func SafeStringAsset(name string) string {
	data, err := Asset(name)
	if err != nil {
		panic(err)
	}

	return string(data)
}

// TemplateAsset Executes one tamplate by applying it to a daata structure.
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
