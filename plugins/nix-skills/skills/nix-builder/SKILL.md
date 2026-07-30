---
name: nix-builder
description: >
  Builds and maintains NixOS host configurations in an existing Nix flake
  repository. Use when adding a service to a host, adding a new host to a
  flake, changing an existing host's configuration, or reviewing generated Nix
  config for correctness. Triggers on requests like "add jellyfin to the nix
  config", "add a new host to the flake", "generate the config for <host>",
  "install <software> on <host>", "wire up a reverse proxy vhost for
  <service>", or "check my host config". Discovers the fleet, layout, and stack
  by reading the repository — it does not assume a fixed set of hosts. For
  general Nix questions, packaging, or debugging build errors, use nix-expert
  instead.
---

# Nix Builder

Creates and maintains NixOS host configurations inside an existing flake
repository.

This skill derives everything it needs about the fleet from the repository
itself. It has no built-in knowledge of your hosts, services, or conventions,
and must not invent any.

---

## Discovery Protocol

Run this before writing a single line of Nix. Facts come from the repository;
prose documents supply only intent.

### Step 1 — Read the flake (authoritative)

| Source | Yields |
|---|---|
| `flake.nix` → `nixosConfigurations` | The complete host list and how hosts are constructed |
| `flake.nix` → `inputs` | nixpkgs channel, home-manager, agenix/sops-nix, hardware modules |
| `flake.lock` | The exact nixpkgs revision the repo builds against |
| `hosts/*/` (or equivalent) | Per-host config, profiles in use, hardware layout |
| `modules/services/*.nix` | Which services already exist and the house style for writing them |
| `modules/common/`, `modules/profiles/` | The baseline every host inherits |
| `secrets/secrets.nix` | Secrets backend, key names, per-host key assignments |
| `lib/` | Helper functions such as `mkHost` that must be used rather than bypassed |

This is ground truth. It builds, so it cannot be stale.

### Step 2 — Read prose docs (intent only)

Check, in order, and use whichever exist:

- `CLAUDE.md` / `AGENTS.md` in the target repository — conventions, commands, guardrails
- `README.md` — what the fleet is for, deployment workflow
- Any architecture or planning doc the user points at

Use these for *why* — a host's purpose, planned changes, naming conventions,
rules Nix cannot express. Never take a host list or service inventory from prose
when the flake disagrees. **The flake wins.**

### Step 3 — Derive the stack

Do not assume. Determine from what Step 1 found:

| Concern | Detect by |
|---|---|
| nixpkgs version | `flake.nix` input URL + `flake.lock` — pin all option names to this release |
| Secrets | `agenix` vs `sops-nix` vs neither, from inputs and `secrets/` |
| Reverse proxy | Caddy vs nginx vs Traefik vs none, from `modules/services/` |
| TLS | ACME provider, internal PKI, or manual certs |
| Storage | ZFS, btrfs, ext4, plus any NFS/CIFS mounts |
| Containers | `virtualisation.oci-containers` backend, if used at all |
| Admin user | The user account declared in the common module |

### Step 4 — Ask only for gaps

If something material is neither in the repo nor in the docs, ask. Do not guess
at hostnames, IP addresses, storage paths, or domain names.

---

## Commands

### add-service

Add a service to an existing host.

```
add jellyfin to media-host
install prometheus node exporter on every host
```

1. Run the discovery protocol
2. Check whether a native NixOS module exists — prefer it over a container
3. Read two or three existing files in `modules/services/` and match their style
4. Write the service module, wire it into the host's imports
5. Add reverse proxy vhost, firewall rules, and secrets declarations as the repo's conventions require
6. Run the validation pass
7. Emit the annotation block

### add-host

Add a new host to the flake.

```
add a new host called <name> for <purpose>
```

1. Run the discovery protocol
2. Read an existing host of the same shape as a template
3. Register the host in `flake.nix` using the repo's own helper
4. Create the host directory, declaring the correct profile
5. Stub `hardware-configuration.nix` — flag clearly that the real one comes from `nixos-generate-config` on the actual machine
6. Add the host's key to the secrets config if one exists
7. Run the validation pass

### review

Validate existing config without changing it.

```
check the config for <host>
review my flake for problems
```

Run the discovery protocol and the validation pass, then report findings. Make
no edits.

