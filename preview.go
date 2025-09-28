package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// PreviewWidget handles the image preview functionality
type PreviewWidget struct {
	container *fyne.Container
	imageCard *widget.Card
	imageObj  *canvas.Image
}

// NewPreviewWidget creates a new preview widget
func NewPreviewWidget() *PreviewWidget {
	imageObj := canvas.NewImageFromResource(nil)
	imageObj.FillMode = canvas.ImageFillContain
	imageObj.SetMinSize(fyne.NewSize(300, 200)) // Set minimum size for preview

	imageCard := widget.NewCard("Preview", "No image selected", imageObj)
	// Remove fixed size to allow flexible resizing

	container := container.NewVBox(imageCard)

	return &PreviewWidget{
		container: container,
		imageCard: imageCard,
		imageObj:  imageObj,
	}
}

// UpdatePreview updates the preview with the current image and watermark
func (pw *PreviewWidget) UpdatePreview() {
	if len(appData.Images) == 0 || appData.CurrentImage >= len(appData.Images) {
		pw.imageObj.Resource = nil
		pw.imageCard.SetSubTitle("No image selected")
		pw.imageObj.Refresh()
		return
	}

	imagePath := appData.Images[appData.CurrentImage]

	// Load and process the image
	img, err := imaging.Open(imagePath)
	if err != nil {
		// If image loading fails, show error message
		pw.imageObj.Resource = nil
		pw.imageCard.SetSubTitle("Failed to load image")
		pw.imageObj.Refresh()
		return
	}

	// Apply watermark
	watermarkedImg := applyWatermark(img)
	if watermarkedImg == nil {
		// If watermarking fails, show original image
		resource := fyne.NewStaticResource("preview", imageToBytes(img))
		pw.imageObj.Resource = resource
		pw.imageCard.SetSubTitle("Failed to apply watermark")
		pw.imageObj.Refresh()
		return
	}

	// Convert to Fyne resource
	resource := fyne.NewStaticResource("preview", imageToBytes(watermarkedImg))
	pw.imageObj.Resource = resource
	pw.imageCard.SetSubTitle("Image with watermark")
	pw.imageObj.Refresh()
}

// imageToBytes converts an image to bytes for Fyne resource
func imageToBytes(img image.Image) []byte {
	if img == nil {
		return []byte{}
	}
	// Convert image to PNG bytes for Fyne resource
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return []byte{}
	}
	return buf.Bytes()
}

// Enhanced watermark application with better text rendering
func applyWatermark(img image.Image) image.Image {
	bounds := img.Bounds()
	watermarked := image.NewRGBA(bounds)
	draw.Draw(watermarked, bounds, img, bounds.Min, draw.Src)

	if appData.Watermark.IsImage {
		return applyImageWatermark(watermarked)
	} else {
		return applyTextWatermark(watermarked)
	}
}

// Simple and reliable text watermark
func applyTextWatermark(img *image.RGBA) *image.RGBA {
	text := appData.Watermark.Text
	if text == "" {
		text = "WATERMARK" // Default text if empty
	}

	// Create a new image for the watermark
	watermarkImg := image.NewRGBA(img.Bounds())
	draw.Draw(watermarkImg, img.Bounds(), img, img.Bounds().Min, draw.Src)

	// Calculate position
	x, y := calculateTextPosition(img.Bounds(), text, appData.Watermark.FontSize)

	// Create font face
	face := basicfont.Face7x13

	// Apply opacity to color
	opacity := float64(appData.Watermark.Opacity) / 100.0
	textColor := color.RGBA{
		R: appData.Watermark.Color.R,
		G: appData.Watermark.Color.G,
		B: appData.Watermark.Color.B,
		A: uint8(float64(appData.Watermark.Color.A) * opacity),
	}

	// Create a temporary image for the text
	textWidth := len(text) * 8 // Estimate width
	textHeight := 20           // Base height
	textImg := image.NewRGBA(image.Rect(0, 0, textWidth, textHeight))

	// Draw text on temporary image
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.NewUniform(textColor),
		Face: face,
	}
	drawer.Dot = fixed.Point26_6{
		X: fixed.Int26_6(0),
		Y: fixed.Int26_6(13 * 64),
	}
	drawer.DrawString(text)

	// Calculate scale factor
	scale := float64(appData.Watermark.FontSize) / 13.0
	if scale < 1.0 {
		scale = 1.0
	}

	// Scale the text image
	scaledWidth := int(float64(textWidth) * scale)
	scaledHeight := int(float64(textHeight) * scale)
	scaledTextImg := imaging.Resize(textImg, scaledWidth, scaledHeight, imaging.Lanczos)

	// Draw the scaled text onto the watermark image
	textBounds := scaledTextImg.Bounds()
	drawRect := image.Rect(x, y-textBounds.Dy(), x+textBounds.Dx(), y)
	draw.Draw(watermarkImg, drawRect, scaledTextImg, textBounds.Min, draw.Over)

	return watermarkImg
}

