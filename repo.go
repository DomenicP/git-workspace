package main

import (
	"errors"
	"fmt"
	"os"
)

// Repo is the representation of a Git repository and provides methods for running Git commands.
type Repo struct {
	RepoConfig
	runner Runner
}

// PathExist returns true if the repository path exists, and false otherwise.
func (r Repo) PathExist() bool {
	_, err := os.Stat(r.Path)
	return !errors.Is(err, os.ErrNotExist)
}

// Checkout performs a Git checkout of the configured ref.
func (r Repo) Checkout() error {
	if r.Ref == "" {
		return fmt.Errorf("no ref specified for %s", r.Path)
	}
	return r.runner.Run("git", "checkout", r.Ref)
}

// Clone performs a Git clone of the repository if the repository path does not exist.
func (r Repo) Clone() error {
	if !r.PathExist() {
		return r.runner.Run("git", "clone", r.Remote, r.Path)
	}
	return nil
}

// Config applies the workspace-specific Git config to the repository.
func (r Repo) Config(cfg map[string]string) error {
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

// FastForward runs a ff-only pull on the repo.
func (r Repo) FastForward() error {
	return r.runner.Run("git", "pull", "--ff-only")
}

// Fetch runs a Git fetch from origin.
func (r Repo) Fetch() error {
	return r.runner.Run("git", "fetch")
}

// Push causes the repository to push the current branch to origin.
func (r Repo) Push() error {
	return r.runner.Run("git", "push")
}

// Run performs an arbitrary shell command in the repository (not a Git command).
func (r Repo) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("run requires a command to execute")
	}
	return r.runner.Run(args[0], args[1:]...)
}

// Status gets the working tree status for the repository.
func (r Repo) Status() error {
	return r.runner.Run("git", "status")
}

// UpdateSubmodules initializes and updates submodules recursively.
func (r Repo) UpdateSubmodules() error {
	return r.runner.Run("git", "submodule", "update", "--init", "--recursive")
}
