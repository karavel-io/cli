{
  description = "Karavel CLI tool";

  inputs.nixpkgs.url = "github:nixos/nixpkgs";

  outputs = { self, nixpkgs }:
    let
      trim = s: nixpkgs.lib.strings.removePrefix " " (nixpkgs.lib.strings.removeSuffix " " s);
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = trim (builtins.readFile ./VERSION);
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {

      # Provide some binary packages for selected system types.
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          karavel = pkgs.buildGoModule {
            pname = "karavel";
            version = "v${version}";
            src = ./.;
            subPackages = [ "cmd/karavel" ];
            ldflags = [
              "-X github.com/karavel-io/cli/internal/version.version=${version}"
            ];

            vendorSha256 = "sha256-QWg69m8Ky8rZjX/B7T0yCBzfz6prZ2gy0HK6IrsBL5I==";
          };
        });

      # The default package for 'nix build'. This makes sense if the
      # flake provides only one package or there is a clear "main"
      # package.
      defaultPackage = forAllSystems (system: self.packages.${system}.karavel);
      defaultApp = forAllSystems (system: self.packages.${system}.karavel);
      devShell = forAllSystems
        (system:
          let pkgs = nixpkgsFor.${system}; in
          pkgs.mkShell {
            buildInputs = with pkgs; [
              go_1_18
              addlicense
              gnumake
              goreleaser
            ];
          });
    };
}
