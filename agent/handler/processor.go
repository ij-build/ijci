package handler

import (
	"io"

	"github.com/efritz/nacelle"
)

type (
	LogProcessor struct {
		logger           nacelle.Logger
		buildLogUploader BuildLogUploader
		files            []*memoryFile
	}

	BuildLogUploader func(name, content string) error
)

func NewLogProcessor(logger nacelle.Logger, buildLogUploader BuildLogUploader) *LogProcessor {
	return &LogProcessor{
		logger:           logger,
		buildLogUploader: buildLogUploader,
	}
}

func (p *LogProcessor) NewFile(prefix string) io.WriteCloser {
	file := newMemoryFile(p.logger, p.buildLogUploader, prefix)
	p.files = append(p.files, file)
	return file
}
