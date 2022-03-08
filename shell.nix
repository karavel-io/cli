let
  pkgs = import <nixpkgs> { };
  addlicense = pkgs.callPackage ./.nix/addlicense.nix { };
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    go_1_17
    addlicense
    gnumake
    unstable.goreleaser
  ];
}
