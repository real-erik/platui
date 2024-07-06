{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/706eef542dec88cc0ed25b9075d3037564b2d164.tar.gz") {} }:

pkgs.mkShell {
  packages = [
    pkgs.go
    pkgs.gopls
    pkgs.just
    pkgs.playwright
    pkgs.watchexec
  ];
}

