{
  description = "CodeDown Go Screenshotter";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/release-22.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    # flake-utils.lib.eachDefaultSystem (system:
    # flake-utils.lib.eachSystem [ "x86_64-linux" "x86_64-darwin" ] (system:
    flake-utils.lib.eachSystem [ "x86_64-linux" ] (system:
      let
        overlays = [];

        pkgs = import nixpkgs { inherit system overlays; };

      in rec {
        packages = (rec {
          static = pkgs.callPackage ./. { static = true; };
          dynamic = pkgs.callPackage ./. { static = false; };
          default = static;

          mkScreenshotter = { chromePath }: with pkgs; runCommand "codedown-screenshotter-go" { buildInputs = [makeWrapper]; } ''
            mkdir -p $out/bin

            makeWrapper ${default}/bin/codedown-screenshotter "$out/bin/codedown-screenshotter" \
              --add-flags chrome-path "${chromePath}"
          '';
        });
      });
}
