{
  description = "Elite Dangerous expedition planning and tracking tool";
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
          glib
          gtk3
          gdk-pixbuf
          webkitgtk_4_1
          libsoup_3
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

        packages.default = pkgs.stdenv.mkDerivation rec {
          pname = "ed-expedition";
          version = "0.3.0";

          src = pkgs.fetchurl {
            url = "https://github.com/Otard95/ed-expedition/releases/download/v${version}/ed-expedition-linux-amd64-webkit2_41.tar.gz";
            # TODO: Update hash after first release (v0.0.1) is published
            # Run: nix-prefetch-url <url> to get the real hash
            hash = "sha256-7dk62H1JkdlYRLJkqjE7eyCMot/aZ9/aKp999gdoN2w=";
          };

          nativeBuildInputs = [ pkgs.makeWrapper ];
          buildInputs = libs;

          sourceRoot = ".";

          unpackPhase = ''
            tar -xzf $src
          '';

          installPhase = ''
            runHook preInstall

            mkdir -p $out/bin
            install -m755 ed-expedition $out/bin/

            # Wrap binary with required library paths
            wrapProgram $out/bin/ed-expedition \
              --prefix LD_LIBRARY_PATH : "${pkgs.lib.makeLibraryPath libs}" \
              --prefix XDG_DATA_DIRS : "${pkgs.gsettings-desktop-schemas}/share/gsettings-schemas/${pkgs.gsettings-desktop-schemas.name}:${pkgs.gtk3}/share/gsettings-schemas/${pkgs.gtk3.name}"

            runHook postInstall
          '';

          meta = with pkgs.lib; {
            description = "Elite Dangerous expedition planning and tracking tool";
            homepage = "https://github.com/Otard95/ed-expedition";
            license = licenses.gpl2Only;
            platforms = [ "x86_64-linux" ];
            mainProgram = "ed-expedition";
          };
        };
      }
    );
}
