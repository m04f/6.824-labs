{ pkgs, ... }:
{
  packages = [ pkgs.zed-editor ];
  languages.go.enable = true;
  cachix.enable = false;
}
