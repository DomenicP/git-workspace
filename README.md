# git-workspace

git-workspace is a convenient way to run Git commands across multiple repositories configured as a workspace.

## Usage

```text
Usage: git workspace [COMMAND]

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
    The workspace configuration is read from git-workspace.json in the working directory.

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
```

## Installation

### Go

Ensure that `$GOPATH/bin` is on your `PATH`.

```sh
go install github.com/DomenicP/git-workspace@latest
```

### Nix

See [flake.nix](flake.nix) for details. Packages are provided for `nixpkgs.lib.systems.flakeExposed` systems. An overlay is also available for use with a home-manager or NixOS system configuration.

### GitHub Releases

Download the binary for your platform from the [Releases](https://github.com/DomenicP/git-workspace/releases) page, rename it to `git-workspace` (or `git-workspace.exe` on Windows), and place it somewhere on your `PATH`.

```sh
# Example for Linux amd64
curl -L -o git-workspace https://github.com/DomenicP/git-workspace/releases/latest/download/git-workspace-linux-amd64
chmod +x git-workspace
mv git-workspace ~/.local/bin/
```
