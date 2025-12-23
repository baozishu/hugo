# Hugo Visual Client

A visual desktop client for Hugo static site generator built with Go and Fyne.

## Project Structure

```
hugo-visual-client/
├── cmd/
│   └── hugo-client/          # Application entry point
│       └── main.go
├── internal/                 # Private application code
│   ├── app/                  # Application controller
│   │   └── controller.go
│   ├── interfaces/           # Core interfaces
│   │   ├── project_manager.go
│   │   ├── content_manager.go
│   │   └── hugo_service.go
│   └── models/               # Data models
│       └── config.go
├── pkg/                      # Public packages
│   └── utils/                # Utility functions
│       └── file.go
├── go.mod                    # Go module definition
└── README.md
```

## Core Interfaces

### ProjectManager
Handles Hugo project creation, opening, validation, and management.

### ContentManager  
Manages content files, front matter parsing, and content operations.

### HugoService
Integrates with Hugo CLI for building, serving, and file watching.

## Getting Started

1. Install Go 1.21 or later
2. Install Hugo CLI
3. Run: `go mod tidy`
4. Build: `go build -o hugo-client cmd/hugo-client/main.go`
5. Run: `./hugo-client`

## Dependencies

- Fyne v2.4+ - GUI framework
- fsnotify - File system notifications
- goldmark - Markdown processing  
- yaml.v3 - YAML parsing
- toml - TOML parsing
- gopter - Property-based testing