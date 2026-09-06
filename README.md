# mcpcmds

mcpcmds generates MCP (Model Context Protocol) servers from a JSON configuration file. Each tool in the config maps to an `os/exec` command, with arguments supplied via the MCP protocol.

## Usage

```bash
mcpcmds -config config.json
```

Or with a log file:

```bash
mcpcmds -config config.json -log-file server.log
```

## Configuration

The config file is a JSON object with the following structure:

```json
{
  "title": "My MCP Server",
  "desc": ["A collection of shell commands exposed as MCP tools"],
  "tools": [
    {
      "name": "echo",
      "title": "Echo",
      "desc": ["Echoes a message back", "A longer description"],
      "required": [
        {"name": "message", "desc": "The message to echo"}
      ],
      "optional": [
        {"name": "prefix", "desc": "Optional prefix for the message"}
      ],
      "command": ["echo", "${message}"]
    },
    {
      "name": "grep",
      "title": "Grep",
      "desc": ["Search for a pattern in a file"],
      "required": [
        {"name": "pattern", "desc": "The regex pattern to search for"},
        {"name": "file", "desc": "The file to search in"}
      ],
      "command": ["grep", "${pattern}", "${file}"]
    },
    {
      "name": "cat",
      "title": "Cat",
      "desc": ["Concatenate and display files"],
      "required": [
        {"name": "file", "desc": "The file to display"}
      ],
      "command": ["cat", "${file}"]
    }
  ]
}
```

### Fields

**Config**
- `title` — Server title (displayed to LLMs)
- `desc` — Server description

**Tool**
- `name` — The tool name used in MCP calls (required)
- `title` — Human-readable title
- `desc` — Slice of description strings
- `required` — Arguments that must be provided in the request
- `optional` — Arguments that can be omitted (default to empty string)
- `command` — The `os/exec` command to run. Supports `${KEY}` placeholders that get replaced with values from the request. Placeholders can appear multiple times and anywhere in the string.

### Placeholder syntax

Command strings can use `${KEY}` to reference tool arguments:

```json
{ "command": ["echo", "${message}"] }
```

If a command argument contains no `${}` placeholders, no replacement occurs for that element.

## How it works

1. The config file is loaded and parsed
2. Each tool becomes an MCP tool with its declared arguments
3. When the LLM calls a tool, `toolHandler` validates required arguments via `RequireString`
4. The `Command` slice is copied and `${KEY}` placeholders are replaced with actual argument values
5. An `os/exec.Command` is constructed and executed
6. Stdout is returned as the tool result; errors are returned to the client
