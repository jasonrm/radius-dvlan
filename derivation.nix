{
  lib,
  buildGoModule,
}:
buildGoModule rec {
  pname = "radius-dvlan";
  version = "1.0.0";

  #  src = fetchFromGitHub {
  #    owner = "jasonrm";
  #    repo = "radius-dvlan";
  #    rev = version;
  #    hash = lib.fakeHash;
  #  };
  src = ./.;

  subPackages = ["."];

  installPhase = ''
    mkdir -p $out/bin
    install -m755 $GOPATH/bin/radius-dvlan $out/bin
  '';

  vendorHash = "sha256-tFzBMzHaBl2FP8lxq6FhcZZiOMuwrqOcWJUSYinrXAw=";

  meta = with lib; {
    description = "RADIUS server for MAC based dynamic VLAN assignment";
    license = licenses.mit;
    homepage = "https://github.com/jasonrm/radius-dvlan";
    maintainer = ["jason@mcneil.dev"];
  };
}
