{
  description = "libp2p-playground";

  inputs = {
    nixpkgs.url = github:NixOS/nixpkgs/release-22.11;
    flake-utils.url = github:numtide/flake-utils;
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
    with nixpkgs.legacyPackages.${system};
    with lib;
    let
    in {
      devShell = stdenvNoCC.mkDerivation {
        name = "shell";
        buildInputs = [
          go_1_18
        ];
      };
    });
}
