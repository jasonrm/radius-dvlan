{
  pkgs,
  lib,
  config,
  options,
  ...
}: let
  cfg = config.services.radius-dvlan;
  opts = options.services.radius-dvlan;
in {
  options = {
    services.radius-dvlan = {
      enable = lib.mkEnableOption "radius-dvlan";

      config = lib.mkOption {
        type = lib.types.attrs;
        default = {
          server = {
            listen = "0.0.0.0:1812";
            secret = "secret";
            defaultVlan = 1;
          };
        };
      };
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.radius-dvlan = {
      wantedBy = ["multi-user.target"];
      description = "RADIUS server for MAC based dynamic VLAN assignment";
      serviceConfig = {
        ExecStart =
          "${pkgs.radius-dvlan}/bin/radius-dvlan"
          + (lib.escapeShellArgs ["--config" (writeText "config.json" (builtins.toJSON cfg.config))]);
      };
    };
  };
}
