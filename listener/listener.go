package listener

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	"github.com/efritz/nacelle"
	"gopkg.in/src-d/go-git.v4"

	"github.com/efritz/ijci/amqp"
)

type Listener struct {
	Logger   nacelle.Logger `service:"logger"`
	Consumer *amqp.Consumer `service:"amqp-consumer"`
}

func NewListener() *Listener {
	return &Listener{
		Logger: nacelle.NewNilLogger(),
	}
}

func (l *Listener) Init(config nacelle.Config) error {
	return nil
}

func (l *Listener) Start() error {
	for delivery := range l.Consumer.Deliveries() {
		if err := l.handle(string(delivery.Body)); err != nil {
			l.Logger.Error(
				"failed to handle message (%s)",
				err.Error(),
			)
		}

		delivery.Ack(false)
	}

	l.Logger.Info("No longer consuming")
	return nil
}

func (l *Listener) Stop() error {
	return l.Consumer.Shutdown()
}

func (l *Listener) handle(url string) error {
	directory, err := ioutil.TempDir("", "build")
	if err != nil {
		return err
	}

	defer os.RemoveAll(directory)

	l.Logger.Info(
		"Cloning repository %s into %s",
		url,
		directory,
	)

	cloneOptions := &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	if _, err := git.PlainClone(directory, false, cloneOptions); err != nil {
		return fmt.Errorf(
			"failed to clone repository (%s)",
			err.Error(),
		)
	}

	command := exec.CommandContext(context.Background(), "/ij", "--no-color")
	command.Dir = directory

	outReader, err := command.StdoutPipe()
	if err != nil {
		return err
	}

	errReader, err := command.StderrPipe()
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() { defer wg.Done(); processOutput(outReader) }()
	go func() { defer wg.Done(); processOutput(errReader) }()

	if err := command.Run(); err != nil {
		return err
	}

	wg.Wait()

	l.Logger.Warning("Build complete")
	return nil
}

func processOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		fmt.Printf("> %#v\n", scanner.Text())
	}
}
