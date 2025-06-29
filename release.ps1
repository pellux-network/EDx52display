$version = "0.1.2"

go clean

Remove-Item  .\EDx52Display -Force -Recurse -ErrorAction SilentlyContinue
Remove-Item .\Release-$version.zip -ErrorAction SilentlyContinue

mkdir EDx52Display

go build -ldflags -H=windowsgui

Copy-Item -Path EDx52display.exe,conf.yaml,LICENSE,README.md,names,DepInclude -Destination .\EDx52Display -Recurse

7z.exe a Release-$version.zip .\EDx52Display
