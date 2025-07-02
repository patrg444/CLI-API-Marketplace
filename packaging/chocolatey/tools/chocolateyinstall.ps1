$ErrorActionPreference = 'Stop'

$packageName = 'apidirect'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$version = '1.0.0'

$packageArgs = @{
  packageName    = $packageName
  unzipLocation  = $toolsDir
  url64bit       = "https://github.com/api-direct/cli/releases/download/v$version/apidirect_${version}_windows_x86_64.zip"
  url            = "https://github.com/api-direct/cli/releases/download/v$version/apidirect_${version}_windows_i386.zip"
  checksum64     = 'PLACEHOLDER_SHA256_WINDOWS_AMD64'
  checksumType64 = 'sha256'
  checksum       = 'PLACEHOLDER_SHA256_WINDOWS_386'
  checksumType   = 'sha256'
}

Install-ChocolateyZipPackage @packageArgs

# Create a shim for the executable
$exePath = Join-Path $toolsDir 'apidirect.exe'
Install-BinFile -Name 'apidirect' -Path $exePath

Write-Host "API Direct CLI has been installed successfully!" -ForegroundColor Green
Write-Host "Run 'apidirect --help' to get started." -ForegroundColor Cyan