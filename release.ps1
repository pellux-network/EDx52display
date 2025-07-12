$version = Read-Host "Enter the release version"

go clean

Remove-Item  .\EDx52display* -Force -Recurse -ErrorAction SilentlyContinue

mkdir EDx52display-$version

go build -ldflags "-H=windowsgui -s -w" -o EDx52display-$version.exe # Strip debug info

Copy-Item -Path EDx52display-$version.exe,conf.yaml,LICENSE,README.md,names,bin -Destination .\EDx52Display-$version -Recurse

7z.exe a EDx52display-$version-portable-amd64.zip .\EDx52display-$version

pause