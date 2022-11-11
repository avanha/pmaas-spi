package spi

import (
	"io"
	"text/template"
)

type TemplateInfo struct {
	Name    string
	FuncMap template.FuncMap
	Paths   []string
	Scripts []string
	Styles  []string
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
