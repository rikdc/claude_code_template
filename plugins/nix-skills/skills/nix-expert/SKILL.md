---
name: nix-expert
description: >
  Expert guidance on Nix, NixOS, home-manager, flakes, and Nix packaging. Use
  when writing or debugging Nix expressions, configuring NixOS or home-manager,
  authoring derivations, managing flake inputs, encrypting secrets with agenix
  or sops-nix, or hardening a NixOS host. Triggers on Nix, NixOS, nixpkgs,
  home-manager, flake.nix, nixos-rebuild, nix build, nix develop, devShell,
  derivation, overlay, agenix, and on Nix errors such as hash mismatches,
  infinite recursion, or evaluation failures. For creating or modifying host
  configurations in an existing flake repository, use nix-builder instead.
---

# Nix Expert

Advisory guidance on Nix, NixOS, and home-manager. This skill answers questions
and reviews Nix code. It does not author host configurations — see
`nix-builder` for that.

## Scope

Use this skill for:

- Explaining or reviewing Nix expressions and module definitions
- Authoring derivations and overlays
- Flake input management and lockfile hygiene
- Debugging build and evaluation failures
- home-manager configuration
- Secrets management patterns
- NixOS security hardening

Defer to `nix-builder` when the task is to create or change `.nix` files for a
specific host in an existing flake repository.

## Workflow Routing

Route to the matching workflow file rather than answering from this file when
the request is specific:

| Workflow | Trigger | File |
|----------|---------|------|
| **Build** | "build nix package", "nixos-rebuild build", "compile nix" | `workflows/Build.md` |
| **Debug** | "debug nix", "nix error", "evaluation error", "infinite recursion" | `workflows/Debug.md` |
| **Develop** | "development shell", "nix develop", "devShell", "direnv" | `workflows/Develop.md` |
| **Deploy** | "deploy nixos", "nixos-rebuild switch", "remote deployment" | `workflows/Deploy.md` |
| **Package** | "create package", "derivation", "buildGoModule", "package app" | `workflows/Package.md` |
| **Flakes** | "create flake", "flake.lock", "update inputs", "flake outputs" | `workflows/Flakes.md` |
| **Secrets** | "manage secrets", "agenix", "encrypt secrets", "age encryption" | `workflows/Secrets.md` |
| **Security** | "harden nixos", "apparmor", "firewall", "security hardening" | `workflows/Security.md` |
| **Troubleshoot** | "hash mismatch", "nix failing", "common errors" | `workflows/Troubleshoot.md` |

Read only the workflow you need. For general guidance, continue with this file.

## Core Principles

### 1. Declarative over imperative

```nix
# Good
services.nginx.enable = true;

# Bad — imperative escape hatch
systemd.services.nginx.postStart = "systemctl start nginx";
```

### 2. Reproducibility

Same inputs produce the same outputs. Pin versions explicitly, commit
`flake.lock`, and avoid impure operations (`builtins.getEnv`, `--impure`,
unpinned `fetchTarball`).

### 3. Modularity

```nix
# Good
imports = [
  ./hardware.nix
  ./networking.nix
  ./services.nix
];
```

### 4. Version control everything

Track all Nix configuration in git, commit `flake.lock` changes deliberately,
and record why an input was pinned or overridden.

### 5. Prefer flakes

Flakes give hermetic evaluation, a standard output schema, dependency locking,
and better caching.

## NixOS Configuration Patterns

### Host configuration structure

```
hosts/<hostname>/
├── default.nix    # Imports modules, host-specific config
├── boot.nix       # Bootloader, initrd, kernel modules
└── hardware.nix   # Hardware settings, filesystems, mounts
```

Layouts vary between repositories. Read the existing tree before assuming one.

### Shared module organization

```
modules/
├── common/        # Baseline imported by every host
├── profiles/      # Opt-in role profiles (bare-metal, vm-guest, nvidia)
└── services/      # One file per service
```

### A mkHost helper

Many flakes wrap host construction to avoid repetition:

```nix
nixosConfigurations = {
  some-host = libx.mkHost {
    hostname = "some-host";
    system = "x86_64-linux";
    profile = "bare-metal";
  };
};
```

