package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const ConfigFile = "git-workspace.json"

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
	fmt.Printf(`Usage: git workspace [COMMAND]

Run commands across a configured set of Git repositories.

Commands:
    checkout                Check out the configured ref in each repo
    config                  Apply workspace git config settings to each repo
    ff, fast-forward        Pull the latest changes using fast-forward only
    fetch                   Fetch from origin (including tags and pruning)
    help                    Print usage information and exit
    run <cmd>               Run an arbitrary command in each repo
    status                  Show the working tree status of each repo
    sup, update-submodules  Initialize and update submodules recursively

Details:
    Workspace configuration is read from %s in the current working directory.

    Example:
    {
      "git_config": {  // optional key/value pairs applied via 'git config'
        "user.email": "you@example.com"
      },
      "repos": [
        {
          "path": "./path/to/repo",  // Relative or absolute path to repo
          "ref": "main",             // branch/tag/commit for checkout
          "submodules": true         // Enable or disable submodules support
        },
        ...
      ]
    }
`, ConfigFile)
	os.Exit(rc)
}

func main() {
	if len(os.Args) < 2 {
		printUsage(0)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	if cmd == "help" {
		printUsage(0)
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not load %s: %v\n", ConfigFile, err)
		os.Exit(1)
	}

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
	f, err := os.Open(ConfigFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	cfg := new(Config)
	if err := dec.Decode(cfg); err != nil && err != io.EOF {
		return nil, fmt.Errorf("could not decode %s: %w", ConfigFile, err)
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
