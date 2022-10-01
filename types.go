package spi

import (
	"io"
)

type EntityRenderFunc func(entity any) string

type StreamingEntityRenderFunc func(w io.Writer, entity any) error
