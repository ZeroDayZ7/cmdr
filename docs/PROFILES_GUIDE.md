# CMDR Profiles Configuration Guide

The `.cmdr_profiles.json` file is located in the same directory as your `cmdr` executable. It allows the CLI to automatically detect the project type and apply specific rules for combining files and removing comments.

## File Structure

### 1. Global Settings
`global.ignore`: A list of files and folders that are **always** ignored, regardless of the project type (e.g., `.git`, `node_modules`).

### 2. Profiles
Each profile defines how to handle a specific environment or language.

| Field | Description |
| :--- | :--- |
| `name` | Unique identifier for the profile. |
| `priority` | Detection weight. Higher values are checked first. |
| `detect` | Criteria used to identify the project (files, folders, or extensions). |
| `extensions` | Which file types should be collected during a `files-combine` operation. |
| `ignore` | Language-specific exclusions (e.g., `.dart_tool` for Flutter). |
| `generated` | Suffixes for auto-generated files that should be skipped. |

---

## How to Add a New Profile

To add support for a new language (e.g., Python), append a new object to the `profiles` array:

```json
{
  "name": "python",
  "priority": 80,
  "detect": {
    "files": ["requirements.txt", "manage.py", "main.py"],
    "folders": ["venv", "env"],
    "extensions": [".py"]
  },
  "extensions": [".py"],
  "ignore": ["__pycache__", "venv"],
  "generated": []
}