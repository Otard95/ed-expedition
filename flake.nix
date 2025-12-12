{
  description = "pass-env is like env (the unix util) but gets the env values from pass";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  inputs.systems.url = "github:nix-systems/default";
  inputs.flake-utils = {
    url = "github:numtide/flake-utils";
    inputs.systems.follows = "systems";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        libs = with pkgs; [
          pkg-config
          gtk3
          webkitgtk_4_1
          gsettings-desktop-schemas
        ];
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            nodejs
            pnpm
            go
            wails
          ] ++ libs;

          buildInputs = libs;
          nativeBuildInputs = libs;

          shellHook = ''
            export LD_LIBRARY_PATH="${pkgs.lib.makeLibraryPath libs}:$LD_LIBRARY_PATH"
            export XDG_DATA_DIRS="${pkgs.gsettings-desktop-schemas}/share/gsettings-schemas/${pkgs.gsettings-desktop-schemas.name}:${pkgs.gtk3}/share/gsettings-schemas/${pkgs.gtk3.name}:$XDG_DATA_DIRS"
          '';
        };

        packages.default = pkgs.buildGoModule rec {

          pname = "pass-env";
          version = "";

          src = pkgs.fetchFromGitHub {
            owner = "otard95";
            repo = "ed-expedition";
            rev = "v${version}";
            hash = "sha256-0n7YaUOxnC2LUcsbitR9/rq1M4ghE4tR93LUIqRWB+E=";
          };

          vendorHash = "sha256-hpAsYPhiYnTpY5Z7QZz9cr5RtleHnR1ezgoVaQ+cvp0=";

          subPackages = ["."];

        };
      }
    );
}
