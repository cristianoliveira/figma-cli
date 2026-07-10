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

        figma = pkgs.buildGoModule {
          pname = "figma";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-Y+fyGkDugE4WmjvhInJ4tp+7BtBxXZi+Pdhas90NaF0=";

          subPackages = [ "cmd/figma" ];
        };
      in {
        packages = {
          inherit figma;
          default = figma;
        };

        apps = {
          figma = utils.lib.mkApp { drv = figma; };
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
