{
  description = "CodeDown Go Screenshotter";

  inputs.gitignore = {
    url = "github:hercules-ci/gitignore.nix";
    inputs.nixpkgs.follows = "nixpkgs";
  };
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/release-22.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, gitignore, nixpkgs, flake-utils }:
    # flake-utils.lib.eachDefaultSystem (system:
    # flake-utils.lib.eachSystem [ "x86_64-linux" "x86_64-darwin" ] (system:
    flake-utils.lib.eachSystem [ "x86_64-linux" ] (system:
      let
        overlays = [];

        pkgs = import nixpkgs { inherit system overlays; };

      in rec {
        packages = (rec {
          default = pkgs.callPackage ./. {};
          dynamic = pkgs.callPackage ./. { static = false; };
        });
      });
}
