package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"charm.land/lipgloss/v2"
	"golang.org/x/term"
)

const (
	ConfigFile = "git-workspace.json"
	Version    = "0.3.0"
)

const Usage = `Usage: git workspace [COMMAND]

Run commands across multiple Git repositories grouped in a workspace.

Commands:
    co, checkout            Check out the configured ref in each repo
    clone                   Ensure all repositories are cloned
    config                  Apply workspace git config settings to each repo
    ff, fast-forward        Attempt to fast-forward each repo
    fetch                   Fetch each repo from origin
    help                    Print usage information and exit
    push                    Attempt to push each repo
    run <cmd>               Run an arbitrary shell command in each repo
    st, status              Show the working tree status for each repo
    sup, update-submodules  Update submodules for each repo
    version                 Print version information and then exit

Configuration:
    The workspace configuration is read from {{.ConfigFile}} in the working directory.

    Example:
    {
      "git_config": {  // optional key/value pairs applied via 'git config'
        "user.email": "you@example.com"
      },
      "repos": [
        {
          "path": "./path/to/repo",  // relative or absolute path to repo directory
          "remote": "...",           // remote URL for clone
          "ref": "main",             // branch/tag/commit for checkout
        },
        ...
      ]
    }
`

// Runner is the interface for running external commands. It allows dependency injection for
// mocking exec.Command.
//
// Run executes the specified command with the provided arguments. It returns any errors from
// the underlying implementation.
type Runner interface {
	Run(name string, args ...string) error
}

// RepoConfig is the configuration of a single Git repository in the workspace.
type RepoConfig struct {
	Path   string `json:"path"`
	Remote string `json:"remote"`
	Ref    string `json:"ref"`
}

// Config is the overall workspace configuration.
type Config struct {
	GitConfig map[string]string `json:"git_config,omitempty"`
	Repos     []RepoConfig      `json:"repos"`
}

func printUsage(rc int) {
	tmpl := template.Must(template.New("usage").Parse(Usage))
	if err := tmpl.Execute(os.Stdout, map[string]any{"ConfigFile": ConfigFile}); err != nil {
		panic("could not print usage: " + err.Error())
	}
	os.Exit(rc)
}

func printVersion() {
	fmt.Println(Version)
	os.Exit(0)
}

func main() {
	// Check command line arguments.
	if len(os.Args) < 2 {
		printUsage(0)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	// Early check for help and version before attempting to load the config.
	if cmd == "help" {
		printUsage(0)
	}
	if cmd == "version" {
		printVersion()
	}

	// Load the configuration.
	cfg, err := loadConfig(ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: could not load %s: %v\n", ConfigFile, err)
		os.Exit(1)
	}

	// Iterate over configured repositories.
	runner := CommandRunner{}
	for _, r := range cfg.Repos {
		repo := Repo{r, runner}

		// Special check for clone command: if the repo is not yet cloned, the directory will
		// not exist.
		if cmd == "clone" {
			fmt.Printf("cloning %s to %s\n", repo.Remote, repo.Path)
			err = repo.Clone()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: clone failed for %s: %v\n", repo.Path, err)
			}
			continue
		}

		// For all other commands, the path should exist
		if !repo.PathExist() {
			fmt.Fprintf(os.Stderr, "warning: path %s does not exist\n", repo.Path)
			continue
		}

		printHeader(repo.Path)
		if err := os.Chdir(repo.Path); err != nil {
			panic("could not change directory: " + err.Error())
		}

		switch cmd {
		case "checkout", "co":
			err = repo.Checkout()
		case "config":
			err = repo.Config(cfg.GitConfig)
		case "fast-forward", "ff":
			err = repo.FastForward()
		case "fetch":
			err = repo.Fetch()
		case "push":
			err = repo.Push()
		case "run":
			err = repo.Run(args)
		case "status", "st":
			err = repo.Status()
		case "update-submodules", "sup":
			err = repo.UpdateSubmodules()
		default:
			printUsage(1)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: command %s failed for %s: %v\n", cmd, repo.Path, err)
		}

		fmt.Println()
	}
}

// CommandRunner is the exec.Command implementation of Runner.
type CommandRunner struct{}

// Run executes the named command with the provided arguments using exec.Command.
//
// It sets the commands Stdin, Stdout, and Stderr to the os streams. Any errors from running the
// command are returned.
func (c CommandRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func loadConfig(configFile string) (*Config, error) {
	// Open the file handle.
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Decode the JSON.
	dec := json.NewDecoder(f)
	cfg := new(Config)
	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("could not decode %s: %w", configFile, err)
	}

	// Convert relative paths to absolute.
	cwd, err := os.Getwd()
	if err != nil {
		panic("could not change directory: " + err.Error())
	}
	for i := range cfg.Repos {
		if !filepath.IsAbs(cfg.Repos[i].Path) {
			cfg.Repos[i].Path = filepath.Clean(filepath.Join(cwd, cfg.Repos[i].Path))
		}
	}

	return cfg, nil
}

func printHeader(msg string) {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	style := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Border(lipgloss.NormalBorder()).
		Foreground(lipgloss.Green).
		Bold(true)
	fmt.Println(style.Render(msg))
}
