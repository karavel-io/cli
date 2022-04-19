let
  pkgs = import <nixpkgs> { };
  addlicense = pkgs.callPackage ./.nix/addlicense.nix { };
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    unstable.go_1_18
    addlicense
    gnumake
    unstable.goreleaser
  ];
}
