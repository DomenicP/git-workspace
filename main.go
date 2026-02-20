package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Runner interface {
	Run(name string, args ...string) error
}

type RepoConfig struct {
	Path       string `json:"path"`
	Ref        string `json:"ref"`
	Submodules bool   `json:"submodules"`
}

type Config struct {
	GitConfig map[string]string `json:"git_config,omitempty"`
	Repos     []RepoConfig      `json:"repos"`
}

func printUsage(rc int) {
	cmds := []string{
		"checkout",
		"config",
		"fast-forward",
		"fetch",
		"run",
		"status",
		"update-submodules",
	}
	fmt.Printf("Usage: git workspace {%s}\n", strings.Join(cmds, "|"))
	os.Exit(rc)
}

func main() {
	if len(os.Args) < 2 {
		printUsage(0)
	}

	cfg, err := loadConfig()
	check(err, "could not load config")

	cmd := os.Args[1]
	args := os.Args[2:]

	for _, r := range cfg.Repos {
		repo := NewRepo(&r, CommandRunner{})

		if repo.PathExist() {
			fmt.Fprintf(os.Stderr, "warning: path %s does not exist\n", repo.Path)
			continue
		}

		fmt.Printf("===== Entering: %s =====\n", repo.Path)
		check(os.Chdir(repo.Path), "chdir failed")

		switch cmd {
		case "checkout":
			err = repo.Checkout()
		case "config":
			err = repo.Config(cfg.GitConfig)
		case "fast-forward", "ff":
			err = repo.FastForward()
		case "fetch":
			err = repo.Fetch()
		case "run":
			err = repo.Run(args)
		case "status":
			err = repo.Status()
		case "update-submodules", "sup":
			err = repo.UpdateSubmodules()
		default:
			printUsage(1)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "command %s failed for %s: %v\n", cmd, repo.Path, err)
		}

		fmt.Printf("===== Exiting: %s =====\n\n", repo.Path)
	}
}

type CommandRunner struct{}

func (c CommandRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func check(err error, msg string) {
	if err != nil {
		panic(msg + ": " + err.Error())
	}
}

func loadConfig() (*Config, error) {
	f, err := os.Open("repos.json")
	if err != nil {
		return nil, fmt.Errorf("could not open repos.json: %w", err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	cfg := new(Config)
	if err := dec.Decode(cfg); err != nil && err != io.EOF {
		return nil, fmt.Errorf("could not decode repos.json: %w", err)
	}

	// Normalize paths: make them relative to cwd if they're not absolute
	cwd, _ := os.Getwd()
	for i := range cfg.Repos {
		if !filepath.IsAbs(cfg.Repos[i].Path) {
			cfg.Repos[i].Path = filepath.Clean(filepath.Join(cwd, cfg.Repos[i].Path))
		}
	}

	return cfg, nil
}
