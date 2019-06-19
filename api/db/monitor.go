package db

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-nacelle/nacelle"
	"github.com/go-nacelle/pgutil"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type (
	Monitor interface {
		nacelle.Process
		Subscribe() (string, <-chan Event)
		Unsubscribe(id string)
	}

	Event struct {
		Table  string                 `json:"table"`
		Action string                 `json:"action"`
		Data   map[string]interface{} `json:"data"`
	}

	monitor struct {
		Logger      nacelle.Logger `service:"logger"`
		connInfo    string
		subscribers map[string]chan<- Event
		mutex       sync.RWMutex
		halt        chan struct{}
		once        sync.Once
	}
)

func NewMonitor() Monitor {
	return &monitor{
		subscribers: map[string]chan<- Event{},
		halt:        make(chan struct{}),
	}
}

func (m *monitor) Init(config nacelle.Config) error {
	dbConfig := &pgutil.Config{}
	if err := config.Load(dbConfig); err != nil {
		return err
	}

	m.connInfo = dbConfig.DatabaseURL
	return nil
}

func (m *monitor) Start() error {
	// TODO - add other things to config
	listener := pq.NewListener(
		m.connInfo,
		10*time.Second, // TODO - configure
		time.Minute,    // TODO - configure
		func(ev pq.ListenerEventType, err error) {
			if err != nil {
				fmt.Println(err.Error())
			}
		},
	)

	// TODO - add other things to config
	if err := listener.Listen("events"); err != nil {
		panic(err)
	}

	m.Logger.Info("Listening for database events")

	for {
		select {
		case <-m.halt:
			return nil

		case notification := <-listener.Notify:
			event := Event{}
			if err := json.Unmarshal([]byte(notification.Extra), &event); err != nil {
				return err
			}

			m.mutex.RLock()

			for _, l := range m.subscribers {
				select {
				case l <- event:
				default:
				}
			}

			m.mutex.RUnlock()

		case <-time.After(60 * time.Second):
			m.Logger.Info("Received no events for 60 seconds, checking connection")

			if err := listener.Ping(); err != nil {
				return err
			}
		}
	}
}

func (m *monitor) Stop() error {
	m.once.Do(func() {
		close(m.halt)
	})

	return nil
}

func (m *monitor) Subscribe() (string, <-chan Event) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	id := uuid.New().String()
	ch := make(chan Event)
	m.subscribers[id] = ch
	return id, ch
}

func (m *monitor) Unsubscribe(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.subscribers, id)
}