If the repository has such a helper, use it rather than calling
`nixpkgs.lib.nixosSystem` directly.

## Module Best Practices

### Define options properly

```nix
{ config, lib, pkgs, ... }:

{
  options.services.myservice = {
    enable = lib.mkEnableOption "my service";

    port = lib.mkOption {
      type = lib.types.port;
      default = 8080;
      description = "Port to listen on";
    };

    configFile = lib.mkOption {
      type = lib.types.path;
      description = "Path to configuration file";
    };
  };

  config = lib.mkIf config.services.myservice.enable {
    # Implementation
  };
}
```

### Common types

`types.bool`, `types.int`, `types.str`, `types.path`, `types.port`,
`types.listOf types.str`, `types.attrsOf`, `types.submodule`, `types.package`,
`types.nullOr`, `types.enum [ ... ]`.

Prefer `types.port` over `types.int` and `types.enum` over `types.str` when the
value set is closed — the module system then catches bad values at eval time.

### mkIf, mkMerge, mkDefault

```nix
config = lib.mkIf config.services.myservice.enable { };

config = lib.mkMerge [
  { always.present = true; }
  (lib.mkIf condition { conditional.value = true; })
];

# Overridable default
services.myservice.port = lib.mkDefault 8080;
```

`mkDefault` has priority 1000, `mkForce` has 50, a plain assignment has 100.
Use `mkForce` sparingly — it defeats the merge system.

## Package Development

### callPackage pattern

```nix
# pkgs/default.nix
{
  mypackage = pkgs.callPackage ./mypackage { };
  mytool = pkgs.callPackage ./mytool { };
}
```

### Package definition

```nix
{ lib
, buildGoModule
, fetchFromGitHub
}:

buildGoModule rec {
  pname = "mypackage";
  version = "1.0.0";

  src = fetchFromGitHub {
    owner = "owner";
    repo = "repo";
    rev = "v${version}";
    hash = "sha256-...";
  };

  vendorHash = "sha256-...";

  meta = with lib; {
    description = "Package description";
    homepage = "https://example.com";
    license = licenses.mit;
    maintainers = with maintainers; [ ];
    platforms = platforms.linux;
  };
}
```

### Overlays

```nix
{ inputs }:
{
  additions = final: _prev: import ../pkgs { pkgs = final; };

  modifications = final: prev: {
    somepackage = prev.somepackage.overrideAttrs (old: {
      version = "custom";
    });
  };
}
```

Use `final` for anything that should see later overlays, `prev` for the input
you are modifying. Referencing `final.foo` inside a definition of `foo` causes
infinite recursion.

## Flake Management

```bash
nix flake lock              # Lock dependencies
nix flake update            # Update all inputs
nix flake update nixpkgs    # Update one input
nix flake check             # Validate outputs
nix flake show              # List outputs
nix flake metadata          # Show inputs and revision
```

```nix
{
  description = "Flake description";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }: {
    nixosConfigurations = { };
    homeConfigurations = { };
    packages = { };
    devShells = { };
  };
}
```

## Home-Manager Patterns

### Session variables

```nix
home.sessionVariables = {
  EDITOR = "vim";
  BROWSER = "firefox";
};
```

### XDG config files

```nix
# Symlink a static file
xdg.configFile."myapp/config.yml".source = ./myapp-config.yml;

# Generate dynamically
xdg.configFile."myapp/generated.conf".text = ''
  setting1 = ${someValue}
'';

# Executable
xdg.configFile."bin/script.sh" = {
  source = ./script.sh;
  executable = true;
};
```

### User services

```nix
systemd.user.services.myservice = {
  Unit = {
    Description = "My Service";
    After = [ "network.target" ];
  };
  Service = {
    ExecStart = "${pkgs.mypackage}/bin/myservice";
    Restart = "on-failure";
  };
  Install.WantedBy = [ "default.target" ];
};
```

## Secrets Management

Never put plaintext secrets in a `.nix` file — everything in the Nix store is
world-readable. Use `agenix` or `sops-nix`, and reference the decrypted path at
runtime.

```nix
# secrets.nix
let
  user   = "ssh-ed25519 AAAAC3...";
  system = "ssh-ed25519 AAAAC3...";
in {
  "secret.age".publicKeys = [ user system ];
}
```

