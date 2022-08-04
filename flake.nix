{
  description = "Karavel CLI tool";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        trim = s: nixpkgs.lib.strings.removePrefix " v" (nixpkgs.lib.strings.removeSuffix " " (nixpkgs.lib.removeSuffix "\n" s));
        lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
        version = trim (builtins.readFile ./VERSION);
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "karavel";
          version = "v${version}";
          src = ./.;
          subPackages = [ "cmd/karavel" ];
          ldflags = [
            "-X github.com/karavel-io/cli/internal/version.version=${version}"
          ];

          vendorSha256 = "sha256-8Ty26vr/J1SgY/+W4ZpMM1RpxrXesByXRTylOqgpYuc=";
        };

        apps.default = utils.lib.mkApp { drv = self.packages.${system}.default; };
        devShells.default = pkgs.mkShell
          {
            buildInputs = with pkgs; [
              go_1_18
              addlicense
              gnumake
              goreleaser
            ];
          };
      });
}
