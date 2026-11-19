param(
  [string]$DllPath = "$(Join-Path $PSScriptRoot '..\build\tsime_tsf.dll')"
)

if (-not (Test-Path $DllPath)) {
  Write-Error "missing $DllPath — build with CMake/MSVC first"
  exit 1
}

$reg = Start-Process -FilePath "regsvr32.exe" -ArgumentList @("/s", $DllPath) -Wait -PassThru
if ($reg.ExitCode -ne 0) {
  Write-Error "regsvr32 failed: $($reg.ExitCode)"
  exit $reg.ExitCode
}

Write-Host "Registered $DllPath"
Write-Host "Add the TSInputMethod language profile in Windows Settings → Time & language → Language & region."
Write-Host "Run imed.exe with Named Pipe TSIME_PIPE or default \\.\pipe\tsime-<user>."
