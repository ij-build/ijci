package logs

import (
	"bytes"
	"strings"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"
)

type memoryFile struct {
	processor  *LogProcessor
	logger     nacelle.Logger
	buildID    uuid.UUID
	buildLogID uuid.UUID
	prefix     string
	buffer     *bytes.Buffer
}

func newMemoryFile(
	processor *LogProcessor,
	logger nacelle.Logger,
	buildID uuid.UUID,
	buildLogID uuid.UUID,
	prefix string,
) *memoryFile {
	return &memoryFile{
		processor:  processor,
		logger:     logger,
		buildID:    buildID,
		buildLogID: buildLogID,
		prefix:     prefix,
		buffer:     &bytes.Buffer{},
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
	return f.processor.close(f.buildID, f.buildLogID, f.buffer.String())
}
