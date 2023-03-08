{ lib
, stdenv
, buildGoModule
, static ? true
}:

buildGoModule ({
  pname = "codedown-screenshotter";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-oePl/GpP31Vv2Yj4VYwxYlNOo5fIVVMrW9SVsOowbjA=";

  meta = with lib; {
    description = "";
    homepage = "";
  };
} // lib.optionalAttrs static {
  CGO_ENABLED = 1;

  buildInputs = [ stdenv.cc.libc.static ];

  ldflags = [
    "-s" "-w"
    "-linkmode" "external"
    "-extldflags" "-static"
  ];
})
