{ pkgs, lib, ... }:

let
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
  languages.go.enable = true;

  languages.javascript = {
    enable = true;
    package = pkgs.nodejs;
    pnpm.enable = true;
  };

  packages = [ pkgs.wails ] ++ libs;

  env.ED_DEV_MODE = "1";

  enterShell = ''
    export LD_LIBRARY_PATH="${lib.makeLibraryPath libs}:$LD_LIBRARY_PATH"
    export XDG_DATA_DIRS="${pkgs.gsettings-desktop-schemas}/share/gsettings-schemas/${pkgs.gsettings-desktop-schemas.name}:${pkgs.gtk3}/share/gsettings-schemas/${pkgs.gtk3.name}:$XDG_DATA_DIRS"
    export ED_EXPEDITION_DATA_DIR="$DEVENV_ROOT/data/local/share"
    export ED_EXPEDITION_CACHE_DIR="$DEVENV_ROOT/data/cache"
    export ED_EXPEDITION_CONFIG_DIR="$DEVENV_ROOT/data/config"
    export ED_EXPEDITION_JOURNAL_DIR="$DEVENV_ROOT/data/journals"
  '';
}
