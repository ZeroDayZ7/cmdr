# Cmdr

A fast, developer-friendly CLI for bootstrapping microservices and handy Go tools.

## Table of Contents

- [Project Profiles (Smart Detection)](./docs/PROFILES_GUIDE.md)
- [Prepare Password](./docs/PREPARE-PASSWORD.md)
- [Tree Usage](./docs/TREE_USAGE.md)
- [Assets](./docs/assets.md)
- [Releasing](./docs/releasing.md)
- [Updating](./docs/updating.md)

## Usage

Cmdr helps you run tasks like migrations, dev servers, and more.

```bash
# Show help
cmdr --help

```

## Asset Utilities

The `assets` command groups image-processing helpers for generating favicon sets and resizing image files for web and application projects.

```bash
# Generate a default favicon set from a source image
cmdr assets favicon --input ./logo.png --output ./generated

# Resize an image to a fixed size and save it to a target path
cmdr assets resize --input ./source.png --output ./out/resized.jpg --width 1200 --height 800 --fit cover

# Initialize the default assets configuration file
cmdr assets config init
```

For the detailed command reference, see [docs/assets.md](./docs/assets.md).

## Example Command: `ask`

The `ask` command allows you to send a question to an API (e.g., Gemini API) and get an immediate response.

```bash
cmdr ask -q "test"

```

## Output:

```bash
Response: Okay, I'm here! What would you like me to do?

```

## Example Command: `crypto`

The `crypto` command provides cryptographic utilities including file encryption (AES-GCM / Argon2id), key generation, and random data creation.

```bash
# Generate a secure 20-character password
cmdr crypto r -t p -l 20

# Generate a 32-byte AES key (HEX/Base64)
cmdr crypto r -t a -l 32

# Encrypt a file using a password (Argon2id + AES-GCM)
cmdr crypto pass-encrypt -f secret.txt -p MyPassword123

# Decrypt a file encrypted with Argon2id
cmdr crypto pass-decrypt -f encrypted/secret.txt.enc -p MyPassword123

```

## Configuration File

When you first run the ask command, a .config.json file will be generated in the same directory as the executable. This file will look like this:

```json
{
  "gemini": {
    "url": "[https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent](https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent)",
    "apiKey": "your_gemini_api_key_here"
  }
}
```

## Other Available Commands

Cmdr provides several commands to help with your Go and microservices tasks. Here’s a list of available commands:

```bash
cmdr helps you run tasks like migrate, dev server, etc.

Usage:
  cmdr [flags]
  cmdr [command]

Available Commands:
  assets          Image utilities: generate favicon sets and resize images
  ask             Ask questions to an API and get the answer
  clear           Deletes all directories with the given name in the current project
  completion      Generate the autocompletion script for the specified shell
  create-service  Create a new microservice from a template
  crypto          Cryptographic operations: keys, encryption, random data, hashing
  files-combine   Combine contents of files with chosen extensions into one file
  hello           Say hello
  help            Help about any command
  info            Show recommended Go project structure and common use cases
  killport        Find and optionally kill the process using a port
  killprocess     Find and optionally kill processes by name
  remove-comments Remove comments from a specific file or all files in a directory
  tree            Display a tree view of folders and files
  version         Shows the CLI version

Flags:
  -h, --help   help for cmdr

Use "cmdr [command] --help" for more information about a command.

```

## Flags

- `-h`, `--help`: Display help for the `cmdr` command or any individual command.

## Builds

Windows amd64 and arm64 artifacts use the shared project version and embedded icon metadata:

```powershell
powershell -ExecutionPolicy Bypass -File ./scripts/build-windows.ps1 -Arch amd64
powershell -ExecutionPolicy Bypass -File ./scripts/build-windows.ps1 -Arch arm64
powershell -ExecutionPolicy Bypass -File ./scripts/build-windows.ps1 -Arch all
```

Linux builds are kept separate and should be run on Linux or inside a Linux container/WSL:

```bash
bash ./scripts/build-linux.sh amd64
bash ./scripts/build-linux.sh arm64
bash ./scripts/build-linux.sh all
```

Build artifacts are written below `bin/windows/<arch>` and `bin/linux/<arch>`.

## Terminal branding

Running `cmdr` without arguments, `cmdr help`, or `cmdr --help` displays a compact branded banner before Cobra help:

```text
  C  cmdr
     DEVELOPER TOOLKIT · v2.1.4

  Automate your development workflow.

  Run cmdr --help to explore commands.
```

The banner uses the version from `config/metadata.go`. ANSI colors are disabled with `NO_COLOR`, when stdout is redirected or piped, and when the terminal environment is not detected as color-capable. Regular commands and subcommand help do not show the root banner.
