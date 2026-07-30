# services.md — NixOS Service Patterns

Reference for correct NixOS module options per service.
Read this file before implementing any service.

> **Release caveat.** The option names and shapes below were verified against
> **NixOS 24.11**. Option shapes drift between releases. Confirm each against
> the nixpkgs revision in the target repository's `flake.lock` — see
> https://search.nixos.org/options — and treat a mismatch as the release being
> right and this file being stale.

---

## Implementation Strategy Quick Reference

| Service | Strategy | Module / image |
|---|---|---|
| PostgreSQL | Native module | `services.postgresql` |
| Gitea | Native module | `services.gitea` |
| HomeAssistant | Native module | `services.home-assistant` |
| Prometheus | Native module | `services.prometheus` |
| Loki | Native module | `services.loki` |
| Grafana | Native module | `services.grafana` |
| Jellyfin | Native module | `services.jellyfin` |
| Nextcloud | Native module | `services.nextcloud` |
| Frigate | oci-container | `ghcr.io/blakeblackshear/frigate:stable` |
| Immich | oci-container | `ghcr.io/immich-app/immich-server:release` |
| Unifi Controller | oci-container | `lscr.io/linuxserver/unifi-network-application:latest` |
| Vault | oci-container | `hashicorp/vault:latest` |
| OpenWebUI | oci-container | `ghcr.io/open-webui/open-webui:main` |
| ComfyUI | oci-container | `ghcr.io/ai-dock/comfyui:latest` |
| Homepage | oci-container | `ghcr.io/gethomepage/homepage:latest` |

---

## Native Module Patterns (NixOS 24.11)

### PostgreSQL

```nix
services.postgresql = {
  enable = true;
  package = pkgs.postgresql_16;
  ensureDatabases = [ "gitea" "nextcloud" ];
  ensureUsers = [
    { name = "gitea";     ensureDBOwnership = true; }
    { name = "nextcloud"; ensureDBOwnership = true; }
  ];
  authentication = pkgs.lib.mkOverride 10 ''
    local all all              trust
    host  all all 127.0.0.1/32 scram-sha-256
  '';
};
```

**Notes**:
- `ensureUsers` does NOT set passwords — use agenix + `initialScript` or
  `ALTER USER` via a systemd oneshot for password setting
- Data dir defaults to `/var/lib/postgresql/{version}`
- Declare `services.postgresql.dataDir` explicitly if using ZFS dataset

---

### Gitea

```nix
services.gitea = {
  enable = true;
  database = {
    type = "postgres";
    host = "127.0.0.1";
    port = 5432;
    name = "gitea";
    user = "gitea";
    passwordFile = config.age.secrets.gitea-db-password.path;
  };
  settings = {
    server = {
      DOMAIN   = "git.home.example.com";
      HTTP_PORT = 3000;
      ROOT_URL = "https://git.home.example.com";
    };
    session.COOKIE_SECURE = true;
  };
};
```

**Notes**:
- `passwordFile` is the correct option name in 24.11 (not `password`)
- Gitea runs as user `gitea` — agenix secret owner must be `gitea`
- App data at `/var/lib/gitea` by default

---

### HomeAssistant

```nix
services.home-assistant = {
  enable = true;
  configDir = "/var/lib/hass";
  extraComponents = [
    "met"
    "radio_browser"
    "cast"
    "zha"           # add/remove based on actual integrations in use
  ];
  config = {
    homeassistant = {
      name      = "Home";
      time_zone = "Europe/London";  # replace with actual timezone
    };
    http = {
      server_port         = 8123;
      use_x_forwarded_for = true;
      trusted_proxies     = [ "127.0.0.1" ];
    };
  };
};
```

**Notes**:
- `extraComponents` triggers pip installs on first start — slow on first boot
- Existing HA config can be placed in `configDir` before starting
- HA runs as user `hass`

---

### Prometheus

```nix
services.prometheus = {
  enable     = true;
  port       = 9090;
  retentionTime = "30d";

  scrapeConfigs = [
    {
      job_name = "node";
      static_configs = [{
        targets = [
          "media-host.lan:9100"
          "gpu-host.lan:9100"
          "service-host.lan:9100"
        ];
      }];
    }
  ];
};
```

**Notes**:
- Node exporter declared separately in `modules/common/monitoring.nix`
- Data at `/var/lib/prometheus2` — declare ZFS dataset here

---

### Prometheus Node Exporter (common/monitoring.nix)

