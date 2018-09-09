package handler

import (
	"strings"

	"github.com/efritz/nacelle"
)

type memoryFile struct {
	logger nacelle.Logger
	prefix string
}

func newMemoryFile(logger nacelle.Logger, prefix string) *memoryFile {
	return &memoryFile{
		logger: logger,
		prefix: prefix,
	}
}

func (f *memoryFile) Write(p []byte) (int, error) {
	f.logger.Debug("Build log %s: %s", f.prefix, strings.TrimSpace(string(p)))
	return len(p), nil
}

func (f *memoryFile) Close() error {
	return nil
}
