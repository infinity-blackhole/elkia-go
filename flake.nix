{
  description = "Elkia game server";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs";
    devenv = {
      url = "github:cachix/devenv";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.pre-commit-hooks.follows = "pre-commit-hooks";
    };
    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  nixConfig = {
    extra-public-keys = [
      "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw="
    ];
    extra-substituters = [
      "https://devenv.cachix.org"
    ];
  };

  outputs = { nixpkgs, devenv, ... }@inputs: {
    devShells = nixpkgs.lib.genAttrs nixpkgs.lib.platforms.unix (system:
      let pkgs = import nixpkgs { inherit system; }; in {
        default = devenv.lib.mkShell {
          inherit inputs pkgs;
          modules = [
            {
              pre-commit.hooks = {
                actionlint.enable = true;
                markdownlint.enable = true;
                shellcheck.enable = true;
                nixpkgs-fmt.enable = true;
                statix.enable = true;
                deadnix.enable = true;
                hadolint.enable = true;
              };
              packages = [
                pkgs.nixpkgs-fmt
                pkgs.docker
                pkgs.gnumake
              ];
            }
            {
              pre-commit.hooks.gofmt.enable = true;
              packages = [
                pkgs.go
                pkgs.gotools
                pkgs.mockgen
                pkgs.protobuf
                pkgs.protoc-gen-go
                pkgs.protoc-gen-go-grpc
              ];
            }
            {
              packages = [
                pkgs.kind
                pkgs.skaffold
                pkgs.kustomize
                pkgs.kubectl
                pkgs.kubernetes-helm
                pkgs.google-cloud-sdk
                pkgs.redis
                pkgs.apacheKafka
              ];
            }
            {
              packages = [
                pkgs.erlang
                pkgs.rebar3
                pkgs.elixir
              ];
            }
          ];
        };
      }
    );
  };
}