```nix
services.prometheus.exporters.node = {
  enable          = true;
  port            = 9100;
  enabledCollectors = [ "systemd" "processes" ];
  openFirewall    = true;
};
```

---

### Loki

```nix
services.loki = {
  enable = true;
  configuration = {
    server.http_listen_port = 3100;
    auth_enabled = false;

    ingester = {
      lifecycler = {
        address = "127.0.0.1";
        ring.kvstore.store = "inmemory";
        ring.replication_factor = 1;
      };
      chunk_idle_period   = "1h";
      max_chunk_age       = "1h";
      chunk_retain_period = "30s";
    };

    schema_config.configs = [{
      from         = "2024-01-01";
      store        = "tsdb";
      object_store = "filesystem";
      schema       = "v13";
      index = {
        prefix = "index_";
        period = "24h";
      };
    }];

    storage_config = {
      tsdb_shipper = {
        active_index_directory = "/var/lib/loki/tsdb-index";
        cache_location         = "/var/lib/loki/tsdb-cache";
      };
      filesystem.directory = "/var/lib/loki/chunks";
    };

    limits_config = {
      reject_old_samples      = true;
      reject_old_samples_max_age = "168h";
    };
  };
};
```

---

### Grafana

```nix
services.grafana = {
  enable = true;
  settings = {
    server = {
      http_addr = "127.0.0.1";
      http_port = 3000;
      domain    = "grafana.home.example.com";
      root_url  = "https://grafana.home.example.com";
    };
    security = {
      admin_user              = "admin";
      admin_password_file     = config.age.secrets.grafana-admin-password.path;
      secret_key_file         = config.age.secrets.grafana-secret-key.path;
    };
  };

  provision = {
    enable = true;
    datasources.settings.datasources = [
      {
        name   = "Prometheus";
        type   = "prometheus";
        url    = "http://127.0.0.1:9090";
        isDefault = true;
      }
      {
        name = "Loki";
        type = "loki";
        url  = "http://127.0.0.1:3100";
      }
    ];
  };
};
```

**Notes**:
- `admin_password_file` and `secret_key_file` are 24.11 options (not `admin_password`)
- Datasources declared in Nix — no manual Grafana UI setup needed

---

### Jellyfin

```nix
services.jellyfin = {
  enable    = true;
  openFirewall = true;
  # dataDir defaults to /var/lib/jellyfin
  # cacheDir defaults to /var/cache/jellyfin
};

# NFS mount for media
fileSystems."/mnt/nas/jellyfin" = {
  device  = "nas.lan:/volume1/jellyfin";
  fsType  = "nfs";
  options = [
    "nfsvers=4"
    "soft"
    "timeo=30"
    "x-systemd.automount"
    "x-systemd.idle-timeout=600"
    "noatime"
  ];
};
```

**Notes**:
- Jellyfin runs as user `jellyfin` — ensure NFS export allows this UID or use
  `mapall` on Synology side
- Hardware transcoding (Intel QSV, VAAPI) is day-2 — do not configure on VM

---

### Nextcloud

```nix
services.nextcloud = {
  enable   = true;
  package  = pkgs.nextcloud29;
  hostName = "nextcloud.home.example.com";
  https    = true;

  config = {
    adminpassFile = config.age.secrets.nextcloud-admin-password.path;
    dbtype        = "pgsql";
    dbhost        = "127.0.0.1";
    dbname        = "nextcloud";
    dbuser        = "nextcloud";
    dbpassFile    = config.age.secrets.nextcloud-db-password.path;
  };

  settings = {
    trusted_domains  = [ "nextcloud.home.example.com" ];
    trusted_proxies  = [ "127.0.0.1" ];
    overwriteprotocol = "https";
  };

  # Data directory on NFS
  datadir = "/mnt/nas/nextcloud";
};

fileSystems."/mnt/nas/nextcloud" = {
  device  = "nas.lan:/volume1/nextcloud";
  fsType  = "nfs";
  options = [ "nfsvers=4" "soft" "timeo=30" "x-systemd.automount" "noatime" ];
};
```

**Notes**:
- `package` must match the currently deployed data version — check the existing
  Nextcloud version before setting this, since Nextcloud refuses to skip major
  versions on upgrade
- Nextcloud module manages its own PHP-FPM and nginx — do not declare a
  separate nginx vhost for it; Caddy should proxy to the Nextcloud-managed port

---

## oci-containers Patterns

### Common boilerplate

