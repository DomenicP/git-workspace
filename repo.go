package main

import (
	"errors"
	"fmt"
	"os"
)

type Repo struct {
	RepoConfig
	runner Runner
}

func NewRepo(cfg RepoConfig, runner Runner) *Repo {
	return &Repo{
		RepoConfig: cfg,
		runner:     runner,
	}
}

func (r *Repo) PathExist() bool {
	_, err := os.Stat(r.Path)
	return !errors.Is(err, os.ErrNotExist)
}

func (r *Repo) Checkout() error {
	if r.Ref == "" {
		return fmt.Errorf("no ref specified for %s", r.Path)
	}
	return r.runner.Run("git", "checkout", r.Ref)
}

func (r *Repo) Clone() error {
	if !r.PathExist() {
		return r.runner.Run("git", "clone", r.Remote, r.Path)
	}
	return nil
}

func (r *Repo) Config(cfg map[string]string) error {
	if len(cfg) == 0 {
		return nil
	}
	for k, v := range cfg {
		if err := r.runner.Run("git", "config", k, v); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) FastForward() error {
	return r.runner.Run("git", "pull", "--ff-only")
}

func (r *Repo) Fetch() error {
	return r.runner.Run("git", "fetch", "-pt")
}

func (r *Repo) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("run requires a command to execute")
	}
	return r.runner.Run(args[0], args[1:]...)
}

func (r *Repo) Status() error {
	return r.runner.Run("git", "status")
}

func (r *Repo) UpdateSubmodules() error {
	return r.runner.Run("git", "submodule", "update", "--init", "--recursive")
}
