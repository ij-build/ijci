package handler

import (
	"bytes"
	"strings"

	"github.com/efritz/nacelle"
)

type memoryFile struct {
	logger           nacelle.Logger
	buildLogUploader BuildLogUploader
	prefix           string
	buffer           *bytes.Buffer
}

func newMemoryFile(logger nacelle.Logger, buildLogUploader BuildLogUploader, prefix string) *memoryFile {
	return &memoryFile{
		logger:           logger,
		buildLogUploader: buildLogUploader,
		prefix:           prefix,
		buffer:           &bytes.Buffer{},
	}
}

func (f *memoryFile) Write(p []byte) (int, error) {
	f.logger.Debug(
		"Build log %s: %s",
		f.prefix,
		strings.TrimSpace(string(p)),
	)

	return f.buffer.Write(p)
}

func (f *memoryFile) Close() error {
	return f.buildLogUploader(f.prefix, f.buffer.String())
}
