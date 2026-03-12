{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    matt = {
      url = "github:mattrobenolt/nixpkgs";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      matt,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ matt.overlays.default ];
        };

        # Common dev tools for all shells
        devTools = with pkgs; [
          just
          gopls
          golangci-lint
          gotestsum
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
        # Default shell uses Go 1.25
        devShells.default = mkGoShell pkgs.go-bin_1_25 synctestEnv;

        # Explicit shells for each Go version
        devShells.go124 = mkGoShell pkgs.go-bin_1_24 synctestEnv;
        devShells.go125 = mkGoShell pkgs.go-bin_1_25 synctestEnv;
        devShells.go126 = mkGoShell pkgs.go-bin_1_26 { };
      }
    );
}
