# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added changelog command with security controls
- Added code quality check command with parallel sub-task strategy
- Added clean command for removing redundant comments
- Added GitHub Super-Linter workflow
- Added MCP Security Scanner for preventing sensitive data leaks

### Changed

- Improved project linting and code quality standards
- Enhanced security scanning capabilities

### Fixed

- Fixed markdown linting issues (line length violations)
- Fixed general linter issues across project files
- Removed superfluous comments from codebase

### Security

- Implemented MCP Security Scanner to prevent sensitive data from being sent to external services
