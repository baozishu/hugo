# Build Instructions

## Prerequisites

### Windows
- Go 1.21+
- CGO enabled
- GCC compiler (via MinGW-w64 or TDM-GCC)
- Git

### Linux/macOS
- Go 1.21+
- Standard build tools

## Building

### Windows (with proper build environment)
```bash
set CGO_ENABLED=1
go build -o hugo-client.exe cmd/hugo-client/main.go
```

### Linux/macOS
```bash
go build -o hugo-client cmd/hugo-client/main.go
```

## Testing

Run all tests:
```bash
go test ./...
```

Run property-based tests:
```bash
go test -v ./internal/interfaces
```

## Development

The project structure is set up and ready for development:

- ✅ Go module initialized
- ✅ Core interfaces defined
- ✅ Basic directory structure created
- ✅ Application entry point created
- ✅ Property-based test implemented and passing
- ✅ All packages compile successfully

Note: GUI compilation requires proper CGO build environment on Windows.