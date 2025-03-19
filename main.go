package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	appName    = "promptlint"
	appVersion = "0.1.0"

	// ANSI color codes
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
)

//go:embed prompt_rules.yaml
var embeddedRules embed.FS

// PromptRule represents a rule structure for prompt checking
type PromptRule struct {
	Name        string `yaml:"name"`
	Rule        string `yaml:"rule"`
	Reason      string `yaml:"reason"`
	Fix         string `yaml:"fix"`
	BadExample  string `yaml:"badExample"`
	GoodExample string `yaml:"goodExample"`
	Pattern     string `yaml:"pattern,omitempty"`
	MinLength   int    `yaml:"minLength,omitempty"`
	MaxLength   int    `yaml:"maxLength,omitempty"`
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

// LLMConfig contains settings for LLM API interaction
type LLMConfig struct {
	APIKey      string
	APIEndpoint string
	ModelName   string
	Timeout     time.Duration
}

// LLMRequest represents a request to the LLM API
type LLMRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

// LLMResponse represents a response from the LLM API
type LLMResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

// Global variables for color configuration
var (
	useColorForProgress = true // Default value, will be updated in main()
)

// printProgress prints a progress message to stderr with color formatting
func printProgress(message string) {
	messageFormatted := message

	if useColorForProgress {
		appNameFormatted := fmt.Sprintf("%s%s%s%s", colorBlue, colorBold, appName, colorReset)

		// Add color to specific message types
		if strings.Contains(message, "Starting") || strings.Contains(message, "Finished") {
			messageFormatted = fmt.Sprintf("%s%s%s", colorGreen, message, colorReset)
		} else if strings.Contains(message, "Error") || strings.Contains(message, "Failed") {
			messageFormatted = fmt.Sprintf("%s%s%s", colorRed, message, colorReset)
		} else if strings.Contains(message, "Processing") || strings.Contains(message, "Validation") {
			messageFormatted = fmt.Sprintf("%s%s%s", colorYellow, message, colorReset)
		}

		fmt.Fprintf(os.Stderr, "[%s] %s\n", appNameFormatted, messageFormatted)
	} else {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", appName, message)
	}
}

// LoadRules loads rules from the embedded YAML file
func LoadRules() (*Rules, error) {
	printProgress("Loading built-in rules")
	data, err := embeddedRules.ReadFile("prompt_rules.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded rules file: %w", err)
	}

	var rules Rules
	printProgress("Parsing built-in rules")
	err = yaml.Unmarshal(data, &rules)
	if err != nil {
		return nil, fmt.Errorf("error parsing embedded YAML file: %w", err)
	}

	printProgress(fmt.Sprintf("Loaded %d built-in rules successfully", len(rules.PromptRules)))
	return &rules, nil
}

// isColorTerminal returns true if the terminal supports color output
func isColorTerminal() bool {
	// Check the NO_COLOR environment variable
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}

	// Check if stdout is a terminal
	fileInfo, _ := os.Stdout.Stat()
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		// Stdout is not a terminal (probably a pipe or a file)
		return false
	}

	// Check if TERM environment variable indicates color support
	term := os.Getenv("TERM")
	if term == "dumb" {
		return false
	}

	// Default to true for most modern terminals
	return true
}

// formatOriginalSnippet highlights the problematic parts of an example
func formatOriginalSnippet(snippet string, useColor bool) string {
	if !useColor {
		return snippet
	}
	return colorRed + snippet + colorReset
}

// formatFixedSnippet highlights the fixed parts of an example
func formatFixedSnippet(snippet string, useColor bool) string {
	if !useColor {
		return snippet
	}
	return colorGreen + snippet + colorReset
}

