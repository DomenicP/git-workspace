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

var GitConfig = map[string]string{
	"commit.gpgsign": "true",
}

type Repo struct {
	Path       string `json:"path"`
	Ref        string `json:"ref"`
	Submodules bool   `json:"submodules"`
}

type Config struct {
	Repos []Repo `json:"repos"`
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

	for _, repo := range cfg.Repos {
		if _, err := os.Stat(repo.Path); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: path %s does not exist\n", repo.Path)
			continue
		}

		fmt.Printf("===== Entering: %s =====\n", repo.Path)
		check(os.Chdir(repo.Path), "chdir failed")

		switch cmd {
		case "checkout":
			err = repo.Checkout()
		case "config":
			err = repo.Config()
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

// Repo Commands

func (r *Repo) Checkout() error {
	if r.Ref == "" {
		return fmt.Errorf("no ref specified for %s", r.Path)
	}
	return runGit("checkout", r.Ref)
}

func (r *Repo) Config() error {
	for k, v := range GitConfig {
		if err := runGit("config", k, v); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) FastForward() error {
	return runGit("pull", "--ff-only")
}

func (r *Repo) Fetch() error {
	return runGit("fetch", "-pt")
}

func (r *Repo) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("run requires a command to execute")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (r *Repo) Status() error {
	return runGit("status")
}

func (r *Repo) UpdateSubmodules() error {
	if r.Submodules {
		return runGit("submodule", "update", "--init", "--recursive")
	}
	return nil
}

// Lib

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
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
