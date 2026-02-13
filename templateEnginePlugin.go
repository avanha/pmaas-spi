package spi

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"text/template"
)

type TemplateInfo struct {
	Name     string
	FuncMap  template.FuncMap
	Paths    []string
	Scripts  []string
	Styles   []string
	SourceFS fs.FS
}

type ITemplate interface {
	Execute(wr io.Writer, data any) error
}

type CompiledTemplate struct {
	Instance ITemplate
	Styles   []string
	Scripts  []string
}

type IPMAASTemplateEnginePlugin interface {
	IPMAASPlugin
	GetTemplate(templateInfo *TemplateInfo) (CompiledTemplate, error)
}

func TemplateBasedRendererFactory(
	container IPMAASContainer,
	templateInfo *TemplateInfo,
	typeCheckFn func(any) bool,
	typeName string) (EntityRenderer, error) {

	// Load the template
	compiledTemplate, err := container.GetTemplate(templateInfo)

	if err != nil {
		return EntityRenderer{},
			fmt.Errorf("unable to load %s template: %w", templateInfo.Name, err)
	}

	// Declare a function that casts the entity to the expected type and evaluates it via the template loaded above
	streamingRenderer := func(w io.Writer, entity any) error {
		if ok := typeCheckFn(entity); !ok {
			return fmt.Errorf("item is not an instance of %s", typeName)
		}

		if err := compiledTemplate.Instance.Execute(w, entity); err != nil {
			return fmt.Errorf("unable to execute %s template: %w", templateInfo.Name, err)
		}

		return nil
	}

	renderer := func(entity any) (string, error) {
		var buffer bytes.Buffer
		err := streamingRenderer(&buffer, entity)

		if err != nil {
			return "", err
		}

		return buffer.String(), nil
	}

	return EntityRenderer{
		RenderFunc:          renderer,
		StreamingRenderFunc: streamingRenderer,
		Styles:              compiledTemplate.Styles,
		Scripts:             compiledTemplate.Scripts,
	}, nil
}
