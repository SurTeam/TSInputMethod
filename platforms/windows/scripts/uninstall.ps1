param(
  [string]$DllPath = "$(Join-Path $PSScriptRoot '..\build\tsime_tsf.dll')"
)

if (Test-Path $DllPath) {
  Start-Process -FilePath "regsvr32.exe" -ArgumentList @("/u", "/s", $DllPath) -Wait | Out-Null
}
Write-Host "Unregistered $DllPath (if present). Remove the language profile manually if it remains."