---

## Writing Rules

- **Match the repo, don't impose a layout.** Conform to whatever structure
  discovery found. If it differs from anything in these instructions, the repo
  wins.
- **Use the repo's helpers.** If `lib/mkHost.nix` exists, use it rather than
  calling `nixosSystem` directly.
- **Pin to the detected release.** Option shapes change between NixOS releases.
  Verify against the version in `flake.lock`, not the newest you know of.
- **Prefer native modules** over `oci-containers`. Reach for a container only
  when nixpkgs has no module, and say why in the annotation.
- **Never generate `flake.lock`.** It comes from `nix flake update`.
- **Never write a plaintext secret.** The Nix store is world-readable. Declare
  the secret and reference its path.
- **Never hardcode an IP** where a hostname works.
- **Prefer targeted edits** to regenerating whole files.

---

## Validation Pass

Work through this before presenting output. Report failures in the annotation
block rather than silently skipping them.

### Syntax and structure
- [ ] Every `.nix` file is a syntactically valid Nix expression
- [ ] Every path in an `imports` list exists
- [ ] No circular imports
- [ ] The host's `default.nix` imports every module it needs

### Release correctness
- [ ] All option names are valid for the nixpkgs release in `flake.lock`
- [ ] No options taken from unstable that don't exist in that release
- [ ] Option *shapes* verified, not just names — several change across releases
      (`services.postgresql.ensureUsers` is a recurring example)
- [ ] `virtualisation.oci-containers.backend` declared if containers are used

### Secrets
- [ ] Every referenced secret has a declaration in the secrets config
- [ ] Secret file paths are correct relative to the flake root
- [ ] Owner and group match the service user that reads the file
- [ ] The new host's key is present if secrets are host-scoped
- [ ] No secret value appears in any `.nix` file

### Services
- [ ] Each externally reachable service has a reverse proxy vhost, if the repo uses one
- [ ] TLS configured the way the rest of the repo does it
- [ ] Network mounts declared with `x-systemd.automount` and a `soft` timeout
- [ ] Referenced datasets or storage paths exist in the declared layout
- [ ] Container volume mounts use absolute paths
- [ ] State directories persist across rebuilds

### Host config
- [ ] Correct profile declared for the hardware type
- [ ] Admin user declared, matching the repo's convention
- [ ] `services.openssh.settings.PasswordAuthentication = false`
- [ ] Firewall opens only the ports the declared services need
- [ ] Host registered in `flake.nix`

---

## Output Format

Print each file preceded by its repo-relative path:

```
=== hosts/media-host/default.nix ===
{ config, pkgs, ... }:
{
  imports = [
    ...
  ];
}
```

Then the annotation block:

```
--- ANNOTATION ---

DISCOVERED CONTEXT:
- Hosts: {from flake.nix nixosConfigurations}
- nixpkgs: {release} @ {lock revision}
- Stack: {secrets backend}, {reverse proxy}, {storage}
- Layout: {structure found}

IMPLEMENTATION DECISIONS:
- {service}: native module / oci-container because {reason}

SECRETS REQUIRED:
- secrets/{name}.age — used by {service} for {purpose}
  Create with: agenix -e secrets/{name}.age

NOT IMPLEMENTED:
- {item}: {reason} — action required before production use

VALIDATION RESULTS:
- PASS / FAIL {check}: {detail if failed}

OPEN QUESTIONS:
- {anything uncertain the user should confirm}

NEXT STEPS:
- nixos-rebuild build --flake .#{host}    # verify before switching
```

Never claim a config builds. You cannot run `nixos-rebuild` against the target
hardware — say what the user should run.

---

## Reference Files

Read on demand. Do not load these upfront.

| File | When to read |
|---|---|
| `references/services.md` | Before implementing any service — native module patterns and option shapes |
| `references/zfs.md` | When declaring ZFS pools, datasets, or dataset-backed state dirs |
| `references/secrets.md` | When declaring agenix secrets or internal PKI integration |
| `references/caddy.md` | When generating Caddy vhosts, including internal-PKI TLS |

Examples in these files use role-based placeholder hostnames
(`media-host`, `gpu-host`, `service-host`). They illustrate patterns — replace
them with the real hosts found during discovery.
