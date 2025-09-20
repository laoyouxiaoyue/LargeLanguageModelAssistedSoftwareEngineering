@echo off
echo 测试图片水印工具
echo.

echo 1. 测试帮助信息
.\watermark.exe
echo.

echo 2. 测试处理不存在的文件
.\watermark.exe -input nonexistent.jpg
echo.

echo 3. 测试处理空目录
.\watermark.exe -input test_images
echo.

echo 测试完成！
pause
