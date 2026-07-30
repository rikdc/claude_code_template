# zfs.md — ZFS Configuration Patterns

Reference for ZFS pool layout and NixOS dataset declarations.
Read when generating ZFS config for any host.

---

## Pool Layout (standard, per host)

```
rpool
├── local              # datasets that can be wiped without data loss
│   ├── nix            # /nix store — large, compressible, rebuildable
│   └── root           # / — root filesystem
└── safe               # datasets that must survive reboots
    ├── home           # /home
    ├── persist        # /persist — explicit state (impermanence pattern)
    └── data           # /var/lib — service state
```

---

## NixOS ZFS declarations

```nix
# In host default.nix or profiles/bare-metal.nix

boot = {
  supportedFilesystems = [ "zfs" ];
  zfs.forceImportRoot  = false;
  zfs.devNodes         = "/dev/disk/by-id";  # stable device names
};

services.zfs = {
  autoScrub.enable   = true;
  autoScrub.interval = "monthly";
  trim.enable        = true;   # for SSDs
};

fileSystems = {
  "/" = {
    device  = "rpool/local/root";
    fsType  = "zfs";
    options = [ "zfsutil" "X-mount.mkdir" ];
  };
  "/nix" = {
    device  = "rpool/local/nix";
    fsType  = "zfs";
    options = [ "zfsutil" "X-mount.mkdir" ];
  };
  "/home" = {
    device  = "rpool/safe/home";
    fsType  = "zfs";
    options = [ "zfsutil" "X-mount.mkdir" ];
  };
  "/persist" = {
    device  = "rpool/safe/persist";
    fsType  = "zfs";
    options = [ "zfsutil" "X-mount.mkdir" ];
    neededForBoot = true;
  };
  "/var/lib" = {
    device  = "rpool/safe/data";
    fsType  = "zfs";
    options = [ "zfsutil" "X-mount.mkdir" ];
  };
};
```

---

## Dataset properties (set at pool creation, not in Nix config)

```bash
# At install time, before nixos-install:
zpool create \
  -o ashift=12 \
  -O compression=zstd \
  -O atime=off \
  -O xattr=sa \
  -O acltype=posixacl \
  rpool /dev/disk/by-id/{disk-id}

zfs create -o mountpoint=none         rpool/local
zfs create -o mountpoint=legacy       rpool/local/nix
zfs create -o mountpoint=legacy       rpool/local/root
zfs create -o mountpoint=none         rpool/safe
zfs create -o mountpoint=legacy       rpool/safe/home
zfs create -o mountpoint=legacy       rpool/safe/persist
zfs create -o mountpoint=legacy       rpool/safe/data
```

**Note**: The provisioning tool (Component 2) handles this — the Nix config
declares the mountpoints, the tool creates the datasets. Do not embed zpool
create commands in Nix config.

---

## Snapshot before nixos-rebuild (recommended pattern)

Document in the host runbook:

```bash
# Before any nixos-rebuild switch:
zfs snapshot rpool/local/root@pre-rebuild-$(date +%Y%m%d-%H%M)
zfs snapshot rpool/safe/data@pre-rebuild-$(date +%Y%m%d-%H%M)

# Rollback if needed:
zfs rollback rpool/local/root@pre-rebuild-20250412-1430
```

---

## Impermanence pattern (optional, per host)

With impermanence, `/` is wiped on each boot. Only `/persist` survives.
Services must declare what state they need explicitly.

```nix
# Only use if impermanence is opted in for this host (see ARCHITECTURE.md)
boot.initrd.postDeviceCommands = pkgs.lib.mkAfter ''
  zfs rollback -r rpool/local/root@blank
'';

# Requires a blank snapshot created at install time:
# zfs snapshot rpool/local/root@blank

environment.persistence."/persist" = {
  hideMounts = true;
  directories = [
    "/var/log"
    "/var/lib/nixos"
    "/var/lib/systemd/coredump"
    "/etc/NetworkManager/system-connections"
  ];
  files = [
    "/etc/machine-id"
    "/etc/ssh/ssh_host_ed25519_key"
    "/etc/ssh/ssh_host_ed25519_key.pub"
  ];
};
```

**Decision**: Impermanence is opt-in per host. Check ARCHITECTURE.md open
decisions before applying. Do not generate impermanence config unless
explicitly confirmed.

---

## VM guest ZFS notes

For Proxmox VMs using virtio disks:
- Use `by-id` paths: `/dev/disk/by-id/virtio-{serial}`
- Set serial in Proxmox VM config to get a stable ID
- `ashift=12` is correct for virtio block devices
- Trim works via `fstrim` — enable `services.zfs.trim.enable = true`
