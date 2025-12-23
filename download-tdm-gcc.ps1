# Download and install TDM-GCC for building Fyne applications
Write-Host "Downloading TDM-GCC..."

$url = "https://jmeubank.github.io/tdm-gcc/download/"
$installerPath = "$env:TEMP\tdm-gcc-installer.exe"

try {
    # Download TDM-GCC installer
    Write-Host "Downloading from GitHub releases..."
    $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/jmeubank/tdm-gcc/releases/latest"
    $asset = $releases.assets | Where-Object { $_.name -like "*tdm64*" -and $_.name -like "*.exe" }
    
    if ($asset) {
        Write-Host "Downloading $($asset.name)..."
        Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $installerPath
        
        Write-Host "Starting TDM-GCC installer..."
        Write-Host "Please follow the installation wizard and install to the default location (C:\TDM-GCC-64)"
        Start-Process -FilePath $installerPath -Wait
        
        # Add to PATH
        $tdmPath = "C:\TDM-GCC-64\bin"
        if (Test-Path $tdmPath) {
            $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
            if ($currentPath -notlike "*$tdmPath*") {
                [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$tdmPath", "User")
                Write-Host "Added TDM-GCC to PATH"
            }
        }
        
        Write-Host "TDM-GCC installation completed!"
        Write-Host "Please restart your command prompt to use the new PATH"
    } else {
        Write-Host "Could not find TDM-GCC installer. Please download manually from:"
        Write-Host "https://jmeubank.github.io/tdm-gcc/"
    }
} catch {
    Write-Host "Error downloading TDM-GCC: $($_.Exception.Message)"
    Write-Host "Please download manually from: https://jmeubank.github.io/tdm-gcc/"
}