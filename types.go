package spi

import (
	"io"
)

type EntityInvocationHandler func(func(entity any)) error

type EntityRenderFunc func(entity any) (string, error)

type StreamingEntityRenderFunc func(w io.Writer, entity any) error

type EntityRendererFactory func() (EntityRenderer, error)

type EntityRenderer struct {
	RenderFunc          EntityRenderFunc
	StreamingRenderFunc StreamingEntityRenderFunc
	Styles              []string
	Scripts             []string
}
