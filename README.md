# PromptLint

> **Warning**
> This is a joke. It shouldn't be taken seriously.

A utility for checking and validating prompts used with language models.

## How It Works

The tool follows a simple pipeline architecture:

```
Input → Parser → Linter → Reporter → Output
```

1. **Input Processing**:
   - Reads prompt text from a file (using the `-file` flag) or from stdin
   - Correctly handles stdin (distinguishes between direct terminal input and redirections)

2. **Prompt Checking**:
   - Checks prompts for errors and potential issues based on rules. Rules are stored in the rules.yaml file before compilation and are included in the binary file during build.
   - The prompt is checked using LLM. It is provided with the original prompt and a list of rules, and a write_results function is described. The LLM then calls the write_results function with the check results.

3. **Reporting**:
   - Formats all issues found during linting
   - Outputs a clean report to stdout
   - Shows "No issues found!" when the prompt passes all checks

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/promptlint.git

# Build the binary
cd promptlint
go build -o promptlint cmd/main.go
```

## Usage

```bash
# Check a prompt from a file
./promptlint -file=your-prompt.txt

# Check a prompt from stdin
cat your-prompt.txt | ./promptlint

# Check version
./promptlint -version
```

## Examples

```bash
# Example 1: Checking a prompt with issues
$ echo "This is a test prompt" | ./promptlint
Found issues:
Token 'is' is too short
Token 'a' is too short

# Example 2: Checking a prompt without issues
$ echo "This test prompt works properly" | ./promptlint
No issues found!
```

## Project Structure

The tool is built with a modular architecture:

- `/cmd/main.go`: Application entry point and I/O handling
- `/pkg/parser`: Tokenization module
- `/pkg/linter`: Rule checking module
- `/pkg/reporter`: Result formatting module

## Functionality

- Syntax and semantics analysis of prompts
- Identification of common errors and problems in prompt formulation
- Providing recommendations for improving the effectiveness and clarity of queries
- Integration into CI/CD pipelines and IDEs for automatic prompt quality checking

## Quick Start

1. Clone the repository
2. Build the binary:
   ```bash
   go build -o promptlint cmd/main.go
   ```
3. Run the utility:
   ```bash
   ./promptlint -file=your-prompt.txt
   ```

## Docker Usage

You can also run PromptLint using Docker:

```bash
# Build the Docker image
docker build -t promptlint .

# Run with a prompt file (mounting the current directory)
docker run --rm -v $(pwd):/data -e PROMPTLINT_API_KEY=your_api_key -e PROMPTLINT_API_ENDPOINT=your_api_endpoint promptlint -file=/data/your-prompt.txt

# Run with stdin
cat your-prompt.txt | docker run --rm -i -e PROMPTLINT_API_KEY=your_api_key -e PROMPTLINT_API_ENDPOINT=your_api_endpoint promptlint
```

### Using GitHub Container Registry

You can also pull the pre-built image from GitHub Container Registry:

```bash
# Pull the latest version
docker pull ghcr.io/username/promptlint:latest

# Run with your API keys
docker run --rm -v $(pwd):/data -e PROMPTLINT_API_KEY=your_api_key -e PROMPTLINT_API_ENDPOINT=your_api_endpoint ghcr.io/username/promptlint:latest -file=/data/your-prompt.txt
```

## Contribution and Collaboration

We welcome community contributions! If you have ideas for improvements or you've found bugs, please create an issue or submit a pull request.

## License

[MIT](LICENSE)

## Implementation Language

I recommend using the Go language. Go allows you to compile the program into a single statically linked binary file, which can be distributed without requiring users to install additional dependencies. In addition, the built-in cross-compilation capabilities make it an excellent choice for CI/CD, simplifying the automation of building, testing, and deployment.



