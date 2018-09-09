package handler

import (
	"io"

	"github.com/efritz/nacelle"
)

type LogProcessor struct {
	logger nacelle.Logger
	files  []*memoryFile
}

func NewLogProcessor(logger nacelle.Logger) *LogProcessor {
	return &LogProcessor{
		logger: logger,
	}
}

func (p *LogProcessor) NewFile(prefix string) io.WriteCloser {
	file := newMemoryFile(p.logger, prefix)
	p.files = append(p.files, file)
	return file
}
