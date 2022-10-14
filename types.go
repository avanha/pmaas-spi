package spi

import (
	"io"
)

type EntityRenderFunc func(entity any) (string, error)

type StreamingEntityRenderFunc func(w io.Writer, entity any) error

type EntityRendererFactory func() (EntityRenderFunc, error)

type StreamingEntityRendererFactory func() (StreamingEntityRenderFunc, error)
