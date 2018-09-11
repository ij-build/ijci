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
	"github.com/google/uuid"
	"gopkg.in/src-d/go-git.v4"

	"github.com/efritz/ijci/agent/api"
	"github.com/efritz/ijci/agent/logs"
	"github.com/efritz/ijci/amqp/message"
)

type (
	Handler interface {
		Handle(message *message.BuildMessage, logger nacelle.Logger) error
	}

	handler struct {
		APIClient    apiclient.Client   `service:"api"`
		LogProcessor *logs.LogProcessor `service:"log-processor"`
		scratchRoot  string
	}
)

func NewHandler(scratchRoot string) *handler {
	return &handler{
		scratchRoot: scratchRoot,
	}
}

func (h *handler) Handle(message *message.BuildMessage, logger nacelle.Logger) error {
	directory, err := ioutil.TempDir("", "build")
	if err != nil {
		return err
	}

	defer os.RemoveAll(directory)

	if err := h.clone(message.RepositoryURL, directory, logger); err != nil {
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

	err = h.runDefaultPlan(
		config,
		directory,
		message.BuildID,
	)

	if err != nil {
		return fmt.Errorf("build failed (%s)", err.Error())
	}

	logger.Info("Build complete")
	return nil
}

func (h *handler) clone(url, directory string, logger nacelle.Logger) error {
	logger.Info(
		"Cloning repository %s into %s",
		url,
		directory,
	)

	cloneOptions := &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	repo, err := git.PlainClone(directory, false, cloneOptions)
	if err != nil {
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		return err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	logger.Info(
		"Cloned repository at commit %s",
		commit.Hash,
	)

	return nil
}

func (h *handler) loadConfig(directory string) (*config.Config, error) {
	return loader.LoadFile(filepath.Join(directory, "ij.yaml"), nil)
}

func (h *handler) runDefaultPlan(
	config *config.Config,
	directory string,
	buildID uuid.UUID,
) error {
	fileFactory := func(prefix string) (io.WriteCloser, io.WriteCloser, error) {
		outFile, err := h.LogProcessor.Open(buildID, fmt.Sprintf("%s.out", prefix))
		if err != nil {
			return nil, nil, err
		}

		errFile, err := h.LogProcessor.Open(buildID, fmt.Sprintf("%s.err", prefix))
		if err != nil {
			return nil, nil, err
		}

		return outFile, errFile, nil
	}

	appOptions := &options.AppOptions{
		ProjectDir:  directory,
		ScratchRoot: h.scratchRoot,
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

	return subcommand.NewRunCommand(
		appOptions,
		runOptions,
	)(config)
}
