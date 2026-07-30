# caddy.md — Caddy Reverse Proxy Patterns

Reference for declaring Caddy vhosts with Vault PKI TLS in NixOS.
Read when generating caddy.nix for any host.

---

## Overview

Each host runs its own Caddy instance. Caddy:
- Terminates TLS using certs from Vault PKI ACME
- Proxies to local services by port
- Runs as a NixOS service (`services.caddy`)

All services are accessed via `{service}.home.{domain}.{tld}`.
Caddy listens on 443. Services bind to localhost only.

---

## Base Caddy module pattern

```nix
# modules/services/caddy.nix
{ config, pkgs, lib, ... }:

let
  vaultAcme = "https://vault.home.{domain}.{tld}/v1/pki/acme/directory";

  mkVhost = port: {
    extraConfig = ''
      tls {
        ca ${vaultAcme}
      }
      reverse_proxy localhost:${toString port}
    '';
  };
in
{
  services.caddy = {
    enable = true;

    # Global options
    globalConfig = ''
      acme_ca ${vaultAcme}
      acme_ca_root /etc/ssl/certs/homelab-root-ca.crt
    '';

    virtualHosts = {
      # Populated per-host — see examples below
    };
  };

  # Firewall
  networking.firewall.allowedTCPPorts = [ 80 443 ];
}
```

---

## Root CA certificate

Caddy must trust the Vault PKI root CA to verify ACME responses.
Declare the root cert as a system certificate:

```nix
# In common/default.nix or per-host config
security.pki.certificateFiles = [
  /etc/ssl/certs/homelab-root-ca.crt
];
```

The actual root CA cert (`homelab-root-ca.crt`) must be placed on the host.
Declare it as an `environment.etc` entry sourced from the flake:

```nix
environment.etc."ssl/certs/homelab-root-ca.crt" = {
  source = ../../certs/homelab-root-ca.crt;
  mode   = "0444";
};
```

Add `certs/homelab-root-ca.crt` to the flake repo (this is a public cert —
safe to commit unencrypted).

---

## Per-host vhost examples

### media-host (Jellyfin + Frigate)

```nix
services.caddy.virtualHosts = {
  "jellyfin.home.{domain}.{tld}" = mkVhost 8096;
  "frigate.home.{domain}.{tld}"  = mkVhost 5000;
};
```

### gpu-host (Immich + OpenWebUI + ComfyUI)

```nix
services.caddy.virtualHosts = {
  "immich.home.{domain}.{tld}"    = mkVhost 2283;
  "openwebui.home.{domain}.{tld}" = mkVhost 3000;
  "comfyui.home.{domain}.{tld}"   = mkVhost 8188;
};
```

### service-host (all service-hub vhosts)

```nix
services.caddy.virtualHosts = {
  "vault.home.{domain}.{tld}"        = mkVhost 8200;
  "git.home.{domain}.{tld}"          = mkVhost 3000;  # Gitea
  "ha.home.{domain}.{tld}"           = mkVhost 8123;  # HomeAssistant
  "grafana.home.{domain}.{tld}"      = mkVhost 3001;  # note: avoid clash with Gitea
  "prometheus.home.{domain}.{tld}"   = mkVhost 9090;
  "nextcloud.home.{domain}.{tld}"    = mkVhost 80;    # Nextcloud manages its own nginx
  "unifi.home.{domain}.{tld}"        = {
    extraConfig = ''
      tls { ca ${vaultAcme} }
      reverse_proxy localhost:8443 {
        transport http { tls_insecure_skip_verify }
      }
    '';
  };
  "homepage.home.{domain}.{tld}"     = mkVhost 3000;
};
```

**Notes**:
- Unifi uses self-signed HTTPS internally — use `tls_insecure_skip_verify`
  when proxying to it
- Nextcloud declares its own nginx on port 80 — Caddy proxies to that
- Gitea and Grafana both default to port 3000 — change one:
  set `services.grafana.settings.server.http_port = 3001`

---

## Port conflict resolution (service-host)

| Service | Default port | Configured port |
|---|---|---|
| Gitea | 3000 | 3000 |
| Grafana | 3000 | 3001 |
| OpenWebUI | 3000 | — (on gpu-host, no conflict) |
| Loki | 3100 | 3100 |
| Prometheus | 9090 | 9090 |
| HomeAssistant | 8123 | 8123 |
| Vault | 8200 | 8200 |
| Unifi | 8443 | 8443 |
| Nextcloud | 80 | 80 |
| Homepage | 3000 | 3002 |

---

## Nextcloud special handling

Nextcloud's NixOS module manages its own nginx. Caddy proxies to it.
Do NOT declare a Caddy vhost that conflicts with Nextcloud's nginx.
Instead, Nextcloud's nginx binds to `127.0.0.1:80` and Caddy proxies:

```nix
services.nextcloud.config.extraTrustedDomains = [ "nextcloud.home.{domain}.{tld}" ];
services.nextcloud.nginx.recommendedHttpHeaders = true;

# In caddy.nix:
"nextcloud.home.{domain}.{tld}" = {
  extraConfig = ''
    tls { ca ${vaultAcme} }
    reverse_proxy localhost:80 {
      header_up Host {host}
      header_up X-Forwarded-Proto https
    }
  '';
};
```