```nix
{
  age.secrets.mySecret = {
    file = ../secrets/mySecret.age;
    owner = "myuser";
    group = "mygroup";
  };

  # Pass the path, never the value
  services.myservice.passwordFile = config.age.secrets.mySecret.path;
}
```

```bash
agenix -e secrets/mySecret.age   # Edit or create
agenix -r                        # Re-key after adding a host
```

See `workflows/Secrets.md` for the full treatment.

## Safety and Testing

```bash
nixos-rebuild build --flake .#<hostname>       # Build only
nixos-rebuild dry-build --flake .#<hostname>   # Show what would change
nixos-rebuild test --flake .#<hostname>        # Activate, skip bootloader
nixos-rebuild switch --flake .#<hostname>      # Activate and make default
```

Always `build` before `switch` on a remote host. `test` is the safe middle
ground — a reboot recovers the previous generation.

### Rollback

```bash
nixos-rebuild list-generations
nixos-rebuild switch --rollback
nixos-rebuild switch --switch-generation <number>
```

Keep at least two or three recent generations. `nix-collect-garbage -d` deletes
all rollback targets — never run it as a first response to a full disk.

## Common Patterns

### Conditional imports

```nix
imports = [
  ./base.nix
] ++ lib.optionals (desktop != null) [
  ./desktop/${desktop}
];
```

### String interpolation

```nix
message = "Hello ${name}";

config = ''
  setting1 = ${value1}
'';

# Escaping $ inside an indented string
script = ''
  echo "Nix variable: ${nixVar}"
  echo "Shell variable: ''${shellVar}"
'';
```

### List and attrset operations

```nix
all      = list1 ++ list2;
filtered = lib.filter (x: x > 5) list;
doubled  = map (x: x * 2) list;

merged   = set1 // set2;                       # Shallow
merged   = lib.recursiveUpdate set1 set2;      # Deep
filtered = lib.filterAttrs (n: v: v != null) attrs;
mapped   = lib.mapAttrs (n: v: v * 2) attrs;
```

## Debugging

```nix
value = lib.traceVal someExpression;
value = lib.traceSeq "message" someExpression;
```

```bash
nix eval .#nixosConfigurations.<host>.config.services.nginx.enable
nix derivation show .#package
nix path-info -Sh .#package
nix why-depends .#package .#dependency
```

Add `--show-trace` to any failing evaluation for the full stack.

### Hash mismatch

```bash
nix-prefetch-github owner repo --rev <commit>
```

For Go modules, set `vendorHash = lib.fakeHash;` and read the correct hash out
of the build failure.

### Infinite recursion

Usually a `config` value read at the top level of a module, or an overlay
referring to `final.foo` while defining `foo`. Move the read inside `config` or
switch to `prev`.

## Performance

- Use binary caches; add project caches via `nixConfig.extra-substituters`
- Keep `flake.lock` updated but stable — churn invalidates caches
- Use `nix-direnv` so dev shells are cached and GC-rooted
- Avoid expensive list operations inside module option defaults

## Security

### Quick wins

```nix
{
  networking.firewall.enable = true;

  services.openssh.settings = {
    PermitRootLogin = "no";
    PasswordAuthentication = false;
  };

  system.autoUpgrade = {
    enable = true;
    allowReboot = false;
  };

  security.apparmor.enable = true;
}
```

### Checklist

- [ ] Firewall enabled, default deny
- [ ] SSH hardened — no root login, no password auth
- [ ] Secrets encrypted, referenced by path not value
- [ ] Automatic updates configured
- [ ] AppArmor or per-service systemd hardening enabled
- [ ] Audit logging on critical services
- [ ] Minimal package set on exposed hosts

See `workflows/Security.md` for the hardened profile and per-service
sandboxing.

## Resources

- NixOS Manual: https://nixos.org/manual/nixos/stable/
- Nixpkgs Manual: https://nixos.org/manual/nixpkgs/stable/
- Home-Manager Manual: https://nix-community.github.io/home-manager/
- Nix Pills: https://nixos.org/guides/nix-pills/
- Option search: https://search.nixos.org/options
