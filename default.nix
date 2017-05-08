with import <nixpkgs> { };

buildGoPackage rec {
  name = "fptrace";
  src = ./.;
  goDeps = ./deps.nix;
  goPackagePath = "github.com/orivej/fptrace";
  preBuild = ''
    mkdir -p $bin/bin
    cc go/src/$goPackagePath/tracee/tracee.c -o $bin/bin/tracee
    buildFlagsArray=("-ldflags=-X main.tracee=$bin/bin/tracee")
  '';
}