// Report formats the found issues into a report.
// If there are no issues, returns a message about the absence of problems.
func Report(issues []Issue, forceColor bool, noColor bool) string {
	useColor := false

	// Determine color usage based on flags and terminal capabilities
	if forceColor {
		useColor = true
	} else if noColor {
		useColor = false
	} else {
		useColor = isColorTerminal()
	}

	if len(issues) == 0 {
		if useColor {
			return fmt.Sprintf("%s%sNo issues found!%s\n", colorGreen, colorBold, colorReset)
		}
		return "No issues found!\n"
	}

	var sb strings.Builder

	// Output the number of issues found
	if useColor {
		sb.WriteString(fmt.Sprintf("Found %s%d issues%s:\n\n", colorBold, len(issues), colorReset))
	} else {
		sb.WriteString(fmt.Sprintf("Found %d issues:\n\n", len(issues)))
	}

	for i, issue := range issues {
		// Issue header with number and name
		if useColor {
			sb.WriteString(fmt.Sprintf("%s%s[Issue %d] %s%s\n", colorBlue, colorBold, i+1, issue.Description, colorReset))
		} else {
			sb.WriteString(fmt.Sprintf("[Issue %d] %s\n", i+1, issue.Description))
		}

		// Problem reason
		if useColor {
			sb.WriteString(fmt.Sprintf("%sReason:%s %s\n", colorBold, colorReset, issue.Reason))
		} else {
			sb.WriteString(fmt.Sprintf("Reason: %s\n", issue.Reason))
		}

		// Fix recommendation
		if useColor {
			sb.WriteString(fmt.Sprintf("%sFix:%s %s\n", colorBold, colorReset, issue.Fix))
		} else {
			sb.WriteString(fmt.Sprintf("Fix: %s\n", issue.Fix))
		}

		// Examples if available
		if issue.OriginalSnippet != "" && issue.FixedSnippet != "" {
			sb.WriteString("\n")

			// Format original snippet - display with indentation for multiline snippets
			if useColor {
				sb.WriteString(fmt.Sprintf("%sOriginal snippet:%s\n", colorBold, colorReset))
				sb.WriteString(formatOriginalSnippet(indentSnippet(issue.OriginalSnippet), useColor))
				sb.WriteString("\n")
			} else {
				sb.WriteString("Original snippet:\n")
				sb.WriteString(indentSnippet(issue.OriginalSnippet))
				sb.WriteString("\n")
			}

			// Format fixed snippet - display with indentation for multiline snippets
			if useColor {
				sb.WriteString(fmt.Sprintf("%sFixed snippet:%s\n", colorBold, colorReset))
				sb.WriteString(formatFixedSnippet(indentSnippet(issue.FixedSnippet), useColor))
				sb.WriteString("\n")
			} else {
				sb.WriteString("Fixed snippet:\n")
				sb.WriteString(indentSnippet(issue.FixedSnippet))
				sb.WriteString("\n")
			}
		}

		// Separator between issues
		if i < len(issues)-1 {
			sb.WriteString("\n" + strings.Repeat("─", 60) + "\n\n")
		}
	}

	return sb.String()
}

// indentSnippet adds indentation to each line of a multiline snippet
func indentSnippet(snippet string) string {
	lines := strings.Split(snippet, "\n")
	for i := range lines {
		lines[i] = "    " + lines[i]
	}
	return strings.Join(lines, "\n")
}

// errHandler processes errors and outputs a message to the user
func errHandler(err error, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
		os.Exit(1)
	}
}

// readFromFile reads file contents
func readFromFile(filePath string) (string, error) {
	printProgress(fmt.Sprintf("Reading prompt from file: %s", filePath))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	printProgress("File read successfully")
	return string(data), nil
}

// readFromStdin reads all input from stdin
func readFromStdin() (string, error) {
	printProgress("Reading prompt from stdin")
	scanner := bufio.NewScanner(os.Stdin)
	var sb strings.Builder

	for scanner.Scan() {
		sb.WriteString(scanner.Text())
		sb.WriteString("\n")
	}

	if scanner.Err() != nil {
		return "", fmt.Errorf("error reading from stdin: %w", scanner.Err())
	}

	printProgress("Stdin read successfully")
	return sb.String(), nil
}

// printUsage prints detailed usage instructions
func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage of %s:
  %s -file=<file>      Check prompt from the specified file
  %s -version          Show version information
  %s --force-color     Force colored output
  %s --no-color        Disable colored output
  %s                   Check prompt from stdin

Examples:
  %s -file=prompt.txt
  cat prompt.txt | %s
`, appName, appName, appName, appName, appName, appName, appName, appName)
}

// checkPromptWithLLM checks the prompt using LLM API
func checkPromptWithLLM(prompt string, rules *Rules, config *LLMConfig) ([]Issue, error) {
	printProgress("Starting LLM-based prompt validation")

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is missing, set PROMPTLINT_API_KEY")
	}

	if config.APIEndpoint == "" {
		return nil, fmt.Errorf("API endpoint is missing, set PROMPTLINT_API_ENDPOINT")
	}

	// Format rules as text for LLM
	printProgress("Preparing rules description for LLM")
	var rulesDescription strings.Builder
	rulesDescription.WriteString("List of prompt checking rules:\n\n")

	for i, rule := range rules.PromptRules {
		rulesDescription.WriteString(fmt.Sprintf("%d. Rule: %s\n", i+1, rule.Name))
		rulesDescription.WriteString(fmt.Sprintf("   Description: %s\n", rule.Rule))
		rulesDescription.WriteString(fmt.Sprintf("   Reason: %s\n", rule.Reason))
		if rule.BadExample != "" {
			rulesDescription.WriteString(fmt.Sprintf("   Original snippet: %s\n", rule.BadExample))
		}
		if rule.GoodExample != "" {
			rulesDescription.WriteString(fmt.Sprintf("   Fixed snippet: %s\n", rule.GoodExample))
		}
		rulesDescription.WriteString("\n")
	}

	// Prepare request to LLM API
	printProgress("Creating system message")
	systemMessage := `You are a prompt evaluation expert. Your task is to analyze a prompt and determine if it follows the provided rules.

