# git-workspace

Manage multiple Git repositories as a workspace.

Run Git commands across a set of multiple repositories. Run `git workspace help` for usage instructions.

Supported commands: `checkout`, `config`, `fast-forward`, `fetch`, `run`, `status`, `update-submodules`.

## Installation

### Go

The tested Go version is 1.25. Ensure that `$GOPATH/bin` is on your `PATH`.

```sh
go install github.com/DomenicP/git-workspace@latest
```

### Nix

See [flake.nix](flake.nix) for details. Packages are provided for `nixpkgs.lib.systems.flakeExposed` systems. Additionally, an overlay is available for use with a home-manager or NixOS system configuration.

### GitHub Releases

Download the binary for your platform from the [Releases](https://github.com/DomenicP/git-workspace/releases) page,
rename it to `git-workspace` (or `git-workspace.exe` on Windows), and place it somewhere on your `PATH`.

```sh
# Example for Linux amd64
curl -L -o git-workspace https://github.com/DomenicP/git-workspace/releases/latest/download/git-workspace-linux-amd64
chmod +x git-workspace
mv git-workspace ~/.local/bin/
```
