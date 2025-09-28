package main

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ColorPicker creates a color picker dialog
type ColorPicker struct {
	window fyne.Window
	color  color.RGBA
}

// NewColorPicker creates a new color picker
func NewColorPicker(window fyne.Window) *ColorPicker {
	return &ColorPicker{
		window: window,
		color:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
}

// ShowColorPicker shows the color picker dialog
func (cp *ColorPicker) ShowColorPicker(currentColor color.RGBA, onColorSelected func(color.RGBA)) {
	cp.color = currentColor

	// Create color input fields
	rEntry := widget.NewEntry()
	rEntry.SetText(strconv.Itoa(int(currentColor.R)))
	rEntry.SetPlaceHolder("Red (0-255)")

	gEntry := widget.NewEntry()
	gEntry.SetText(strconv.Itoa(int(currentColor.G)))
	gEntry.SetPlaceHolder("Green (0-255)")

	bEntry := widget.NewEntry()
	bEntry.SetText(strconv.Itoa(int(currentColor.B)))
	bEntry.SetPlaceHolder("Blue (0-255)")

	aEntry := widget.NewEntry()
	aEntry.SetText(strconv.Itoa(int(currentColor.A)))
	aEntry.SetPlaceHolder("Alpha (0-255)")

	// Color preview - use a canvas rectangle for better color display
	colorPreview := canvas.NewRectangle(cp.color)
	colorPreview.SetMinSize(fyne.NewSize(100, 50))

	// Update preview function
	updatePreview := func() {
		r, _ := strconv.Atoi(rEntry.Text)
		g, _ := strconv.Atoi(gEntry.Text)
		b, _ := strconv.Atoi(bEntry.Text)
		a, _ := strconv.Atoi(aEntry.Text)

		// Clamp values
		if r < 0 {
			r = 0
		}
		if r > 255 {
			r = 255
		}
		if g < 0 {
			g = 0
		}
		if g > 255 {
			g = 255
		}
		if b < 0 {
			b = 0
		}
		if b > 255 {
			b = 255
		}
		if a < 0 {
			a = 0
		}
		if a > 255 {
			a = 255
		}

		cp.color = color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
		// Update color preview
		colorPreview.FillColor = cp.color
		colorPreview.Refresh()
	}

	// Bind text changes to preview update
	rEntry.OnChanged = func(string) { updatePreview() }
	gEntry.OnChanged = func(string) { updatePreview() }
	bEntry.OnChanged = func(string) { updatePreview() }
	aEntry.OnChanged = func(string) { updatePreview() }

	// Preset colors
	presetColors := []color.RGBA{
		{R: 255, G: 255, B: 255, A: 255}, // White
		{R: 0, G: 0, B: 0, A: 255},       // Black
		{R: 255, G: 0, B: 0, A: 255},     // Red
		{R: 0, G: 255, B: 0, A: 255},     // Green
		{R: 0, G: 0, B: 255, A: 255},     // Blue
		{R: 255, G: 255, B: 0, A: 255},   // Yellow
		{R: 255, G: 0, B: 255, A: 255},   // Magenta
		{R: 0, G: 255, B: 255, A: 255},   // Cyan
	}

	// Create preset color buttons
	presetButtons := container.NewGridWithColumns(4)
	for _, presetColor := range presetColors {
		// Create a colored rectangle for each preset color
		colorRect := canvas.NewRectangle(presetColor)
		colorRect.SetMinSize(fyne.NewSize(40, 40))

		// Create a transparent button overlay
		btn := widget.NewButton("", func(c color.RGBA) func() {
			return func() {
				cp.color = c
				rEntry.SetText(strconv.Itoa(int(c.R)))
				gEntry.SetText(strconv.Itoa(int(c.G)))
				bEntry.SetText(strconv.Itoa(int(c.B)))
				aEntry.SetText(strconv.Itoa(int(c.A)))
				colorPreview.FillColor = c
				colorPreview.Refresh()
				updatePreview()
			}
		}(presetColor))

		// Make button transparent and same size as rectangle
		btn.Importance = widget.LowImportance
		btn.Resize(fyne.NewSize(40, 40))

		// Create a container with the colored rectangle and transparent button
		colorContainer := container.NewStack(colorRect, btn)
		presetButtons.Add(colorContainer)
	}

	// Create dialog content
	content := container.NewVBox(
		widget.NewLabel("Select Color:"),
		container.NewHBox(
			container.NewVBox(
				widget.NewLabel("Red:"),
				rEntry,
				widget.NewLabel("Green:"),
				gEntry,
			),
			container.NewVBox(
				widget.NewLabel("Blue:"),
				bEntry,
				widget.NewLabel("Alpha:"),
				aEntry,
			),
			colorPreview,
		),
		widget.NewSeparator(),
		widget.NewLabel("Preset Colors:"),
		presetButtons,
	)

	dialog.ShowCustomConfirm("Select Color", "OK", "Cancel", content, func(ok bool) {
		if ok {
			onColorSelected(cp.color)
		}
	}, cp.window)
}

// EnhancedControls creates enhanced control widgets
type EnhancedControls struct {
	window         fyne.Window
	templateMgr    *TemplateManager
	colorPicker    *ColorPicker
	rotationSlider *widget.Slider
	shadowCheck    *widget.Check
	outlineCheck   *widget.Check
}

// NewEnhancedControls creates new enhanced controls
func NewEnhancedControls(window fyne.Window) *EnhancedControls {
	return &EnhancedControls{
		window:      window,
		templateMgr: NewTemplateManager(window),
		colorPicker: NewColorPicker(window),
	}
}

// CreateAdvancedControls creates advanced control widgets
func (ec *EnhancedControls) CreateAdvancedControls() *fyne.Container {
	// Rotation control
	ec.rotationSlider = widget.NewSlider(0, 360)
	ec.rotationSlider.Value = appData.Watermark.Rotation
	ec.rotationSlider.OnChanged = func(value float64) {
		appData.Watermark.Rotation = value
		updatePreview()
	}

	// Shadow effect
	ec.shadowCheck = widget.NewCheck("Shadow Effect", func(checked bool) {
		// TODO: Implement shadow effect
		updatePreview()
	})

	// Outline effect
	ec.outlineCheck = widget.NewCheck("Outline Effect", func(checked bool) {
		// TODO: Implement outline effect
		updatePreview()
	})

	// Color picker button
	colorBtn := widget.NewButton("Select Color", func() {
		ec.colorPicker.ShowColorPicker(appData.Watermark.Color, func(selectedColor color.RGBA) {
			appData.Watermark.Color = selectedColor
			updatePreview()
		})
	})

	// Font size with better control
	fontSizeEntry := widget.NewEntry()
	fontSizeEntry.SetText(strconv.Itoa(appData.Watermark.FontSize))
	fontSizeEntry.OnChanged = func(text string) {
		if size, err := strconv.Atoi(text); err == nil && size > 0 && size <= 500 {
			appData.Watermark.FontSize = size
			updatePreview()
		}
	}

	// Template management buttons
	saveTemplateBtn := widget.NewButton("Save Template", func() {
		ec.templateMgr.SaveTemplate()
	})

	loadTemplateBtn := widget.NewButton("Load Template", func() {
		ec.templateMgr.LoadTemplate()
		updatePreview()
	})

	deleteTemplateBtn := widget.NewButton("Delete Template", func() {
		ec.templateMgr.DeleteTemplate()
	})

	exportTemplateBtn := widget.NewButton("Export Template", func() {
		ec.templateMgr.ExportTemplate()
	})

	importTemplateBtn := widget.NewButton("Import Template", func() {
		ec.templateMgr.ImportTemplate()
	})

	// Advanced controls layout
	advancedControls := container.NewVBox(
		widget.NewLabel("Advanced Settings"),
		widget.NewSeparator(),

		widget.NewLabel("Rotation Angle:"),
		ec.rotationSlider,

		widget.NewLabel("Font Size:"),
		fontSizeEntry,

		widget.NewLabel("Color:"),
		colorBtn,

		widget.NewSeparator(),

		widget.NewLabel("Effects:"),
		ec.shadowCheck,
		ec.outlineCheck,

		widget.NewSeparator(),

		widget.NewLabel("Template Management:"),
		container.NewGridWithColumns(2,
			saveTemplateBtn,
			loadTemplateBtn,
			deleteTemplateBtn,
			exportTemplateBtn,
		),
		importTemplateBtn,
	)

	return advancedControls
}

// CreatePositionControls creates position control widgets
func (ec *EnhancedControls) CreatePositionControls() *fyne.Container {
	// 9-grid position buttons
	positionButtons := container.NewGridWithColumns(3)

	positions := []string{
		"top-left", "top-center", "top-right",
		"center-left", "center", "center-right",
		"bottom-left", "bottom-center", "bottom-right",
	}

	labels := []string{
		"Top-Left", "Top-Center", "Top-Right",
		"Center-Left", "Center", "Center-Right",
		"Bottom-Left", "Bottom-Center", "Bottom-Right",
	}

	for i, pos := range positions {
		btn := widget.NewButton(labels[i], func(selectedPos string) func() {
			return func() {
				appData.Watermark.Position = selectedPos
				updatePreview()
			}
		}(pos))
		positionButtons.Add(btn)
	}

	// Manual position controls
	xEntry := widget.NewEntry()
	xEntry.SetText(strconv.Itoa(appData.Watermark.X))
	xEntry.OnChanged = func(text string) {
		if x, err := strconv.Atoi(text); err == nil {
			appData.Watermark.X = x
			updatePreview()
		}
	}

	yEntry := widget.NewEntry()
	yEntry.SetText(strconv.Itoa(appData.Watermark.Y))
	yEntry.OnChanged = func(text string) {
		if y, err := strconv.Atoi(text); err == nil {
			appData.Watermark.Y = y
			updatePreview()
		}
	}

	// Position controls layout
	positionControls := container.NewVBox(
		widget.NewLabel("Position Settings"),
		widget.NewSeparator(),

		widget.NewLabel("Preset Positions:"),
		positionButtons,

		widget.NewSeparator(),

		widget.NewLabel("Manual Position:"),
		container.NewHBox(
			widget.NewLabel("X:"),
			xEntry,
			widget.NewLabel("Y:"),
			yEntry,
		),
	)

	return positionControls
}

// CreateOutputControls creates output control widgets
func (ec *EnhancedControls) CreateOutputControls() *fyne.Container {
	// Output format selection
	outputFormat := widget.NewSelect([]string{"JPEG", "PNG"}, func(value string) {
		appData.OutputFormat = value
	})
	outputFormat.SetSelected(appData.OutputFormat)

	// Quality control for JPEG
	qualitySlider := widget.NewSlider(1, 100)
	qualitySlider.Value = float64(appData.OutputQuality)
	qualitySlider.OnChanged = func(value float64) {
		appData.OutputQuality = int(value)
	}

	// File naming controls
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

	// Size scaling controls
	scaleEntry := widget.NewEntry()
	scaleEntry.SetText("100")
	scaleEntry.OnChanged = func(text string) {
		// TODO: Implement scaling
	}

	// Output controls layout
	outputControls := container.NewVBox(
		widget.NewLabel("Output Settings"),
		widget.NewSeparator(),

		widget.NewLabel("Output Format:"),
		outputFormat,

		widget.NewLabel("Quality (JPEG):"),
		qualitySlider,

		widget.NewLabel("File Naming:"),
		container.NewHBox(
			widget.NewLabel("Prefix:"),
			prefixEntry,
		),
		container.NewHBox(
			widget.NewLabel("Suffix:"),
			suffixEntry,
		),

		widget.NewLabel("Scale (%):"),
		scaleEntry,
	)

	return outputControls
}
