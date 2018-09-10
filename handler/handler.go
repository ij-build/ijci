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

	"github.com/efritz/ijci/api-client"
	"github.com/efritz/ijci/message"
)

type (
	Handler interface {
		Handle(message *message.BuildMessage, logger nacelle.Logger) error
	}

	handler struct {
		APIClient api.Client `service:"api"`
	}
)

func NewHandler() *handler {
	return &handler{}
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

	buildLogUploader := func(name, content string) error {
		err := h.APIClient.UploadBuildLog(uuid.Must(uuid.Parse(message.BuildID)), name, content)
		if err != nil {
			logger.Error("Failed to upload build log (%s)", err.Error())
		}

		return err
	}

	if err := h.runDefaultPlan(config, logger, buildLogUploader); err != nil {
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

	if _, err := git.PlainClone(directory, false, cloneOptions); err != nil {
		return err
	}

	return nil
}

func (h *handler) loadConfig(directory string) (*config.Config, error) {
	return loader.LoadFile(filepath.Join(directory, "ij.yaml"), nil)
}

func (h *handler) runDefaultPlan(
	config *config.Config,
	logger nacelle.Logger,
	buildLogUploader BuildLogUploader,
) error {
	processor := NewLogProcessor(logger, buildLogUploader)

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

	return subcommand.NewRunCommand(
		appOptions,
		runOptions,
	)(config)
}
