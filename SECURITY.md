# Security Policy

## Supported Versions

Only the latest version of ZAR Blockchain is currently supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| v0.1.x  | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability within this project, please do not report it publicly. Instead, please follow these steps:

1.  **Do not** open a GitHub issue for security vulnerabilities.
2.  Send an email to `security@zar-chain.org` (Conceptual).
3.  Provide a detailed description of the vulnerability and steps to reproduce it.

We will acknowledge your report within 48 hours and provide a timeline for a fix.

## GitHub Security Features

To ensure the safety of this project, we recommend enabling the following in the repository settings:

- **Secret Scanning**: Automatically detects secrets committed to the repository.
- **Push Protection**: Prevents commits containing secrets from being pushed.
- **CodeQL Analysis**: Automated code scanning for vulnerabilities (Configured in `.github/workflows/security.yml`).
