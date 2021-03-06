package handler

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-nacelle/nacelle"
	"github.com/google/uuid"
	"github.com/ij-build/ij/config"
	"github.com/ij-build/ij/loader"
	"github.com/ij-build/ij/options"
	"github.com/ij-build/ij/subcommand"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/ij-build/ijci/agent/api"
	agentctx "github.com/ij-build/ijci/agent/context"
	"github.com/ij-build/ijci/agent/log"
	"github.com/ij-build/ijci/amqp/message"
)

type (
	Handler interface {
		Handle(message *message.BuildMessage, logger nacelle.Logger) error
	}

	handler struct {
		APIClient        apiclient.Client           `service:"api"`
		LogProcessor     *log.LogProcessor          `service:"log-processor"`
		ContextProcessor *agentctx.ContextProcessor `service:"context-processor"`
		scratchRoot      string
	}
)

func NewHandler(scratchRoot string) *handler {
	return &handler{
		scratchRoot: scratchRoot,
	}
}

func (h *handler) Handle(message *message.BuildMessage, logger nacelle.Logger) error {
	ctx, cancel := h.ContextProcessor.Create(message.BuildID)
	defer cancel()

	directory, err := ioutil.TempDir("", "build")
	if err != nil {
		return err
	}

	defer os.RemoveAll(directory)

	commit, err := h.clone(
		message.RepositoryURL,
		message.CommitBranch,
		message.CommitHash,
		directory,
		logger,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to clone repository (%s)",
			err.Error(),
		)
	}

	commitHash := commit.Hash.String()

	logger.Info(
		"Cloned repository branch %s at %s",
		message.CommitBranch,
		commitHash,
	)

	if ok, err := h.APIClient.UpdateBuild(message.BuildID, &apiclient.BuildPayload{
		CommitBranch:         &message.CommitBranch,
		CommitHash:           &commitHash,
		CommitMessage:        &commit.Message,
		CommitAuthorName:     &commit.Author.Name,
		CommitAuthorEmail:    &commit.Author.Email,
		CommitAuthoredAt:     &commit.Author.When,
		CommitCommitterName:  &commit.Committer.Name,
		CommitCommitterEmail: &commit.Committer.Email,
		CommitCommitedAt:     &commit.Committer.When,
	}); err != nil || !ok {
		if err != nil {
			return err
		}

		logger.Warning("Build is no longer active in API")
		return nil
	}

	config, err := h.loadConfig(directory)
	if err != nil {
		return fmt.Errorf(
			"failed to load build config (%s)",
			err.Error(),
		)
	}

	err = h.runDefaultPlan(
		ctx,
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

func (h *handler) clone(url, branch, hash, directory string, logger nacelle.Logger) (*object.Commit, error) {
	logger.Info(
		"Cloning repository %s into %s",
		url,
		directory,
	)

	cloneOptions := &git.CloneOptions{
		URL:               url,
		ReferenceName:     plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}

	repo, err := git.PlainClone(directory, false, cloneOptions)
	if err != nil {
		return nil, err
	}

	if hash != "" {
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, err
		}

		hash := plumbing.NewHash(hash)

		if hash.IsZero() {
			return nil, fmt.Errorf("illegal commit hash")
		}

		if err := worktree.Checkout(&git.CheckoutOptions{Hash: hash}); err != nil {
			return nil, err
		}
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	return repo.CommitObject(ref.Hash())
}

func (h *handler) loadConfig(directory string) (*config.Config, error) {
	return loader.LoadFile(filepath.Join(directory, "ij.yaml"), nil)
}

func (h *handler) runDefaultPlan(
	ctx context.Context,
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
		ProjectDir:   directory,
		ScratchRoot:  h.scratchRoot,
		DisableColor: true,
		ConfigPath:   "",
		Env:          nil,
		EnvFiles:     nil,
		Quiet:        true,
		Verbose:      true,
		FileFactory:  fileFactory,
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
		Context:             ctx,
	}

	return subcommand.NewRunCommand(
		appOptions,
		runOptions,
	)(config)
}
