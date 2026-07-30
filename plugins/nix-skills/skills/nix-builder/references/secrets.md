# secrets.md — Secrets Management Patterns

Reference for agenix declarations and Vault PKI integration.
Read when generating secrets declarations or Caddy TLS config.

---

## agenix Overview

Secrets are encrypted age files stored in the flake repo under `secrets/`.
They are decrypted at deploy time using the host's SSH host key.

Each secret is:
1. Declared in `secrets/secrets.nix` with the public keys that can decrypt it
2. Referenced in NixOS config via `config.age.secrets.{name}.path`
3. Decrypted to a path under `/run/agenix/` at boot (tmpfs — never on disk)

---

## secrets/secrets.nix

```nix
let
  # Host SSH public keys — update with real keys after each host is built
  media-host      = "ssh-ed25519 AAAA... root@media-host";
  gpu-host       = "ssh-ed25519 AAAA... root@gpu-host";
  service-host = "ssh-ed25519 AAAA... root@service-host";

  # Operator key — can decrypt any secret for emergency access
  admin    = "ssh-ed25519 AAAA... admin@workstation";

  allHosts   = [ media-host gpu-host service-host admin ];
in
{
  # service-host secrets
  "secrets/postgres-password.age".publicKeys       = [ service-host admin ];
  "secrets/gitea-db-password.age".publicKeys       = [ service-host admin ];
  "secrets/gitea-secret-key.age".publicKeys        = [ service-host admin ];
  "secrets/grafana-admin-password.age".publicKeys  = [ service-host admin ];
  "secrets/grafana-secret-key.age".publicKeys      = [ service-host admin ];
  "secrets/nextcloud-admin-password.age".publicKeys = [ service-host admin ];
  "secrets/nextcloud-db-password.age".publicKeys   = [ service-host admin ];
  "secrets/vault-unseal-key.age".publicKeys        = [ service-host admin ];
  "secrets/unifi-env.age".publicKeys               = [ service-host admin ];
  "secrets/immich-env.age".publicKeys              = [ gpu-host admin ];

  # Per-host secrets can reference only that host + admin
}
```

---

## Referencing secrets in NixOS config

```nix
# In any host or module config:
age.secrets.postgres-password = {
  file  = ../../secrets/postgres-password.age;
  owner = "postgres";   # systemd service user that reads this file
  mode  = "0400";
};

# Then reference in service config:
services.postgresql.initialScript = pkgs.writeText "init.sql" ''
  ALTER USER postgres PASSWORD '$(cat ${config.age.secrets.postgres-password.path})';
'';
```

**Important**: The `file` path in `age.secrets` is relative to the module file
location. From a file in `hosts/{host}/default.nix`, use `../../secrets/`.
From a file in `modules/services/`, use `../../secrets/`.

---

## Environment file secrets (for oci-containers)

For containers that accept env files, create a single `.age` file containing
all secrets for that container in KEY=VALUE format:

```
# immich.env (contents before encryption)
DB_PASSWORD=supersecret
DB_USERNAME=immich
DB_DATABASE_NAME=immich
REDIS_HOSTNAME=immich-redis
```

Reference in container config:
```nix
environmentFiles = [ config.age.secrets.immich-env.path ];
```

---

## Vault PKI for TLS

Vault acts as the internal CA. It exposes an ACME endpoint that Caddy uses
to obtain certificates for `*.home.{domain}.{tld}`.

### Vault PKI setup (one-time, done in existing Vault)
These steps are performed manually after Vault is running — not automated by Nix:
```bash
vault secrets enable pki
vault secrets tune -max-lease-ttl=87600h pki
vault write pki/root/generate/internal \
  common_name="homelab-root" ttl=87600h
vault write pki/config/urls \
  issuing_certificates="https://vault.home.example.com/v1/pki/cert/ca" \
  crl_distribution_points="https://vault.home.example.com/v1/pki/crl"
vault write pki/roles/homelab \
  allowed_domains="home.example.com" \
  allow_subdomains=true \
  max_ttl=720h
# Enable ACME
vault write pki/config/acme enabled=true
```

### Vault config.hcl (declared as environment.etc in Nix)

```nix
environment.etc."vault/config.hcl".text = ''
  ui = true
  storage "file" {
    path = "/vault/data"
  }
  listener "tcp" {
    address     = "0.0.0.0:8200"
    tls_disable = "true"   # TLS terminated by Caddy
  }
  api_addr = "https://vault.home.example.com"
'';
```

**Note**: Vault itself runs HTTP internally; Caddy provides TLS externally.
This avoids a chicken-and-egg problem where Vault needs a cert to issue certs.

---

## Creating secrets

After `secrets/secrets.nix` is declared, create each secret:

```bash
# From the flake root, using your SSH key
agenix -e secrets/postgres-password.age
# Opens $EDITOR — type the secret value, save, close
```

Secrets are committed to the repo encrypted. Anyone with a listed public key
can decrypt them at deploy time.

---

## Placeholder pattern

When generating configs before secrets are created, use this pattern in
annotations to flag secrets that need to be created:

```
SECRETS REQUIRED:
- secrets/postgres-password.age
  Create with: agenix -e secrets/postgres-password.age
  Used by: services.postgresql (password for postgres superuser)
  Owner: postgres
```
