# 图片水印工具

一个使用 Go 语言开发的命令行图片水印工具，可以从图片的 EXIF 信息中提取拍摄时间作为水印。

## 功能特性

- 从图片 EXIF 信息中提取拍摄时间作为水印文本
- 支持自定义字体大小、颜色和位置
- 支持批量处理目录中的所有图片
- 自动创建输出目录
- 支持多种图片格式（JPG、PNG、BMP、GIF）
- 完善的日志系统，支持详细模式和进度显示
- 处理统计信息，包括成功/失败计数和耗时统计

## 安装依赖

```bash
go mod tidy
```

## 使用方法

### 基本用法

```bash
# 处理单个图片文件
go run main.go -input ./photo.jpg

# 处理整个目录
go run main.go -input ./photos/
```

### 高级用法

```bash
# 自定义字体大小、颜色和位置
go run main.go -input ./photo.jpg -size 32 -color white -position bottomright

# 处理目录并自定义参数
go run main.go -input ./photos/ -size 24 -color black -position topleft

# 启用详细日志模式
go run main.go -input ./photos/ -verbose
```

### 参数说明

- `-input`: 输入图片文件路径或目录路径（必需）
- `-size`: 字体大小，默认 24
- `-color`: 水印颜色，支持 white、black、red、blue、green，默认 white
- `-position`: 水印位置，支持 topleft、topright、bottomleft、bottomright、center，默认 bottomright
- `-verbose`: 显示详细日志信息，包括处理进度和详细状态

### 位置选项

- `topleft`: 左上角
- `topright`: 右上角
- `bottomleft`: 左下角
- `bottomright`: 右下角（默认）
- `center`: 居中

## 输出

程序会在原目录下创建一个名为 `原目录名_watermark` 的子目录，并将处理后的图片保存在其中。处理后的图片文件名会添加 `_watermark` 后缀。

例如：
- 输入：`./photos/image.jpg`
- 输出：`./photos/photos_watermark/image_watermark.jpg`

## 编译

```bash
# 编译为可执行文件
go build -o watermark main.go

# 使用编译后的程序
./watermark -input ./photo.jpg
```

## 日志功能

程序提供多级日志系统：

### 日志级别
- `[INFO]`: 基本信息，如开始处理、文件统计等
- `[WARNING]`: 警告信息，如未找到EXIF信息
- `[ERROR]`: 错误信息，如文件读取失败
- `[SUCCESS]`: 成功信息，如处理完成
- `[VERBOSE]`: 详细信息，仅在 `-verbose` 模式下显示
- `[PROGRESS]`: 处理进度，仅在 `-verbose` 模式下显示

### 统计信息
程序会在处理完成后显示详细的统计信息：
- 总文件数
- 已处理文件数
- 成功/失败文件数
- 总耗时
- 平均处理时间

## 注意事项

1. 如果图片没有 EXIF 信息或无法读取拍摄时间，程序会使用当前日期作为水印
2. 程序会自动创建输出目录
3. 支持递归处理子目录中的图片文件
4. 保持原图片的质量和格式
5. 使用 `-verbose` 参数可以查看详细的处理过程和进度
