@echo off
chcp 65001 >nul
echo ========================================
echo    Watermark Tool v1.0.0 - Windows
echo ========================================
echo.
echo Starting watermark application...
echo.
watermark-app.exe
if %errorlevel% neq 0 (
    echo.
    echo Application error, please check system environment
    pause
)
