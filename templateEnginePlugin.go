package spi

type TemplateInfo struct {
	Name string
	Paths []string
}

type IPMAASTemplateEnginePlugin interface {
	IPMAASPlugin
	GetTemplate(templateInfo *TemplateInfo) (ITemplate, error)
}