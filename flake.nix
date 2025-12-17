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
    { nixpkgs, flake-utils, matt, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ matt.overlays.default ];
        };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go-bin_1_25
            gopls
            just
            gnumake
          ];

          shellHook = ''
            export PATH="$PWD/bin:$PATH"
          '';
        };
      }
    );
}
