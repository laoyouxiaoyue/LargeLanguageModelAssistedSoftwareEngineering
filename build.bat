@echo off
echo Building Watermark Application...

REM Install dependencies
echo Installing dependencies...
go mod tidy

REM Build for Windows
echo Building for Windows...
go build -o watermark-app.exe -ldflags "-s -w"

REM Build for macOS (if on macOS)
REM go build -o watermark-app -ldflags "-s -w"

echo Build complete!
echo Run with: watermark-app.exe
pause
