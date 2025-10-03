# ADR 0007: Use Cobra for CLI Framework

**Status:** Accepted

**Date:** 2024

## Context

We need a CLI framework to handle:
- Command-line argument parsing
- Flag definitions and validation
- Help text generation
- Subcommands (if needed in future)
- Configuration binding

Popular Go CLI frameworks:
- **Cobra** - Feature-rich, used by kubectl, Hugo, GitHub CLI
- **urfave/cli** - Simpler, less features
- **flag** (stdlib) - Basic, no advanced features
- **pflag** - POSIX-style flags, less structure

## Decision

We will use Cobra with Viper for CLI and configuration management.

## Rationale

**Pros:**
- **Industry standard** - Used by major projects (kubectl, Hugo)
- **Rich features** - Subcommands, persistent flags, auto-help
- **Viper integration** - Seamless config file support
- **Good documentation** - Well-documented and maintained
- **Extensible** - Easy to add subcommands later
- **Professional UX** - Consistent with popular tools

**Cons:**
- **Dependency overhead** - Larger than stdlib
- **Learning curve** - More complex than simple flag parsing
- **Opinionated** - Enforces certain patterns

**Alternatives considered:**
- **stdlib flag** - Too basic for our needs
- **urfave/cli** - Less features, smaller ecosystem
- **Custom implementation** - Reinventing the wheel

## Consequences

**Positive:**
- Professional CLI experience
- Easy to add features (version, completion, etc.)
- Config file support via Viper
- Familiar to Go developers
- Good error messages and help text

**Negative:**
- Additional dependency
- More code for simple use cases

## Features Enabled

- Automatic help generation
- Flag validation
- Environment variable binding
- Config file support (YAML)
- Future subcommands (e.g., `dbt2lookml validate`)
- Shell completion support

## Example

```go
var rootCmd = &cobra.Command{
    Use:   "dbt2lookml",
    Short: "Convert dbt models to LookML views",
    Run:   run,
}

rootCmd.Flags().StringVar(&manifestPath, "manifest-path", "", "Path to manifest.json")
```

## Notes

Cobra is the de facto standard for Go CLI tools and provides a professional user experience that users expect from modern CLI tools.
