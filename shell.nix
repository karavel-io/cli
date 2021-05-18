let
  pkgs = import <nixpkgs> { };
  addlicense = pkgs.callPackage ./.nix/addlicense.nix { };
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    addlicense
    gnumake
  ];
}