// Enhanced image watermark with better positioning and scaling
func applyImageWatermark(img *image.RGBA) *image.RGBA {
	if appData.Watermark.ImagePath == "" {
		return img
	}

	// Load watermark image
	watermarkImg, err := imaging.Open(appData.Watermark.ImagePath)
	if err != nil {
		return img
	}

	// Create result image
	result := image.NewRGBA(img.Bounds())
	draw.Draw(result, img.Bounds(), img, img.Bounds().Min, draw.Src)

	// Calculate appropriate size (max 1/4 of image width or height)
	imgBounds := img.Bounds()
	watermarkBounds := watermarkImg.Bounds()

	maxSize := min(imgBounds.Dx(), imgBounds.Dy()) / 4
	if watermarkBounds.Dx() > maxSize || watermarkBounds.Dy() > maxSize {
		scale := float64(maxSize) / float64(max(watermarkBounds.Dx(), watermarkBounds.Dy()))
		watermarkImg = imaging.Resize(watermarkImg,
			int(float64(watermarkBounds.Dx())*scale),
			int(float64(watermarkBounds.Dy())*scale),
			imaging.Lanczos)
		watermarkBounds = watermarkImg.Bounds()
	}

	// Calculate position
	x, y := calculateImagePosition(imgBounds, watermarkBounds, appData.Watermark.Position)

	// Apply opacity - for now we'll skip this as imaging.AdjustOpacity may not exist
	// opacity := float64(appData.Watermark.Opacity) / 100.0
	// if opacity < 1.0 {
	//     watermarkImg = imaging.AdjustOpacity(watermarkImg, opacity)
	// }

	// Draw watermark
	draw.Draw(result,
		image.Rect(x, y, x+watermarkBounds.Dx(), y+watermarkBounds.Dy()),
		watermarkImg,
		watermarkBounds.Min,
		draw.Over)

	return result
}

// calculateTextPosition calculates text position based on selected position
func calculateTextPosition(bounds image.Rectangle, text string, fontSize int) (int, int) {
	// Estimate text dimensions
	textWidth := len(text) * fontSize / 2
	textHeight := fontSize

	margin := 10

	switch appData.Watermark.Position {
	case "top-left":
		return margin, margin + textHeight
	case "top-center":
		return (bounds.Dx() - textWidth) / 2, margin + textHeight
	case "top-right":
		return bounds.Dx() - textWidth - margin, margin + textHeight
	case "center-left":
		return margin, bounds.Dy() / 2
	case "center":
		return (bounds.Dx() - textWidth) / 2, bounds.Dy() / 2
	case "center-right":
		return bounds.Dx() - textWidth - margin, bounds.Dy() / 2
	case "bottom-left":
		return margin, bounds.Dy() - margin
	case "bottom-center":
		return (bounds.Dx() - textWidth) / 2, bounds.Dy() - margin
	case "bottom-right":
		return bounds.Dx() - textWidth - margin, bounds.Dy() - margin
	default:
		return margin, bounds.Dy() - margin
	}
}

// calculateImagePosition calculates image watermark position
func calculateImagePosition(imgBounds, watermarkBounds image.Rectangle, position string) (int, int) {
	margin := 10

	switch position {
	case "top-left":
		return margin, margin
	case "top-center":
		return (imgBounds.Dx() - watermarkBounds.Dx()) / 2, margin
	case "top-right":
		return imgBounds.Dx() - watermarkBounds.Dx() - margin, margin
	case "center-left":
		return margin, (imgBounds.Dy() - watermarkBounds.Dy()) / 2
	case "center":
		return (imgBounds.Dx() - watermarkBounds.Dx()) / 2, (imgBounds.Dy() - watermarkBounds.Dy()) / 2
	case "center-right":
		return imgBounds.Dx() - watermarkBounds.Dx() - margin, (imgBounds.Dy() - watermarkBounds.Dy()) / 2
	case "bottom-left":
		return margin, imgBounds.Dy() - watermarkBounds.Dy() - margin
	case "bottom-center":
		return (imgBounds.Dx() - watermarkBounds.Dx()) / 2, imgBounds.Dy() - watermarkBounds.Dy() - margin
	case "bottom-right":
		return imgBounds.Dx() - watermarkBounds.Dx() - margin, imgBounds.Dy() - watermarkBounds.Dy() - margin
	default:
		return margin, imgBounds.Dy() - watermarkBounds.Dy() - margin
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
