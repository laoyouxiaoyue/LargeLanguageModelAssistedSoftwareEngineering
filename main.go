package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// WatermarkConfig 水印配置
type WatermarkConfig struct {
	FontSize int
	Color    string
	Position string
}

// Position 位置枚举
type Position int

const (
	TopLeft Position = iota
	TopRight
	BottomLeft
	BottomRight
	Center
)

var positionMap = map[string]Position{
	"topleft":     TopLeft,
	"topright":    TopRight,
	"bottomleft":  BottomLeft,
	"bottomright": BottomRight,
	"center":      Center,
}

func main() {
	// 命令行参数
	var (
		inputPath = flag.String("input", "", "输入图片文件路径")
		fontSize  = flag.Int("size", 24, "字体大小")
		color     = flag.String("color", "white", "水印颜色 (white, black, red, blue, green)")
		position  = flag.String("position", "bottomright", "水印位置 (topleft, topright, bottomleft, bottomright, center)")
	)
	flag.Parse()

	if *inputPath == "" {
		fmt.Println("使用方法: watermark -input <图片路径> [-size <字体大小>] [-color <颜色>] [-position <位置>]")
		fmt.Println("示例: watermark -input ./photos/image.jpg -size 32 -color white -position bottomright")
		os.Exit(1)
	}

	config := &WatermarkConfig{
		FontSize: *fontSize,
		Color:    *color,
		Position: *position,
	}

	// 检查输入路径
	if _, err := os.Stat(*inputPath); os.IsNotExist(err) {
		log.Fatalf("输入文件不存在: %s", *inputPath)
	}

	// 处理单个文件或目录
	if isDir(*inputPath) {
		processDirectory(*inputPath, config)
	} else {
		processFile(*inputPath, config)
	}
}

// isDir 检查路径是否为目录
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// processDirectory 处理目录中的所有图片
func processDirectory(dirPath string, config *WatermarkConfig) {
	// 创建输出目录
	outputDir := dirPath + "_watermark"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 遍历目录中的文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查是否为支持的图片格式
		if isImageFile(path) {
			processFile(path, config)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("遍历目录失败: %v", err)
	}

	fmt.Printf("处理完成！输出目录: %s\n", outputDir)
}

// processFile 处理单个图片文件
func processFile(inputPath string, config *WatermarkConfig) {
	fmt.Printf("处理文件: %s\n", inputPath)

	// 读取图片
	img, err := imaging.Open(inputPath)
	if err != nil {
		log.Printf("读取图片失败 %s: %v", inputPath, err)
		return
	}

	// 获取EXIF信息
	watermarkText := getExifDate(inputPath)
	if watermarkText == "" {
		watermarkText = time.Now().Format("2006-01-02")
		fmt.Printf("未找到EXIF日期信息，使用当前日期: %s\n", watermarkText)
	}

	// 添加水印
	watermarkedImg := addWatermark(img, watermarkText, config)

	// 生成输出路径
	outputPath := generateOutputPath(inputPath)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		log.Printf("创建输出目录失败: %v", err)
		return
	}

	// 保存图片
	if err := saveImage(watermarkedImg, outputPath); err != nil {
		log.Printf("保存图片失败 %s: %v", outputPath, err)
		return
	}

	fmt.Printf("已保存: %s\n", outputPath)
}

// getExifDate 从EXIF信息中获取拍摄日期
func getExifDate(imagePath string) string {
	file, err := os.Open(imagePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// 注册相机厂商的EXIF解析器
	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(file)
	if err != nil {
		return ""
	}

	// 尝试获取拍摄时间
	tm, err := x.DateTime()
	if err != nil {
		// 如果DateTime失败，尝试获取原始日期时间
		dateTime, err := x.Get(exif.DateTime)
		if err != nil {
			return ""
		}
		dateTimeStr := dateTime.String()
		if len(dateTimeStr) >= 10 {
			return dateTimeStr[:10] // 返回 YYYY:MM:DD 格式
		}
		return ""
	}

	return tm.Format("2006-01-02")
}

// addWatermark 在图片上添加水印
func addWatermark(img image.Image, text string, config *WatermarkConfig) image.Image {
	// 创建RGBA图像
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// 获取颜色
	textColor := getColor(config.Color)

	// 获取位置
	pos := getPosition(config.Position)

	// 绘制文本
	drawText(rgba, text, textColor, pos, config.FontSize, bounds)

	return rgba
}

// getColor 根据字符串获取颜色
func getColor(colorStr string) color.RGBA {
	switch strings.ToLower(colorStr) {
	case "black":
		return color.RGBA{0, 0, 0, 255}
	case "red":
		return color.RGBA{255, 0, 0, 255}
	case "blue":
		return color.RGBA{0, 0, 255, 255}
	case "green":
		return color.RGBA{0, 255, 0, 255}
	case "white":
		fallthrough
	default:
		return color.RGBA{255, 255, 255, 255}
	}
}

// getPosition 根据字符串获取位置
func getPosition(posStr string) Position {
	if pos, exists := positionMap[strings.ToLower(posStr)]; exists {
		return pos
	}
	return BottomRight // 默认右下角
}

// drawText 在图像上绘制文本
func drawText(img *image.RGBA, text string, textColor color.RGBA, pos Position, fontSize int, bounds image.Rectangle) {
	// 使用基本字体
	face := basicfont.Face7x13

	// 计算文本尺寸
	textWidth := len(text) * 7 // 基本字体每个字符约7像素宽
	textHeight := 13           // 基本字体高度

	// 计算位置
	var x, y int
	switch pos {
	case TopLeft:
		x = 10
		y = 10 + textHeight
	case TopRight:
		x = bounds.Dx() - textWidth - 10
		y = 10 + textHeight
	case BottomLeft:
		x = 10
		y = bounds.Dy() - 10
	case BottomRight:
		x = bounds.Dx() - textWidth - 10
		y = bounds.Dy() - 10
	case Center:
		x = (bounds.Dx() - textWidth) / 2
		y = (bounds.Dy() + textHeight) / 2
	}

	// 绘制文本
	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  point,
	}
	d.DrawString(text)
}

// isImageFile 检查文件是否为支持的图片格式
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".bmp" || ext == ".gif"
}

// generateOutputPath 生成输出文件路径
func generateOutputPath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	filename := filepath.Base(inputPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	// 创建输出目录
	outputDir := filepath.Join(dir, dir+"_watermark")

	// 生成输出文件名
	outputFilename := name + "_watermark" + ext
	return filepath.Join(outputDir, outputFilename)
}

// saveImage 保存图片到文件
func saveImage(img image.Image, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
	case ".png":
		return png.Encode(file, img)
	default:
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
	}
}