```nix
virtualisation.oci-containers = {
  backend = "podman";
  containers.{name} = {
    image   = "{image}:{tag}";
    volumes = [ "/host/path:/container/path" ];
    ports   = [ "host_port:container_port" ];
    environment = {
      KEY = "value";
    };
    environmentFiles = [
      config.age.secrets.{name}-env.path
    ];
    extraOptions = [ "--network=host" ];  # only if needed
  };
};
```

**Notes**:
- Prefer named volumes (`/var/lib/{service}:/data`) over anonymous volumes
- Use `environmentFiles` for secrets — never hardcode in `environment`
- Podman systemd unit name: `podman-{container-name}.service`
- Always pin image tags — never use `latest` in production configs

### Frigate

```nix
virtualisation.oci-containers.containers.frigate = {
  image   = "ghcr.io/blakeblackshear/frigate:0.14.1";
  volumes = [
    "/etc/frigate/config.yaml:/config/config.yml:ro"
    "/var/lib/frigate:/media/frigate"
  ];
  ports = [
    "5000:5000"
    "8554:8554"
    "8555:8555"
  ];
  environment.TZ = "Europe/London";
  # Coral TPU: day-2 — add after bare metal deploy:
  # devices = [ "/dev/bus/usb:/dev/bus/usb" ];
  # extraOptions = [ "--privileged" ];
};
```

### Immich

Immich requires multiple containers (server, microservices, machine-learning,
redis, postgres). Use a Podman pod or declare each container and link via
a shared network.

```nix
# Immich uses its own Postgres — do not use shared services.postgresql
virtualisation.oci-containers.containers = {
  immich-server = {
    image = "ghcr.io/immich-app/immich-server:v1.106.4";
    volumes = [
      "/mnt/nas/immich:/usr/src/app/upload"
      "/etc/localtime:/etc/localtime:ro"
    ];
    ports           = [ "2283:3001" ];
    environmentFiles = [ config.age.secrets.immich-env.path ];
    dependsOn       = [ "immich-postgres" "immich-redis" ];
  };

  immich-machine-learning = {
    image   = "ghcr.io/immich-app/immich-machine-learning:v1.106.4";
    volumes = [ "immich-model-cache:/cache" ];
    environmentFiles = [ config.age.secrets.immich-env.path ];
  };

  immich-redis = {
    image = "registry.hub.docker.com/library/redis:6.2-alpine";
  };

  immich-postgres = {
    image = "registry.hub.docker.com/tensorchord/pgvecto-rs:pg14-v0.2.0";
    volumes = [ "/var/lib/immich-postgres:/var/lib/postgresql/data" ];
    environmentFiles = [ config.age.secrets.immich-env.path ];
  };
};
```

**Notes**:
- Immich requires `pgvecto-rs` Postgres image — not standard PostgreSQL
- Pin versions: server, machine-learning, and postgres versions must match
- `immich-env` secret file should contain: `DB_PASSWORD`, `DB_USERNAME`,
  `DB_DATABASE_NAME`, `REDIS_HOSTNAME`

### Unifi Controller

```nix
virtualisation.oci-containers.containers.unifi = {
  image = "lscr.io/linuxserver/unifi-network-application:8.4.59";
  volumes = [ "/var/lib/unifi:/config" ];
  ports = [
    "8443:8443"
    "3478:3478/udp"
    "10001:10001/udp"
    "8080:8080"
  ];
  environment = {
    PUID     = "1000";
    PGID     = "1000";
    TZ       = "Europe/London";
    MONGO_USER = "unifi";
    MONGO_DBNAME = "unifi";
    MONGO_HOST = "immich-mongo"; # or separate mongo container
  };
  environmentFiles = [ config.age.secrets.unifi-env.path ];
};
```

**Notes**:
- Unifi requires MongoDB — declare a separate mongo container or use the
  linuxserver image's bundled mongo (set `MONGO_HOST=localhost` and add
  `--network=host` or use a pod)
- Ports 3478/udp and 10001/udp must be open in firewall for device adoption

### Vault

```nix
virtualisation.oci-containers.containers.vault = {
  image = "hashicorp/vault:1.17.3";
  volumes = [
    "/var/lib/vault/data:/vault/data"
    "/etc/vault/config.hcl:/vault/config/config.hcl:ro"
  ];
  ports        = [ "8200:8200" ];
  capabilities = { IPC_LOCK = true; };
  cmd          = [ "vault" "server" "-config=/vault/config/config.hcl" ];
};
```

Vault config file (`/etc/vault/config.hcl`) should be declared as a
`environment.etc` entry in Nix, not inside the container image.
