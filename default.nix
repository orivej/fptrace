with import <nixpkgs> { };

buildGoPackage rec {
  name = "fptrace";
  src = ./.;
  goDeps = ./deps.nix;
  goPackagePath = "github.com/orivej/fptrace";
  preBuild = ''
    mkdir -p $bin/bin
    cc go/src/$goPackagePath/_fptracee.c -o $bin/bin/_fptracee
    buildFlagsArray=("-ldflags=-X main.tracee=$bin/bin/_fptracee")
  '';
}
