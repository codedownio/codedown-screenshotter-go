{ lib
, stdenv
, buildGoModule
, static ? true
}:

buildGoModule ({
  pname = "codedown-screenshotter";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-zecMVEVsYEZJ+lxFF12+GbhX+URtOGktOnC2xMLr+yo=";

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
