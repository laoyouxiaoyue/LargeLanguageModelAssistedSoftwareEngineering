package main

import (
	"errors"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

// WatermarkConfig holds all watermark configuration
type WatermarkConfig struct {
	Text      string
	FontSize  int
	Color     color.RGBA
	Opacity   int
	Position  string
	X         int
	Y         int
	Rotation  float64
	ImagePath string
	IsImage   bool
}

// AppData holds the main application state
type AppData struct {
	Images        []string
	CurrentImage  int
	Watermark     WatermarkConfig
	OutputFolder  string
	OutputFormat  string
	OutputQuality int
	Prefix        string
	Suffix        string
}

var appData = &AppData{
	Watermark: WatermarkConfig{
		Text:     "WATERMARK",
		FontSize: 52, // Much larger default font size (4x scale)
		Color:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Opacity:  80,
		Position: "bottom-right",
		X:        10,
		Y:        10,
		Rotation: 0,
		IsImage:  false,
	},
	OutputFormat:  "JPEG",
	OutputQuality: 90,
	Prefix:        "wm_",
	Suffix:        "",
}

// setWindowsUTF8 sets Windows console to UTF-8 mode
func setWindowsUTF8() {
	if runtime.GOOS == "windows" {
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		setConsoleCP := kernel32.NewProc("SetConsoleCP")
		setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")

		// Set console code page to UTF-8 (65001)
		setConsoleCP.Call(uintptr(65001))
		setConsoleOutputCP.Call(uintptr(65001))
	}
}

func main() {
	// Set UTF-8 encoding for Windows
	setWindowsUTF8()

	myApp := app.NewWithID("com.watermark.app")

	myWindow := myApp.NewWindow("Watermark Tool")
	myWindow.Resize(fyne.NewSize(800, 500)) // Smaller initial size
	myWindow.CenterOnScreen()
	// Allow window to be resized by user

	// Create main UI
	content := createMainUI(myWindow)
	myWindow.SetContent(content)

	myWindow.ShowAndRun()
}

func createMainUI(window fyne.Window) *container.Split {
	// Left panel - Image list
	imageList := widget.NewList(
		func() int {
			return len(appData.Images)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(nil),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			container := obj.(*fyne.Container)
			label := container.Objects[1].(*widget.Label)
			if id < len(appData.Images) {
				label.SetText(filepath.Base(appData.Images[id]))
			}
		},
	)

	imageList.OnSelected = func(id widget.ListItemID) {
		appData.CurrentImage = id
		updatePreview()
	}

	// Import buttons
	importSingleBtn := widget.NewButton("Import Single Image", func() {
		importSingleImage(window)
	})
	importFolderBtn := widget.NewButton("Import Folder", func() {
		importFolder(window)
	})

	leftPanel := container.NewVBox(
		widget.NewLabel("Image List"),
		importSingleBtn,
		importFolderBtn,
		widget.NewSeparator(),
		imageList,
	)

	// Center panel - Preview
	previewWidget := NewPreviewWidget()
	globalPreviewWidget = previewWidget // Set global reference
	previewContainer := container.NewScroll(previewWidget.container)

	// Right panel - Enhanced Controls
	enhancedControls := NewEnhancedControls(window)

	// Create tabbed controls
	basicTab := createBasicControls(window)
	advancedTab := enhancedControls.CreateAdvancedControls()
	positionTab := enhancedControls.CreatePositionControls()
	outputTab := enhancedControls.CreateOutputControls()

	// Wrap each tab in a scroll container
	basicScroll := container.NewScroll(basicTab)
	advancedScroll := container.NewScroll(advancedTab)
	positionScroll := container.NewScroll(positionTab)
	outputScroll := container.NewScroll(outputTab)

	controlsTabs := container.NewAppTabs(
		container.NewTabItem("Basic Settings", basicScroll),
		container.NewTabItem("Advanced Settings", advancedScroll),
		container.NewTabItem("Position Settings", positionScroll),
		container.NewTabItem("Output Settings", outputScroll),
	)

	// Main split container
	mainSplit := container.NewHSplit(
		container.NewBorder(nil, nil, nil, nil, leftPanel),
		container.NewHSplit(
			previewContainer,
			controlsTabs,
		),
	)
	mainSplit.SetOffset(0.1)  // Left panel 20%
	mainSplit.SetOffset(0.65) // Right panel 35%, Preview gets 45% (balanced)

	return mainSplit
}

func createBasicControls(window fyne.Window) *fyne.Container {
	// Watermark type selection
	watermarkType := widget.NewRadioGroup([]string{"Text Watermark", "Image Watermark"}, func(value string) {
		appData.Watermark.IsImage = value == "Image Watermark"
		updatePreview()
	})
	watermarkType.SetSelected("Text Watermark")

	// Text watermark controls
	textEntry := widget.NewEntry()
	textEntry.SetText(appData.Watermark.Text)
	textEntry.OnChanged = func(text string) {
		appData.Watermark.Text = text
		updatePreview()
	}

	fontSizeSlider := widget.NewSlider(13, 200) // Much larger font size range
	fontSizeSlider.Value = float64(appData.Watermark.FontSize)
	fontSizeSlider.OnChanged = func(value float64) {
		appData.Watermark.FontSize = int(value)
		updatePreview()
	}

	opacitySlider := widget.NewSlider(0, 100)
	opacitySlider.Value = float64(appData.Watermark.Opacity)
	opacitySlider.OnChanged = func(value float64) {
		appData.Watermark.Opacity = int(value)
		updatePreview()
	}

	// Position selection
	positionGroup := widget.NewRadioGroup([]string{
		"top-left", "top-center", "top-right",
		"center-left", "center", "center-right",
		"bottom-left", "bottom-center", "bottom-right",
	}, func(value string) {
		appData.Watermark.Position = value
		updatePreview()
	})
	positionGroup.SetSelected(appData.Watermark.Position)

	// Image watermark controls
	imageSelectBtn := widget.NewButton("Select Image Watermark", func() {
		selectWatermarkImage(window)
	})

	// Output settings
	outputFormat := widget.NewSelect([]string{"JPEG", "PNG"}, func(value string) {
		appData.OutputFormat = value
	})
	outputFormat.SetSelected(appData.OutputFormat)

	qualitySlider := widget.NewSlider(1, 100)
	qualitySlider.Value = float64(appData.OutputQuality)
	qualitySlider.OnChanged = func(value float64) {
		appData.OutputQuality = int(value)
	}

	prefixEntry := widget.NewEntry()
	prefixEntry.SetText(appData.Prefix)
	prefixEntry.OnChanged = func(text string) {
		appData.Prefix = text
	}

	suffixEntry := widget.NewEntry()
	suffixEntry.SetText(appData.Suffix)
	suffixEntry.OnChanged = func(text string) {
		appData.Suffix = text
	}

	// Export button
	exportBtn := widget.NewButton("Export Images", func() {
		exportImages(window)
	})

	// Controls layout
	controls := container.NewVBox(
		widget.NewLabel("Watermark Settings"),
		widget.NewLabel("Type:"),
		watermarkType,
		widget.NewSeparator(),

		widget.NewLabel("Text Content:"),
		textEntry,
		widget.NewLabel("Font Size:"),
		fontSizeSlider,
		widget.NewLabel("Opacity:"),
		opacitySlider,
		widget.NewSeparator(),

		widget.NewLabel("Position:"),
		positionGroup,
		widget.NewSeparator(),

		widget.NewLabel("Image Watermark:"),
		imageSelectBtn,
		widget.NewSeparator(),

		widget.NewLabel("Output Settings"),
		widget.NewLabel("Format:"),
		outputFormat,
		widget.NewLabel("Quality:"),
		qualitySlider,
		widget.NewLabel("Prefix:"),
		prefixEntry,
		widget.NewLabel("Suffix:"),
		suffixEntry,
		widget.NewSeparator(),

		exportBtn,
	)

	return controls
}

func importSingleImage(window fyne.Window) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		path := reader.URI().Path()
		if isValidImageFormat(path) {
			appData.Images = append(appData.Images, path)
			updatePreview()
		} else {
			dialog.ShowError(errors.New("Unsupported format: Please select JPEG, PNG, BMP or TIFF images"), window)
		}
	}, window)
}

