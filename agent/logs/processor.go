package logs

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/efritz/nacelle"
	"github.com/google/uuid"

	"github.com/efritz/ijci/agent/api"
)

type (
	LogProcessor struct {
		APIClient       apiclient.Client `service:"api"`
		Logger          nacelle.Logger   `service:"logger"`
		activeBuildLogs map[[32]byte]*fileEntry
	}

	fileEntry struct {
		file   *memoryFile
		closed chan struct{}
	}
)

var ErrUnknownBuildLog = fmt.Errorf("unknown build log")

func NewLogProcessor() *LogProcessor {
	return &LogProcessor{
		Logger:          nacelle.NewNilLogger(),
		activeBuildLogs: map[[32]byte]*fileEntry{},
	}
}

func (p *LogProcessor) Open(buildID uuid.UUID, prefix string) (io.WriteCloser, error) {
	buildLogID, err := p.APIClient.OpenBuildLog(buildID, prefix)
	if err != nil {
		return nil, err
	}

	logger := p.Logger.WithFields(nacelle.LogFields{
		"build_id":     buildID,
		"build_log_id": buildLogID,
	})

	file := newMemoryFile(
		p,
		logger,
		buildID,
		buildLogID,
		prefix,
	)

	p.activeBuildLogs[hashKey(buildID, buildLogID)] = &fileEntry{
		file:   file,
		closed: make(chan struct{}),
	}

	return file, nil
}

func (p *LogProcessor) close(buildID, buildLogID uuid.UUID, content string) error {
	// Upload content to S3 through API
	if err := p.APIClient.UploadBuildLog(buildID, buildLogID, content); err != nil {
		p.Logger.Error(
			"Failed to upload build log (%s)",
			err.Error(),
		)
	}

	// Remove record from active list
	key := hashKey(buildID, buildLogID)
	entry := p.activeBuildLogs[key]
	delete(p.activeBuildLogs, key)

	// Close listeners
	close(entry.closed)
	return nil
}

func (p *LogProcessor) GetBuildLogStream(
	ctx context.Context,
	buildID uuid.UUID,
	buildLogID uuid.UUID,
) (<-chan []byte, error) {
	entry, ok := p.activeBuildLogs[hashKey(buildID, buildLogID)]
	if !ok {
		return nil, ErrUnknownBuildLog
	}

	ch := make(chan []byte)

	go func() {
		var (
			offset = 0
			closed = false
			ticker = time.NewTicker(time.Second)
		)

		defer close(ch)
		defer ticker.Stop()

		for !closed {
			select {
			case <-ctx.Done():
				return
			case <-entry.closed:
				closed = true
			case <-ticker.C:
			}

			if bytes := entry.file.buffer.Bytes(); offset < len(bytes) {
				ch <- bytes[offset:]
				offset = len(bytes)
			}
		}
	}()

	return ch, nil
}

//
// Helpers

func hashKey(id1, id2 uuid.UUID) [32]byte {
	key := [32]byte{}
	copy(key[:], id1[:])
	copy(key[16:], id2[:])
	return key
}
