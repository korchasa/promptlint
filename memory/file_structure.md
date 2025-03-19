# File Structure

```
promptlint/
├── .cursorrules         # Rules for Cursor editor
├── .env                 # Environment variables for API configuration
├── .gitignore           # File exclusion rules for repository
├── .goreleaser.yml      # GoReleaser configuration for creating releases
├── .github/             # GitHub Actions directory
│   └── workflows/       # CI/CD workflows
│       ├── build.yml    # Workflow for building and testing
│       └── release.yml  # Workflow for creating releases
├── .vscode/             # VSCode settings
├── Dockerfile           # Docker container configuration
├── bad_example.md       # Example of a bad prompt for testing
├── go.mod               # Module and dependency description
├── go.sum               # Dependency checksums
├── main.go              # Application entry point and all functionality (377 lines)
├── memory/              # Memory files for project context
│   ├── architecture.md  # Architecture description
│   ├── file_structure.md # File structure (this file)
│   ├── implementation.md # Implementation details
│   └── project.md       # Project overview
├── prompt_rules.yaml    # Rules for checking prompts (324 lines)
├── promptlint           # Compiled binary file
└── README.md            # Project documentation
```

## Key Files Description

### main.go (377 lines)
```go
// Main application file containing all functionality:
- Types and structures (PromptRule, Rules, Issue, LLMConfig)
- CLI argument processing (file, rules, version)
- Functions for reading from file and stdin
- LoadRules function for loading rules from YAML
- Report function for formatting results
- checkPromptWithLLM function for LLM API integration
  - System prompt preparation
  - HTTP request to LLM API
  - JSON response processing
- Main application logic and error handling
```

### prompt_rules.yaml (324 lines)
Contains a set of rules for checking prompts in YAML format:
- Rule name (name)
- Rule description (rule)
- Reason for the rule's importance (reason)
- Recommendation for fixing (fix)
- Examples of bad and good prompts

### bad_example.md
Example of a poorly formatted prompt used for testing the application:
- Lacks specificity and context
- Missing examples and formatting instructions
- Would trigger multiple rule violations when analyzed

### .env
Contains necessary environment variables for API configuration:
- PROMPTLINT_API_KEY - API key for LLM integration
- PROMPTLINT_API_ENDPOINT - URL of API endpoint
- PROMPTLINT_MODEL_NAME - Optional model name specification

### Dockerfile
```
- Multi-stage build process:
  - First stage uses golang:1.18-alpine to build the application
  - Second stage uses alpine:3.15 for a minimal runtime image
- Copies only necessary files (binary, rules, examples)
- Sets up default environment variables
- Runs as non-root user for security
```

### .github/workflows/build.yml
```
- GitHub Actions workflow for CI/CD:
  - Triggered on pushes to main and pull requests
  - Sets up Go environment and dependencies
  - Runs golangci-lint for code quality
  - Builds and tests the application
  - Creates Docker image (but doesn't push it)
```

### .github/workflows/release.yml
```
- GitHub Actions workflow for releases:
  - Triggered when a tag is pushed (v*)
  - Creates GitHub Release with binaries for multiple platforms
  - Builds and pushes Docker image to GitHub Container Registry
  - Generates checksums and changelog
```

### .goreleaser.yml
```
- Configuration for GoReleaser:
  - Builds for multiple platforms (Linux, Windows, macOS)
  - Creates archives with necessary files
  - Generates checksums and changelog
  - Configures release naming and versioning
```