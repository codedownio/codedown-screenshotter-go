{ lib
, buildGoModule
}:

buildGoModule rec {
  pname = "pet";
  version = "0.3.4";

  src = ./.;

  vendorHash = "sha256-ciBIR+a1oaYH+H1PcC8cD8ncfJczk1IiJ8iYNM+R4aA=";

  meta = with lib; {
    description = "";
    homepage = "";
    license = licenses.mit;
    maintainers = with maintainers; [ kalbasit ];
  };
}
