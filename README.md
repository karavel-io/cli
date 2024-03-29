# Karavel CLI

The Karavel CLI implements different tools used by various Karavel subprojects. It is primarily used by the [Karavel Container Platform]
to render components configurations (see the [official documentation](https://platform.karavel.io/cli/) for more information).

```
Sailing smoothly on the Cloud sea

Usage:
  karavel [command]

Available Commands:
  help        Help about any command
  init        Initialize a new Karavel project
  render      Render a Karavel project
  version     Prints the CLI version and exits

Flags:
      --colors    Enable colored logs (default true)
  -d, --debug     Output debug logs
  -h, --help      help for karavel
  -q, --quiet     Suppress all logs except errors
  -v, --version   version for karavel

Use "karavel [command] --help" for more information about a command.
```

## Install

Binaries for all mainstream operating systems can be downloaded from [GitHub](https://github.com/karavel-io/cli/releases).

### Nix

The CLI is packaged as a [Flake](https://nixos.wiki/wiki/Flakes). You can run it as a simple command:

`nix run github:karavel-io/cli <regular karavel arguments>`

e.g.

`nix run github:karavel-io/cli render --debug`

Or you can import it in another `flake.nix`, e.g. to add it to a [devShell](https://nixos.wiki/wiki/Development_environment_with_nix-shell):

```nix
{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
    karavel.url = "github:karavel-io/cli";
  };

  outputs = { self, nixpkgs, flake-utils, karavel }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        karavel-cli = karavel.defaultPackage.${system};
      in {
        devShell = pkgs.mkShell { buildInputs = with pkgs; [
          karavel-cli
          kubectl
          kustomize
        ];
        };
      });
}
```

By appending a refernce at the end of the Flake URL you can select:
 - a specific tag/version: `github:karavel-io/cli/v0.4.1`
 - a Git branch: `github:karavel-io/cli/main`

More information on references can be found on the [Flakes manual](https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake.html#flake-references).

### Docker

The CLI is packaged in a container image and published on [Quay](https://quay.io/karavel/cli) and [GitHub](https://github.com/karavel-io/cli/pkgs/container/cli).

You can run it like so:

```bash
# Inside a Karavel project directory
$ docker run --rm -v $PWD:/karavel -u (id -u) quay.io/karavel/cli:main render
$ docker run --rm -v $PWD:/karavel -u (id -u) ghcr.io/karavel-io/cli:main render
```

Stable releases are tagged using their semver (`x.y.z`) version, with aliases to the latest patch (`x.y`) and minor (`x`) versions.
This is what you should be using most of the time.  
The `main` tag points to the latest unstable build from the `main` branch. It's useful if you want to try out the latest
features before they are released.

## Requirements

- Go 1.18+
- make

For [Nix] or [NixOS] users, the provided [shell.nix](shell.nix) already configures the required tooling.

## Build

`make` outputs the `karavel` executable in the `bin` folder
`make install` installs the executable in the PATH. Install location can be changed by passing the `INSTALL_PATH` variable:
`make INSTALL_PATH=/path/to/karavel install`

[karavel container platform]: https://platform.karavel.io
[nix]: https://nixos.org/explore.html
[nixos]: https://nixos.org
