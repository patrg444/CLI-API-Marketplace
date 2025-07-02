$ErrorActionPreference = 'Stop'

$packageName = 'apidirect'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

# Remove the shim
Uninstall-BinFile -Name 'apidirect'

Write-Host "API Direct CLI has been uninstalled." -ForegroundColor Green