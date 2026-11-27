# CMDR

**A fast, developer-friendly CLI toolkit for Go developers.**

CMDR is a collection of practical command-line tools designed to simplify everyday development tasks, automate repetitive work, and improve productivity.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Available Commands](#available-commands)
- [Documentation](#documentation)
- [Supported Platforms](#supported-platforms)
- [Updates](#updates)

## Features

- **Project Management** – detect project structures and bootstrap development environments.
- **Asset Utilities** – generate favicon sets and resize images.
- **Cryptography** – encrypt and decrypt files, generate keys, and create secure random data.
- **File Management** – combine files, remove comments, and organize project contents.
- **System Utilities** – inspect processes, manage ports, and perform system operations.
- **Developer Tools** – explore project structures, inspect Go environments, and automate common tasks.

## Installation

Download the latest version for your platform from [GitHub Releases](https://github.com/ZeroDayZ7/cmdr/releases).

Extract the archive and make the `cmdr` executable available in your system PATH.

## Usage

### CMD-R Inspector

CMD-R Inspector is a standalone C++ program that inspects basic file information using binary signatures.

Display inspector help:

```bash
cmdr-inspector --help
```

Inspect a file:

```bash
cmdr-inspector file "C:/Files/my document.pdf"
```

Display the main help:

```bash
cmdr --help
```

Get help for a specific command:

```bash
cmdr assets --help
cmdr crypto --help
cmdr tree --help
```

## Available Commands

| Command | Description |
|---|---|
| `assets` | Image utilities for generating favicons and resizing images. |
| `ask` | Send questions to an AI API and receive responses. |
| `clear` | Delete directories with a specified name in the current project. |
| `completion` | Generate shell autocompletion scripts. |
| `create-service` | Create a new microservice from a template. |
| `crypto` | Cryptographic tools for encryption, decryption, keys, hashing, and random data. |
| `files-combine` | Combine files with selected extensions into a single file. |
| `hello` | Display a greeting. |
| `info` | Show recommended Go project structures and development information. |
| `killport` | Find and optionally terminate the process using a specified port. |
| `killprocess` | Find and terminate processes by name. |
| `remove-comments` | Remove comments from a file or directory. |
| `tree` | Display directory structures in a readable tree format. |
| `version` | Display application version and build information. |
| `doctor` | Check application configuration, permissions, connectivity, and updater readiness. |

For detailed usage and available flags, run:

```bash
cmdr <command> --help
```

## Examples

### Asset Utilities

Generate a favicon set:

```bash
cmdr assets favicon --input ./logo.png --output ./generated
```

Resize an image:

```bash
cmdr assets resize --input ./source.png --output ./out/resized.jpg --width 1200 --height 800 --fit cover
```

Initialize the default assets configuration:

```bash
cmdr assets config init
```

### Gemini ASK

Set `GEMINI_API_KEY` or configure `apiKey` in the CMDR configuration:

```powershell
$env:GEMINI_API_KEY="YOUR_API_KEY"
```

Ask a question, discover compatible models, or change the default:

```bash
cmdr ask -q "Explain this Go code"
cmdr ask "Explain this Go code"
cmdr ask models
cmdr ask models --refresh
cmdr ask --model gemini-3.8-flash "Review this code"
cmdr ask config
cmdr ask config --model gemini-3.8-flash
```

Configuration is stored at `%USERPROFILE%\.cmdr\.config.json` or `$HOME/.cmdr/.config.json`. `cmdr ask config` only reports key status, never the key itself.

### Cryptography

Generate a secure 20-character password:

```bash
cmdr crypto r -t p -l 20
```

Generate a 32-byte AES key:

```bash
cmdr crypto r -t a -l 32
```

Encrypt a file using Argon2id and AES-GCM:

```bash
cmdr crypto pass-encrypt -f secret.txt -p MyPassword123
```

Decrypt an encrypted file:

```bash
cmdr crypto pass-decrypt -f encrypted/secret.txt.enc -p MyPassword123
```

### Version and Diagnostics

Display version information:

```bash
cmdr version
```

Check application health:

```bash
cmdr doctor
```

Run detailed diagnostics:

```bash
cmdr doctor --verbose
```

## Documentation

Detailed guides and command documentation are available in the `docs/` directory.

| Document | Description |
|---|---|
| [Commands Reference](./docs/commands.md) | Detailed command reference and flags. |
| [Project Profiles](./docs/PROFILES_GUIDE.md) | Project detection and profile management. |
| [Prepare Password](./docs/PREPARE-PASSWORD.md) | Password preparation utilities. |
| [Tree Usage](./docs/TREE_USAGE.md) | Directory tree generation and options. |
| [Asset Utilities](./docs/assets.md) | Image processing and favicon generation. |
| [Updating](./docs/updating.md) | Stable and pre-release update channels. |
| [Releasing](./docs/releasing.md) | Building and publishing releases. |

## Supported Platforms

- Windows x64 (AMD64)
- Linux x64 (AMD64)

## Updates

CMDR provides two update channels:

- **Stable** – recommended for everyday use.
- **Pre-release** – early access to upcoming versions.

Check for stable updates:

```bash
cmdr update --check
```

Install the latest stable version:

```bash
cmdr update
```

Check for pre-releases:

```bash
cmdr update --check --prerelease
```

Install the latest pre-release:

```bash
cmdr update --prerelease
```

## License

See [LICENSE](./LICENSE) for license information.