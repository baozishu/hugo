@echo off
echo Building Hugo Visual Client for Windows...

REM Set build information
set APP_NAME=hugo-visual-client
set VERSION=1.0.0
set BUILD_DIR=dist

REM Create build directory
if not exist %BUILD_DIR% mkdir %BUILD_DIR%

echo.
echo Building for Windows (amd64)...
go build -ldflags "-s -w -X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-windows-amd64.exe cmd/hugo-client/main.go

echo.
if %ERRORLEVEL% == 0 (
    echo Build completed successfully!
    echo Executable is in the %BUILD_DIR% directory.
    dir %BUILD_DIR%
) else (
    echo Build failed with error code %ERRORLEVEL%
)

pause