Analyze the prompt against each rule and identify violations. The rules are provided in a separate message.

Use the find_prompt_issues tool to return the issues found in the prompt. If there are no issues, return an empty array.`

	// Define a tool for finding prompt issues
	printProgress("Configuring tools for structured response")
	tools := []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "find_prompt_issues",
				"description": "Reports issues found in a prompt based on predefined rules",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"issues": map[string]interface{}{
							"type":        "array",
							"description": "List of issues found in the prompt",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type":        "string",
										"description": "Name of the violated rule",
									},
									"description": map[string]interface{}{
										"type":        "string",
										"description": "Description of the problem",
									},
									"reason": map[string]interface{}{
										"type":        "string",
										"description": "Why this is a problem (from the rules)",
									},
									"fix": map[string]interface{}{
										"type":        "string",
										"description": "Recommendation for fixing",
									},
									"originalSnippet": map[string]interface{}{
										"type":        "string",
										"description": "Problematic part of the prompt (if applicable)",
									},
									"fixedSnippet": map[string]interface{}{
										"type":        "string",
										"description": "Improved version of the snippet (if applicable)",
									},
								},
								"required": []string{"name", "description", "reason", "fix", "originalSnippet", "fixedSnippet"},
							},
						},
					},
					"required": []string{"issues"},
				},
			},
		},
	}

	printProgress("Building request payload")
	requestBody := map[string]interface{}{
		"model": config.ModelName,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemMessage,
			},
			{
				"role":    "user",
				"content": rulesDescription.String(),
			},
			{
				"role":    "user",
				"content": "Analyze the following prompt against the specified rules:\n\n" + prompt,
			},
		},
		"tools": tools,
		"tool_choice": map[string]interface{}{
			"type": "function",
			"function": map[string]string{
				"name": "find_prompt_issues",
			},
		},
	}

	printProgress("Serializing request to JSON")
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("request serialization error: %w", err)
	}

	// Prepare HTTP request
	printProgress(fmt.Sprintf("Setting up HTTP client with timeout %v", config.Timeout))
	client := &http.Client{
		Timeout: config.Timeout,
	}

	printProgress(fmt.Sprintf("Creating HTTP request to %s", config.APIEndpoint))
	req, err := http.NewRequest("POST", config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// Execute request
	printProgress("Sending request to LLM API")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	printProgress(fmt.Sprintf("Received response with status code: %d", resp.StatusCode))
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Process response
	printProgress("Decoding API response")
	var responseData map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&responseData); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Extract tool call results
	printProgress("Extracting tool call results")
	var issues []Issue

	// Navigate through the response structure to extract tool calls
	if choices, ok := responseData["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if toolCalls, ok := message["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
					// We found tool calls, extract the function arguments
					for _, tc := range toolCalls {
						if toolCall, ok := tc.(map[string]interface{}); ok {
							if function, ok := toolCall["function"].(map[string]interface{}); ok {
								if args, ok := function["arguments"].(string); ok {
									// Parse the arguments as JSON
									var toolResponse map[string]interface{}
									if err := json.Unmarshal([]byte(args), &toolResponse); err != nil {
										return nil, fmt.Errorf("error parsing tool response: %w", err)
									}

									// Extract issues from the tool response
									if issuesData, ok := toolResponse["issues"].([]interface{}); ok {
										printProgress(fmt.Sprintf("Processing %d issues found by LLM", len(issuesData)))
										for _, issueData := range issuesData {
											if issueMap, ok := issueData.(map[string]interface{}); ok {
												issue := Issue{
													RuleName:        getStringValue(issueMap, "name"),
													Description:     getStringValue(issueMap, "description"),
													Reason:          getStringValue(issueMap, "reason"),
													Fix:             getStringValue(issueMap, "fix"),
													OriginalSnippet: getStringValue(issueMap, "originalSnippet"),
													FixedSnippet:    getStringValue(issueMap, "fixedSnippet"),
												}
												issues = append(issues, issue)
											}
										}
									}
								}
							}
						}
					}
				} else {
					printProgress("No tool calls found in response, trying legacy format")
					// Fallback to content-based response (older model or API version)
					if content, ok := message["content"].(string); ok && content != "" {
						var legacyIssues []map[string]string
						// Try to parse JSON array from the content
						jsonStartIdx := strings.Index(content, "[")
						jsonEndIdx := strings.LastIndex(content, "]")

						if jsonStartIdx >= 0 && jsonEndIdx > jsonStartIdx {
							jsonContent := content[jsonStartIdx : jsonEndIdx+1]
							if err := json.Unmarshal([]byte(jsonContent), &legacyIssues); err != nil {
								return nil, fmt.Errorf("error parsing legacy response: %w", err)
							}
						} else {
							// Try to parse the entire content
							if err := json.Unmarshal([]byte(content), &legacyIssues); err != nil {
								return nil, fmt.Errorf("failed to parse legacy response as JSON: %w\nResponse: %s", err, content)
							}
						}

						// Convert legacy format to Issue structure
						for _, issueMap := range legacyIssues {
							issue := Issue{
								RuleName:        issueMap["name"],
								Description:     issueMap["description"],
								Reason:          issueMap["reason"],
								Fix:             issueMap["fix"],
								OriginalSnippet: issueMap["originalSnippet"],
								FixedSnippet:    issueMap["fixedSnippet"],
							}
							issues = append(issues, issue)
						}
					}
				}
			}
		}
	}

	printProgress("Validation completed successfully")
	return issues, nil
}

// getStringValue safely extracts a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// setupLLMConfig configures the LLM API settings
func setupLLMConfig() (LLMConfig, error) {
	printProgress("Setting up LLM API configuration")

	apiKey := os.Getenv("PROMPTLINT_API_KEY")
	if apiKey == "" {
		return LLMConfig{}, fmt.Errorf("API key not specified, set PROMPTLINT_API_KEY environment variable")
	}

	apiEndpoint := os.Getenv("PROMPTLINT_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.openai.com/v1/chat/completions" // Default value
		printProgress("Using default API endpoint: " + apiEndpoint)
	}

	modelName := os.Getenv("PROMPTLINT_MODEL_NAME")
	if modelName == "" {
		modelName = "o3-mini" // Default value
		printProgress("Using default model: " + modelName)
	}

	timeout := 300 * time.Second
	printProgress("Configuration completed")

	return LLMConfig{
		APIKey:      apiKey,
		APIEndpoint: apiEndpoint,
		ModelName:   modelName,
		Timeout:     timeout,
	}, nil
}

func main() {
	printProgress("Starting " + appName + " v" + appVersion)

	// Parse command line arguments
	fileFlag := flag.String("file", "", "Path to file with prompt")
	versionFlag := flag.Bool("version", false, "Show version information")
	forceColorFlag := flag.Bool("force-color", false, "Force colored output even when stdout is not a terminal")
	noColorFlag := flag.Bool("no-color", false, "Disable colored output")

	printProgress("Parsing command line arguments")
	flag.Parse()

	// Configure color settings based on flags
	if *forceColorFlag {
		useColorForProgress = true
	} else if *noColorFlag {
		useColorForProgress = false
	} else {
		useColorForProgress = isColorTerminal()
	}

	// Display version information
	if *versionFlag {
		fmt.Printf("%s version %s\n", appName, appVersion)
		return
	}

	// Load built-in rules
	rules, err := LoadRules()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load built-in rules: %v\n", err)
		os.Exit(1)
		return
	}

	// Check if there's data on stdin
	printProgress("Checking input method")
	stdinInfo, _ := os.Stdin.Stat()
	hasStdin := (stdinInfo.Mode() & os.ModeCharDevice) == 0

	// Check if application is launched correctly
	if *fileFlag == "" && !hasStdin {
		fmt.Fprintf(os.Stderr, "Error: No input provided. Please specify a file or pipe data to stdin.\n\n")
		printUsage()
		os.Exit(1)
		return
	}

	// Read prompt from file or stdin
	var input string
	if *fileFlag != "" {
		input, err = readFromFile(*fileFlag)
		errHandler(err, "Error reading file")
	} else {
		input, err = readFromStdin()
		errHandler(err, "Error reading from stdin")
	}

	// Check if input is empty
	if strings.TrimSpace(input) == "" {
		fmt.Fprintf(os.Stderr, "Error: Empty input. Please provide a prompt to check.\n\n")
		printUsage()
		os.Exit(1)
		return
	}

	// Setup LLM configuration
	llmConfig, err := setupLLMConfig()
	errHandler(err, "Error setting up LLM API")

	// Check prompt using only LLM API
	printProgress("Starting prompt validation process")
	issues, err := checkPromptWithLLM(input, rules, &llmConfig)
	errHandler(err, "Error checking prompt with LLM API")

	// Format and output report
	printProgress("Generating final report")
	report := Report(issues, *forceColorFlag, *noColorFlag)
	fmt.Println(report)

	printProgress("Finished")
}
