package spi

import (
	"io"
)

type EntityRenderFunc func(entity any) string

type StreamingEntityRenderFunc func(w io.Writer, entity any) error

type EntityRenderer interface {
	Render(any) string
}

type StreamingEntityRenderer interface {
	Render(io.Writer, any) error
}

type EntityRendererFactory func() EntityRenderer

type StreamingEntityRendererFactory func() StreamingEntityRenderer
