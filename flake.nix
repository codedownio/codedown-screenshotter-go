{
  description = "CodeDown Go Screenshotter";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/release-22.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];

        pkgs = import nixpkgs { inherit system overlays; };

      in {
        packages = (rec {
          screenshotterStatic = pkgs.callPackage ./. { static = true; };
          screenshotterDynamic = pkgs.callPackage ./. { static = false; };
          default = screenshotterStatic;

          mkScreenshotter = { chromePath, static ? true }:
            let
              screenshotter = if static then screenshotterStatic else screenshotterDynamic;
            in
              with pkgs; runCommand "codedown-screenshotter-go" {
                buildInputs = [makeWrapper];
                inherit (screenshotter) meta version;
              } ''
                mkdir -p $out/bin

                makeWrapper ${screenshotter}/bin/codedown-screenshotter "$out/bin/codedown-screenshotter" \
                  --add-flags "--chrome-path \"${chromePath}\""
              '';
        });
      });
}
