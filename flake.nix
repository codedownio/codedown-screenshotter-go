{
  description = "CodeDown Go Screenshotter";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/release-25.05";
  # Only for the dev shell's browser, and the same one codedown runs this against, so the tests
  # exercise the browser production uses. Kept separate so what gets built and shipped does not
  # move just to get a browser.
  inputs.nixpkgs-browser.url = "github:codedownio/nixpkgs/release-24.11-codedown-apr17-2025";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, nixpkgs-browser, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];

        pkgs = import nixpkgs { inherit system overlays; };

        browser = (import nixpkgs-browser { inherit system; }).chromiumHeadlessShell;

      in {
        # So that `go test` finds a browser without being told where one is: the tests look for
        # headless-shell on PATH, which is what the wrapper below runs on Linux too.
        devShells.default = pkgs.mkShell {
          packages = [pkgs.go browser];
        };

        # The browser above is built from source and is in no cache a hosted runner can reach,
        # so CI takes the toolchain from here and its browser from the runner image.
        devShells.ci = pkgs.mkShell {
          packages = [pkgs.go];
        };

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
                passthru = {
                  unwrapped = screenshotter;
                };
              } ''
                mkdir -p $out/bin

                makeWrapper ${screenshotter}/bin/codedown-screenshotter "$out/bin/codedown-screenshotter" \
                  --add-flags "--chrome-path \"${chromePath}\""
              '';
        });
      });
}
