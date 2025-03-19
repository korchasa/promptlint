# Implementation

## Repository Structure
```
promptlint/
├── main.go             # Entry point, CLI interface, all application logic
├── prompt_rules.yaml   # Rules in YAML format (embedded in binary at build time)
├── .env                # Environment variables for API configuration
├── bad_example.md      # Example of a bad prompt for testing
├── Dockerfile          # Docker container configuration
├── .goreleaser.yml     # GoReleaser configuration for releases
├── .github/            # GitHub Actions workflows
│   └── workflows/
│       ├── build.yml   # CI workflow for building and testing
│       └── release.yml # Release workflow for creating releases
└── memory/             # Project documentation
```

## CLI Flags
| Flag | Type | Description |
|------|-----|----------|
| `-file=<path>` | string | Path to prompt file |
| `-version` | bool | Print program version |
| `--force-color` | bool | Force colored output even when stdout is not a terminal |
| `--no-color` | bool | Disable colored output |

## Execution Flow
1. Parsing command line arguments
2. Loading built-in rules (embedded at compile time)
3. Reading prompt from file or stdin
4. Checking and configuring required LLM API environment variables
5. Checking prompt via LLM API with tools for structured validation
6. Formatting and displaying report on found issues

## Core Interfaces & Types

### Types and Structures
```go
// PromptRule represents a rule for checking prompts
type PromptRule struct {
    Name        string `yaml:"name"`
    Rule        string `yaml:"rule"`
    Reason      string `yaml:"reason"`
    Fix         string `yaml:"fix"`
    BadExample  string `yaml:"badExample"`
    GoodExample string `yaml:"goodExample"`
}

// Rules contains a list of rules for linting
type Rules struct {
    PromptRules []PromptRule `yaml:"prompt_rules"`
}

// Issue represents a problem found during linting
type Issue struct {
    RuleName        string
    Description     string
    Reason          string
    Fix             string
    OriginalSnippet string
    FixedSnippet    string
}

// LLMConfig settings for LLM API interaction
type LLMConfig struct {
    APIKey      string
    APIEndpoint string
    ModelName   string
    Timeout     time.Duration
}
```

### Core Functions
```go
// LoadRules loads rules from the embedded YAML file
func LoadRules() (*Rules, error)

// Report formats the found issues into a report
func Report(issues []Issue, forceColor bool, noColor bool) string

// isColorTerminal returns true if the terminal supports color output
func isColorTerminal() bool

// formatOriginalSnippet highlights the problematic parts of an example
func formatOriginalSnippet(snippet string, useColor bool) string

// formatFixedSnippet highlights the good parts of an example
func formatFixedSnippet(snippet string, useColor bool) string

// indentSnippet adds indentation to each line of a multiline snippet
func indentSnippet(snippet string) string

// checkPromptWithLLM checks the prompt using LLM API
func checkPromptWithLLM(prompt string, rules *Rules, config *LLMConfig) ([]Issue, error)
// - Creates a textual representation of rules from YAML file
// - Configures tools for structured response processing
// - Sends rules and prompt to LLM for analysis
// - Processes tool call results into Issue structure

// getStringValue safely extracts a string value from a map
func getStringValue(m map[string]interface{}, key string) string

// printProgress prints a progress message to stderr
func printProgress(message string)
// - Displays formatted progress messages at key stages of execution
// - Output format: [appName] message
```

## Embedded Rules
The application now embeds the rules YAML file at compile time using Go's `embed` package:
```go
//go:embed prompt_rules.yaml
var embeddedRules embed.FS
```

This approach eliminates the need for distributing the rules file alongside the binary and ensures consistent rule application across all environments.

## LLM API Integration with Tools
The application uses OpenAI's function calling capabilities to get structured responses:

- **Tool Definition**: A `find_prompt_issues` tool is defined with a JSON schema that specifies the expected response format
- **Structure Enforcement**: The schema guarantees consistent response structure with proper typing
- **Forced Usage**: The `tool_choice` parameter forces the model to use the defined tool
- **Fallback Mechanism**: Includes a fallback to legacy content-based parsing for older API versions or models
- **Reliable Processing**: Structured responses reduce parsing errors and inconsistencies

```go
tools := []map[string]interface{}{
    {
        "type": "function",
        "function": map[string]interface{}{
            "name": "find_prompt_issues",
            "description": "Reports issues found in a prompt based on predefined rules",
            "parameters": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "issues": map[string]interface{}{
                        "type": "array",
                        "items": map[string]interface{}{
                            // Issue schema definition
                        },
                    },
                },
            },
        },
    },
}
```

## Report Formatting
The application generates structured and colorized reports for better readability:

- **Structured Layout**:
  - Clear headers, sections, and separators with proper indentation
  - Multiline snippets displayed with 4-space indentation for better readability
  - Original and fixed snippets displayed on separate lines after their headers

- **Color Highlighting**:
  - Green for positive elements (fixed snippets, "No issues found" message)
  - Red for problem areas (original snippets)
  - Blue for issue numbers and titles
  - Bold for section headers

- **Color Control Options**:
  - `--force-color`: Override auto-detection and always use colors
  - `--no-color`: Disable colors regardless of terminal capabilities
  - Environment variable `NO_COLOR`: Disable colors via environment setting

- **Color Detection Logic**:
  1. If `--force-color` flag is set → enable colors
  2. If `--no-color` flag is set → disable colors
  3. Otherwise, check terminal capabilities:
     - Verify if stdout is a terminal
     - Check `NO_COLOR` environment variable
     - Inspect `TERM` environment variable value

## Environment Variables
| Variable | Description | Usage |
|------------|----------|------------|
| `PROMPTLINT_API_KEY` | API key for LLM | Required |
| `PROMPTLINT_API_ENDPOINT` | URL of API endpoint | Optional, default "https://api.openai.com/v1/chat/completions" |
| `PROMPTLINT_MODEL_NAME` | LLM model name | Optional, default "o3-mini" |
| `NO_COLOR` | Disable colorized output | Optional, any value disables colors |

## Progress Reporting
The application displays colorized progress messages at each stage of execution:
- Application startup & configuration (green)
- Rules loading and parsing
- Prompt reading (file/stdin)
- LLM API configuration
- Request preparation and serialization
- API request sending & response handling
- Processing & validation steps (yellow)
- Error and failure messages (red)
- Report generation
- Completion message (green)

Progress message features:
- Color-coded by message type for better visual distinction
- Application name highlighted in bold blue
- Consistent format: `[promptlint] message`
- Colors respect the same flags as report formatting
- All progress messages are sent to stderr

## Testing
For testing the application, use the included bad prompt example:
```bash
./promptlint -file=bad_example.md
```

The `.env` file in the root directory contains the necessary environment variables that will be automatically loaded.

## Error Handling
- When API key is missing or invalid — program termination
- When API endpoint is not available — program termination
- During HTTP request failures — program termination with detailed information
- When response parsing problems occur — program termination

## Tech Stack
- Go 1.18+
- gopkg.in/yaml.v3
- Go standard library (net/http, encoding/json, embed, etc.)

## Deployment Options

### Docker
- Multi-stage build for minimal image size
- Alpine-based lightweight container
- Environment variable configuration
- Non-root user for security

### CI/CD with GitHub Actions
- Automated testing and linting
- Docker image building
- Release automation with GoReleaser
- Binary distribution for multiple platforms
- GitHub Container Registry integration