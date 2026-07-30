# Nix Skills Plugin

Expert skills for Nix, NixOS, and home-manager — one for authoring host
configurations, one for advisory guidance.

## Skills Included

### `/nix-builder`

Builds and maintains NixOS host configurations in an existing Nix flake
repository. Adds services to hosts, registers new hosts, and validates
generated config.

**Use when**: adding a service to a host, adding a host to a flake, changing an
existing host's configuration, or reviewing config for correctness.

**Key property — no hardcoded fleet.** The skill discovers your hosts, layout,
and stack by reading the repository at runtime rather than carrying built-in
assumptions:

1. **The flake is authoritative.** `nixosConfigurations` gives the host list,
   `inputs` and `flake.lock` give the nixpkgs release, `hosts/` and `modules/`
   give the layout and house style, `secrets/` gives the secrets backend.
2. **Prose docs supply intent only.** `CLAUDE.md`, `AGENTS.md`, and `README.md`
   in the target repo contribute conventions and purpose. Where prose and flake
   disagree, the flake wins — it builds, so it cannot be stale.
3. **Gaps become questions.** Anything not in the repo or docs gets asked about
   rather than guessed.

This means the skill stops going stale when you rename or add a host, and it
works on any flake, not just one.

### `/nix-expert`

Advisory guidance on Nix expressions, modules, derivations, overlays, flake
management, home-manager, secrets, and NixOS hardening.

**Use when**: writing or reviewing Nix expressions, authoring packages,
managing flake inputs, or debugging evaluation and build failures.

Routes to nine focused reference files, each a task-oriented procedure rather
than passive background:

| Reference | Covers |
|---|---|
| [Build](skills/nix-expert/references/Build.md) | `nix build`, `nixos-rebuild build`, build failures |
| [Debug](skills/nix-expert/references/Debug.md) | Evaluation errors, infinite recursion, `--show-trace` |
| [Develop](skills/nix-expert/references/Develop.md) | `devShell`, `nix develop`, direnv integration |
| [Deploy](skills/nix-expert/references/Deploy.md) | `nixos-rebuild switch`, remote deployment, rollback |
| [Package](skills/nix-expert/references/Package.md) | Derivations, `buildGoModule`, overlays |
| [Flakes](skills/nix-expert/references/Flakes.md) | Inputs, `flake.lock`, outputs schema |
| [Secrets](skills/nix-expert/references/Secrets.md) | agenix, sops-nix, age encryption |
| [Security](skills/nix-expert/references/Security.md) | Hardening, AppArmor, firewall, systemd sandboxing |
| [Troubleshoot](skills/nix-expert/references/Troubleshoot.md) | Hash mismatches, common errors |

`/nix-expert`'s `SKILL.md` also carries a five-rung safety ladder for
`nixos-rebuild` (`dry-build → build → dry-activate → test → switch`), with the
failure recovery for each rung.

## Choosing Between Them

| Task | Skill |
|---|---|
| "Add Jellyfin to media-host" | `/nix-builder` |
| "Add a new host to my flake" | `/nix-builder` |
| "Review my host config" | `/nix-builder` |
| "Why is this expression infinitely recursing?" | `/nix-expert` |
| "Package this Go app for Nix" | `/nix-expert` |
| "How do overlays work?" | `/nix-expert` |

Rule of thumb: changing `.nix` files for a specific host is `nix-builder`;
everything else is `nix-expert`.

## Installation

Install via the Claude Code marketplace:

```bash
claude code plugins install nix-skills
```

Or install from this repository:

```bash
claude code plugins install github:rikdc/ai-skills/nix-skills
```

## Usage Examples

```bash
# Add a service to an existing host
/nix-builder add jellyfin to media-host

# Register a new host
/nix-builder add a new host called build-box for CI runners

# Validate without changing anything
/nix-builder review the config for gpu-host

# Debug an evaluation failure
/nix-expert Why does this module cause infinite recursion?

# Package an application
/nix-expert Package this Go application for Nix

# Harden a host
/nix-expert How should I harden an internet-facing NixOS host?
```

## Assumptions and Portability

`nix-builder` detects rather than assumes, but it is opinionated about
*practice*:

- Prefers native NixOS modules over `oci-containers`, and says why when it
  reaches for a container
- Requires secrets to be encrypted (agenix or sops-nix) and referenced by path,
  never by value
- Will not generate `flake.lock`
- Will not claim a config builds — it tells you the `nixos-rebuild build`
  command to run

The bundled reference files (`services.md`, `zfs.md`, `secrets.md`,
`caddy.md`) use role-based placeholder hostnames — `media-host`, `gpu-host`,
`service-host` — and their option shapes were verified against **NixOS 24.11**.
Confirm option names against the nixpkgs revision your repository actually
pins.

## License

Mozilla Public License 2.0 — see [LICENSE](../../LICENSE).
