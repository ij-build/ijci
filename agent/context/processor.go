package context

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type (
	ContextProcessor struct {
		contexts map[uuid.UUID]*contextPair
		mutex    sync.Mutex
	}

	contextPair struct {
		ctx    context.Context
		cancel func()
	}
)

var ErrUnknownContext = fmt.Errorf("unknown build context")

func NewContextProcessor() *ContextProcessor {
	return &ContextProcessor{
		contexts: map[uuid.UUID]*contextPair{},
	}
}

func (p *ContextProcessor) Create(buildID uuid.UUID) (context.Context, func()) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	p.contexts[buildID] = &contextPair{ctx, cancel}
	return ctx, cancel
}

func (p *ContextProcessor) Cancel(buildID uuid.UUID) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	pair, ok := p.contexts[buildID]
	if !ok {
		return ErrUnknownContext
	}

	pair.cancel()
	delete(p.contexts, buildID)
	return nil
}
