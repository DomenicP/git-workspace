{
  description = "Manage a workspace of multiple Git repositories";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";

  outputs =
    { self, nixpkgs }:
    let
      inherit (nixpkgs) lib;
      systems = lib.systems.flakeExposed;
      forAllSystems = f: lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "git-workspace";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;
        };
      });

      overlays.default = final: prev: {
        git-workspace = self.packages.${prev.stdenv.hostPlatform.system}.default;
      };
    };
}
