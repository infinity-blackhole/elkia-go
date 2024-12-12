{
  perSystem =
    { pkgs, ... }:
    {
      packages = {
        default = pkgs.stdenv.mkDerivation {
          name = "manifests";
          src = ./.;
          installPhase = ''
            mkdir -p $out/share
            cp -r apps clusters infra $out/share
          '';
        };
      };
      treefmt = {
        projectRootFile = "flake.nix";
        enableDefaultExcludes = true;
        programs = {
          actionlint.enable = true;
          deadnix.enable = true;
          gofmt.enable = true;
          nixfmt.enable = true;
          prettier.enable = true;
          shfmt.enable = true;
          statix.enable = true;
          terraform.enable = true;
        };
        settings.global.excludes = [
          "*.excalidraw"
          "*.terraform.lock.hcl"
          ".gitattributes"
          "LICENSE"
        ];
      };
      devenv.shells.default = {
        containers = pkgs.lib.mkForce { };
        languages = {
          go.enable = true;
          terraform = {
            enable = true;
            package = pkgs.opentofu;
          };
          nix.enable = true;
        };
        cachix = {
          enable = true;
          push = "shikanime";
        };
        pre-commit.hooks = {
          hadolint.enable = true;
          shellcheck.enable = true;
          tflint.enable = true;
        };
        packages = [
          pkgs.gh
          pkgs.docker
          pkgs.gnumake
          pkgs.minikube
          pkgs.kubectl
          pkgs.skaffold
          pkgs.kustomize
          pkgs.kubernetes-helm
          pkgs.go
          pkgs.gotools
          pkgs.mockgen
          pkgs.protobuf
          pkgs.protoc-gen-go
          pkgs.protoc-gen-go-grpc
        ];
      };
    };
}