func importFolder(window fyne.Window) {
	dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if list == nil {
			return
		}

		// Get all image files from the folder
		files, err := list.List()
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		for _, file := range files {
			if isValidImageFormat(file.Path()) {
				appData.Images = append(appData.Images, file.Path())
			}
		}
		updatePreview()
	}, window)
}

func selectWatermarkImage(window fyne.Window) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		path := reader.URI().Path()
		if isValidImageFormat(path) {
			appData.Watermark.ImagePath = path
			appData.Watermark.IsImage = true
			updatePreview()
		} else {
			dialog.ShowError(errors.New("Unsupported format: Please select a valid image format"), window)
		}
	}, window)
}

func isValidImageFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".bmp" || ext == ".tiff" || ext == ".tif"
}

// Global preview widget reference
var globalPreviewWidget *PreviewWidget

func updatePreview() {
	// This function will be called when watermark settings change
	if globalPreviewWidget != nil {
		globalPreviewWidget.UpdatePreview()
	}
}

func exportImages(window fyne.Window) {
	if len(appData.Images) == 0 {
		dialog.ShowInformation("Info", "Please import images first", window)
		return
	}

	dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		if list == nil {
			return
		}

		appData.OutputFolder = list.Path()

		// Process each image
		for i, imagePath := range appData.Images {
			err := processImage(imagePath, i)
			if err != nil {
				dialog.ShowError(errors.New("Processing failed: "+err.Error()), window)
				return
			}
		}

		dialog.ShowInformation("Complete", "All images have been exported successfully", window)
	}, window)
}

func processImage(inputPath string, index int) error {
	// Load the original image
	img, err := imaging.Open(inputPath)
	if err != nil {
		return err
	}

	// Apply watermark
	watermarkedImg := applyWatermark(img)

	// Generate output filename
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	ext := ".jpg"
	if appData.OutputFormat == "PNG" {
		ext = ".png"
	}
	outputName := appData.Prefix + baseName + appData.Suffix + ext
	outputPath := filepath.Join(appData.OutputFolder, outputName)

	// Save the image
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	if appData.OutputFormat == "PNG" {
		err = png.Encode(outputFile, watermarkedImg)
	} else {
		err = jpeg.Encode(outputFile, watermarkedImg, &jpeg.Options{Quality: appData.OutputQuality})
	}

	return err
}
