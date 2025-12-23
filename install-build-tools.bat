@echo off
echo Installing build tools for Hugo Visual Client...

echo.
echo Checking if Chocolatey is installed...
choco --version >nul 2>&1
if %errorlevel% neq 0 (
    echo Chocolatey not found. Installing Chocolatey...
    powershell -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
    
    echo Please restart this script after Chocolatey installation completes.
    pause
    exit /b 1
)

echo.
echo Installing MinGW-w64 (C compiler for CGO)...
choco install mingw -y

echo.
echo Refreshing environment variables...
refreshenv

echo.
echo Verifying installation...
gcc --version
if %errorlevel% neq 0 (
    echo Warning: GCC not found in PATH. You may need to restart your command prompt.
    echo Manual installation: Download TDM-GCC from https://jmeubank.github.io/tdm-gcc/
)

echo.
echo Enabling CGO...
go env -w CGO_ENABLED=1

echo.
echo Build tools installation completed!
echo You can now run: go build -o hugo-visual-client.exe cmd/hugo-client/main.go
echo.

pause