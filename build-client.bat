@echo off
echo Building Hugo Visual Client...

REM Set build information
set APP_NAME=hugo-visual-client
set VERSION=1.0.0
set BUILD_DIR=dist

REM Create build directory
if not exist %BUILD_DIR% mkdir %BUILD_DIR%

echo.
echo Building for Windows (amd64)...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w -X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-windows-amd64.exe cmd/hugo-client/main.go

echo.
echo Building for Windows (386)...
set GOOS=windows
set GOARCH=386
go build -ldflags "-s -w -X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-windows-386.exe cmd/hugo-client/main.go

echo.
echo Building for Linux (amd64)...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w -X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-linux-amd64 cmd/hugo-client/main.go

echo.
echo Building for macOS (amd64)...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-s -w -X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-darwin-amd64 cmd/hugo-client/main.go

echo.
echo Building for macOS (arm64)...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-s -w -X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-darwin-arm64 cmd/hugo-client/main.go

echo.
echo Build completed! Executables are in the %BUILD_DIR% directory.
echo.
dir %BUILD_DIR%

pause