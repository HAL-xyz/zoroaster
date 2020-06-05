package action

import (
	"bytes"
	"html/template"
	"strings"
)

func renderTemplateWithData(templateText string, data interface{}) (string, error) {

	funcMap := template.FuncMap{
		"upperCase": strings.ToUpper,
	}

	tmpl := template.New("").Funcs(funcMap)
	t, err := tmpl.Parse(templateText)

	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	err = t.Execute(&output, data)

	return output.String(), err
}
