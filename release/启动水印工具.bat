@echo off
chcp 65001 >nul
echo ========================================
echo    水印工具 v1.0.0 - Windows版本
echo ========================================
echo.
echo 正在启动水印应用...
echo.
watermark-app.exe
if %errorlevel% neq 0 (
    echo.
    echo 程序运行出错，请检查系统环境
    pause
)
