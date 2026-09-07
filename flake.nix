{
  description = "Figma CLI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, utils, ... }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        mkCLI = pname: subPackage: pkgs.buildGoModule {
          inherit pname;
          version = "0.1.0";
          src = ./. ;
          vendorHash = "sha256-yN6RmmJD1ir+2LDnjMCySiiO31iE4jg/SuPp6FylrBw=";
          subPackages = [ subPackage ];
          proxyVendor = true;
        };
        figma = mkCLI "figma" "cmd/figma";
        pixel-perfect = mkCLI "pixel-perfect" "cmd/pixel-perfect";
      in {
        packages = {
          inherit figma pixel-perfect;
          default = figma;
        };

        apps = {
          figma = utils.lib.mkApp { drv = figma; };
          pixel-perfect = utils.lib.mkApp { drv = pixel-perfect; };
          default = utils.lib.mkApp { drv = figma; };
        };

        devShells.default = pkgs.mkShell {
           packages = with pkgs; [
             go

             golangci-lint
             gotools

             # Test runner with good output
             # USAGE: gotestsum --watch
             gotestsum

             # To create new subcommands, run:
             # cobra-cli add <subcommand-name>
             cobra-cli

             # To generate the mock for the interfaces, run:
             # mockgen -source=./pkg/cli/cli.go -destination=./pkg/cli/mock/mock_cli.go -package=mock
             mockgen

             # Pre-commit hooks manager
             lefthook

             # OpenAPI code generation (npm package for spec conversion)
             nodejs

             # Go code generation from OpenAPI specs
             oapi-codegen
           ];
        };
    });
}
