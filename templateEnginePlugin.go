package spi

import (
	"io"
	"text/template"
)

type TemplateInfo struct {
	Name    string
	FuncMap template.FuncMap
	Paths   []string
}

type ITemplate interface {
	Execute(wr io.Writer, data any) error
}

type IPMAASTemplateEnginePlugin interface {
	IPMAASPlugin
	GetTemplate(templateInfo *TemplateInfo) (ITemplate, error)
}
