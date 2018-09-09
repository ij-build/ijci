package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/efritz/ij/config"
	"github.com/efritz/ij/loader"
	"github.com/efritz/ij/options"
	"github.com/efritz/ij/subcommand"
	"github.com/efritz/nacelle"
	"gopkg.in/src-d/go-git.v4"

	"github.com/efritz/ijci/message"
)

type (
	Handler interface {
		Handle(message *message.BuildRequest) error
	}

	handler struct {
		Logger nacelle.Logger `service:"logger"`
	}
)

func NewHandler() *handler {
	return &handler{
		Logger: nacelle.NewNilLogger(),
	}
}

func (h *handler) Handle(message *message.BuildRequest) error {
	directory, err := ioutil.TempDir("", "build")
	if err != nil {
		return err
	}

	defer os.RemoveAll(directory)

	if err := h.clone(message.RepositoryURL, directory); err != nil {
		return fmt.Errorf(
			"failed to clone repository (%s)",
			err.Error(),
		)
	}

	config, err := h.loadConfig(directory)
	if err != nil {
		return fmt.Errorf(
			"failed to load build config (%s)",
			err.Error(),
		)
	}

	if err := h.runDefaultPlan(config); err != nil {
		return fmt.Errorf("build failed (%s)", err.Error())
	}

	h.Logger.Info("Build complete")
	return nil
}

func (h *handler) clone(url, directory string) error {
	h.Logger.Info(
		"Cloning repository %s into %s",
		url,
		directory,
	)

	cloneOptions := &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	if _, err := git.PlainClone(directory, false, cloneOptions); err != nil {
		return err
	}

	return nil
}

func (h *handler) loadConfig(directory string) (*config.Config, error) {
	return loader.LoadFile(filepath.Join(directory, "ij.yaml"), nil)
}

func (h *handler) runDefaultPlan(config *config.Config) error {
	processor := NewLogProcessor(h.Logger)

	fileFactory := func(prefix string) (io.WriteCloser, io.WriteCloser, error) {
		outFile := processor.NewFile(fmt.Sprintf("%s.out", prefix))
		errFile := processor.NewFile(fmt.Sprintf("%s.err", prefix))

		return outFile, errFile, nil
	}

	appOptions := &options.AppOptions{
		Colorize:    false,
		ConfigPath:  "",
		Env:         nil,
		EnvFiles:    nil,
		Quiet:       true,
		Verbose:     true,
		FileFactory: fileFactory,
	}

	runOptions := &options.RunOptions{
		Plans:               []string{"default"},
		CPUShares:           "",
		ForceSequential:     false,
		HealthcheckInterval: 0,
		KeepWorkspace:       false,
		Login:               true,
		Memory:              "",
		PlanTimeout:         0,
		SSHIdentities:       nil,
	}

	runner := subcommand.NewRunCommand(
		appOptions,
		runOptions,
	)

	if err := runner(config); err != nil && err != subcommand.ErrBuildFailed {
		return err
	}

	return nil
}
