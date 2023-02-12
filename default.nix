{ lib
, buildGoModule
}:

buildGoModule rec {
  pname = "codedown-screenshotter";
  version = "0.1.0";

  src = ./.;

  vendorHash = "sha256-zecMVEVsYEZJ+lxFF12+GbhX+URtOGktOnC2xMLr+yo=";

  meta = with lib; {
    description = "";
    homepage = "";
  };
}
