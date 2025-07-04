$version = "0.1.2"

go clean

Remove-Item  .\EDx52Display -Force -Recurse -ErrorAction SilentlyContinue
Remove-Item .\Release-$version.zip -ErrorAction SilentlyContinue

mkdir EDx52Display

go build -ldflags "-H=windowsgui -s -w" # Strip debug info

# UPX compress (assumes upx.exe is in PATH or in a known folder, e.g., .\upx\upx.exe)
if (Test-Path .\upx\upx.exe) {
    .\upx\upx.exe --best --lzma .\EDx52display.exe
} elseif (Get-Command upx.exe -ErrorAction SilentlyContinue) {
    upx.exe --best --lzma .\EDx52display.exe
} else {
    Write-Host "UPX not found, skipping compression."
}

Copy-Item -Path EDx52display.exe,conf.yaml,LICENSE,README.md,names,DepInclude -Destination .\EDx52Display -Recurse

7z.exe a Release-$version.zip .\EDx52Display
