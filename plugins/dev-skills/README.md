# Development Skills Plugin

Expert development skills for Go, documentation, architecture, and project management.

## Skills Included

### `/go-implementor`

Expert Go software engineer for implementing production-grade backend services with idiomatic Go patterns, testing, and observability.

**Use when**: Implementing Go code following best practices.

### `/go-review`

Senior Go code reviewer for comprehensive code reviews focused on correctness, security, performance, and maintainability.

**Use when**: Reviewing Go code for bugs, best practices, and production readiness.

### `/document`

Technical documentation expert for creating clear, comprehensive documentation including API docs (OpenAPI), ADRs, system architecture docs, developer guides, and runbooks.

**Use when**: Creating or improving technical documentation.

### `/mentor`

Senior Staff Engineer mentor for architecture, design decisions, technical mentorship, and career guidance.

**Use when**: Seeking senior-level perspective on system design, code review, technical strategy, or career growth.

### `/manager`

Engineering manager agent for orchestrating complex software development tasks by coordinating specialized sub-agents and managing parallel work streams.

**Use when**: Large initiatives requiring coordination across multiple specialists.

### `/specify`

Software specification expert for transforming high-level designs into detailed, implementable technical specifications.

**Use when**: Converting requirements or designs into precise specs that developers can implement directly.

### `/taskify`

Task decomposition expert for breaking technical specifications into atomic, implementable tasks with dependencies and priorities.

**Use when**: Converting specs into actionable task lists for development teams.

## Installation

Install via Claude Code marketplace:

```bash
claude code plugins install dev-skills
```

Or install from this repository:

```bash
claude code plugins install github:rikdc/claude_code_template/dev-skills
```

## Usage Examples

```bash
# Implement a new Go service
/go-implementor Implement a user authentication service

# Review Go code
/go-review Review the authentication middleware for security issues

# Create API documentation
/document Create OpenAPI spec for the user service

# Get architecture guidance
/mentor Should I use microservices or monolith for this project?

# Break down a specification
/taskify Convert the authentication spec into implementable tasks

# Create detailed specification
/specify Create detailed spec for user authentication flow

# Orchestrate complex project
/manager Implement user authentication with tests and docs
```

## License

MIT
