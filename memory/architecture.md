# Architecture

## Overview
```
Input → LLM API → Reporter → Output
```

Simplified pipeline architecture with exclusive use of LLM API for prompt validation.

## Components
| Component | Functionality | Implementation |
|-----------|------------------|------------|
| **Reporter** | Formatting of results | `main.go:Report()` |
| **LLM Integration** | Checking prompts via API using rules from YAML | `main.go:checkPromptWithLLM()` |
| **Rules Engine** | Loading and storing YAML rules | `main.go:LoadRules()` |

## Key Design Patterns
- **Data processing pipeline**: step-by-step data transformation
- **Delegation**: transferring prompt analysis task to external LLM API
- **Monolithic design**: all functionality in a single file for simplicity

## Processing Flow
1. **Rules loading**: Parsing the YAML rules file
2. **Reading prompt**: From file or stdin
3. **LLM API setup**: Checking and setting environment variables
4. **Validation**: Sending prompt and rules to LLM API
5. **Report formatting**: Converting found issues into readable output

## LLM Integration
1. Creating a textual representation of rules from the YAML file
2. Preparing a system prompt for analysis based on rules
3. Sending rules and prompt to LLM API for compliance checking
4. Processing and parsing the JSON response
5. Converting results into Issue objects
6. Passing to Reporter for formatting

## Error Handling Strategy
1. Error detection at the point of occurrence
2. Enriching with context via `fmt.Errorf("context: %w", err)`
3. Centralized error handling through `errHandler()`
4. Clear API error messages for the user
5. Program termination on any LLM API error

## Validation Approach
- Complete delegation of prompt validation to external LLM API
- Sending all rules from YAML file to LLM
- Using structured JSON format for results
- Strict requirements for API availability and correct operation

## Testing Approach
- Integration tests with real LLM API
- Mock tests for LLM API emulation in automated tests
- Testing handling of various API errors
- Golden file testing for output formatting