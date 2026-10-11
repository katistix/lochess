# Installs or updates lochess on Windows (no admin rights needed):
#   irm https://raw.githubusercontent.com/katistix/lochess/main/install.ps1 | iex
#
# Installs to %LOCALAPPDATA%\Programs\lochess and adds a Start menu shortcut.
# The chess engine ships inside the app. Set LOCHESS_DIR to install elsewhere.

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue' # much faster Invoke-WebRequest

$repo = 'katistix/lochess'
$url = "https://github.com/$repo/releases/latest/download/lochess-windows-amd64.zip"
$dest = if ($env:LOCHESS_DIR) { $env:LOCHESS_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\lochess' }

if (-not [Environment]::Is64BitOperatingSystem) {
  throw 'lochess needs 64-bit Windows.'
}

$tmp = Join-Path ([IO.Path]::GetTempPath()) ("lochess-" + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
  Write-Host 'Downloading lochess for Windows...'
  $zip = Join-Path $tmp 'lochess.zip'
  Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing

  # Close a running copy so its files can be replaced.
  Get-Process lochess -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

  New-Item -ItemType Directory -Path $dest -Force | Out-Null
  Expand-Archive -Path $zip -DestinationPath $dest -Force

  $exe = Join-Path $dest 'lochess.exe'
  $startMenu = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs'
  $shell = New-Object -ComObject WScript.Shell
  $link = $shell.CreateShortcut((Join-Path $startMenu 'lochess.lnk'))
  $link.TargetPath = $exe
  $link.WorkingDirectory = $dest
  $link.Description = 'Play Stockfish at beginner-friendly levels and review your games'
  $link.Save()

  Write-Host "Installed lochess to $dest"
  Write-Host 'Open it from the Start menu, or run:' "`"$exe`""
}
finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
