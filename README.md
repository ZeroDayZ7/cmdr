# Cmdr

A fast, developer-friendly CLI for bootstrapping microservices and handy Go tools.

## Table of Contents
<!-- * [Features](#features) -->
<!-- * [Usage](#usage) -->
<!-- * [Configuration](#configuration-file) -->
* [Project Profiles (Smart Detection)](./docs/PROFILES_GUIDE.md)
<!-- * [Available Commands](#other-available-commands) -->
<!-- * [Flags](#flags) -->

## Features

* Scaffold microservices in seconds
* Generate boilerplate for Go projects
* Handy developer utilities
* Custom templates for your services

## Usage

Cmdr helps you run tasks like migrations, dev servers, and more.

```bash
# Show help
cmdr --help
```

## Example Command: `ask`

The `ask` command allows you to send a question to an API (e.g., Gemini API) and get an immediate response.

```bash
cmdr ask -q "test"
```

## Output:

```bash
Response: Okay, I'm here! What would you like me to do?
```

## Configuration File

When you first run the ask command, a .config.json file will be generated in the same directory as the executable. This file will look like this:

```bash
{
  "gemini": {
    "url": "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent",
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
  ask             Ask questions to an API and get the answer
  clear           Deletes all directories with the given name in the current project
  completion      Generate the autocompletion script for the specified shell
  create-service  Create a new microservice from a template
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

* `-h`, `--help`: Display help for the `cmdr` command or any individual command.

## Example: Ask Command with Additional File Flag

You can also send a file along with your query to the `ask` command using the `-f` flag:

```bash
cmdr ask -q "What is the weather today?" -f example.txt
```

This will send the contents of `example.txt` along with your question to the API.
