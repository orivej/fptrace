with import <nixpkgs> { };

buildGoModule rec {
  name = "fptrace";
  src = lib.cleanSource ./.;
  vendorHash = "sha256-hk2FEff/37yJVlpOcca0KgSnI+gTylVhqcYiIjzp/i8=";
  ldflags = [
    "-X main.tracee=${placeholder "out"}/bin/_fptracee"
  ];
  subPackages = [ "." ];
  preBuild = ''
    mkdir -p $out/bin
    go run seccomp/seccomp.go
    cc _fptracee.c -o $out/bin/_fptracee
  '';
  overrideModAttrs.preBuild = "";
}
