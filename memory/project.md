# Project

## Overview
PromptLint - CLI utility for validating LLM prompts with exclusive use of LLM API to check compliance with best practices.

## Goals
- Automating the process of checking prompt quality before use
- Standardizing prompt writing in teams
- Simplifying integration of prompt validation into CI/CD
- Deep validation of prompts through LLM API
- Providing recommendations for improving prompts

## Requirements
- Reading prompts from files and stdin
- Support for customizable checking rules (YAML)
- Advanced validation through LLM API
- Clear error output with explanations and recommendations
- Exclusive use of LLM API for validation
- Strict requirements for API availability
- Compact binary file without external dependencies
- Monolithic design with all functionality in a single file
- Containerized deployment option with Docker
- CI/CD integration with GitHub Actions

## Target Audience
- LLM application developers
- Prompt engineers
- DevOps engineers (CI/CD integration)

## Use Cases
| Case | Description |
|--------|----------|
| CI check | Automatic validation of prompts in CI/CD pipeline |
| Local development | Quick checking of prompts during development |
| Team standardization | Unifying approach to prompt writing |
| Training | Helping newcomers learn principles of effective prompts |
| Deep validation | Using external LLM for detailed analysis |

## Success Metrics
- ⏱️ Prompt processing speed (average <2s)
- 🎯 Issue detection accuracy (>95%)
- 🛠️ Usefulness of recommendations for fixing
- 🔄 Ease of integration into existing processes
- 📝 Compliance of all prompts with a unified quality standard

## Constraints
- Minimal number of dependencies
- Using only Go standard library + yaml.v3
- Following idiomatic Go code principles
- Compact binary file size
- Mandatory API availability for application operation
- Simplified codebase with all functionality in one file