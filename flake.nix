{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    matt = {
      url = "github:mattrobenolt/nixpkgs";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      perSystem =
        { pkgs, system, ... }:
        let
          # Common dev tools for all shells
          devTools = with pkgs; [
            just
            gopls
            golangci-lint
            gotestsum
            zizmor
            pinact
          ];

          mkGoShell =
            goPackage: attrs:
            pkgs.mkShell (
              {
                packages = [ goPackage ] ++ devTools;
              }
              // attrs
            );
          synctestEnv = {
            GOEXPERIMENT = "synctest";
          };
        in
        {
          # mkGoShell above uses the overlayed pkgs from this
          # _module.args override, not the default nixpkgs.
          _module.args.pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ inputs.matt.overlays.default ];
          };

          # Default shell uses Go 1.25
          devShells.default = mkGoShell pkgs.go-bin_1_25 synctestEnv;

          # Explicit shells for each Go version
          devShells.go124 = mkGoShell pkgs.go-bin_1_24 synctestEnv;
          devShells.go125 = mkGoShell pkgs.go-bin_1_25 synctestEnv;
          devShells.go126 = mkGoShell pkgs.go-bin_1_26 { };
        };
    };
}
