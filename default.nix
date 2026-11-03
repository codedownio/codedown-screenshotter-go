{ lib
, stdenv
, buildGoModule
, static ? true
}:

buildGoModule ({
  pname = "codedown-screenshotter";
  version = "0.1.1";

  src = ./.;

  vendorHash = "sha256-m8PSIWduQNMyZpgP5qdt79NkEih+apg5iAbqkqY1q5c=";

  meta = {
    description = "";
    homepage = "";
  };
} // lib.optionalAttrs static {
  env.CGO_ENABLED = 1;

  buildInputs = [
    stdenv.cc.libc.static
  ];

  ldflags = [
    "-s" "-w"
    "-linkmode" "external"
    "-extldflags" "-static"
  ];
})
