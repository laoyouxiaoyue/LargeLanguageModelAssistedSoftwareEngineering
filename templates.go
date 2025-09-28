package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// TemplateManager handles saving and loading watermark templates
type TemplateManager struct {
	templates map[string]*WatermarkConfig
	window    fyne.Window
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(window fyne.Window) *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*WatermarkConfig),
		window:    window,
	}
	tm.loadTemplates()
	return tm
}

// SaveTemplate saves the current watermark configuration as a template
func (tm *TemplateManager) SaveTemplate() {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter template name")

	dialog.ShowCustomConfirm("Save Template", "OK", "Cancel", entry, func(save bool) {
		if !save {
			return
		}

		name := entry.Text
		if name == "" {
			dialog.ShowError(errors.New("Error: Template name cannot be empty"), tm.window)
			return
		}

		// Create a copy of current watermark config
		template := &WatermarkConfig{
			Text:      appData.Watermark.Text,
			FontSize:  appData.Watermark.FontSize,
			Color:     appData.Watermark.Color,
			Opacity:   appData.Watermark.Opacity,
			Position:  appData.Watermark.Position,
			X:         appData.Watermark.X,
			Y:         appData.Watermark.Y,
			Rotation:  appData.Watermark.Rotation,
			ImagePath: appData.Watermark.ImagePath,
			IsImage:   appData.Watermark.IsImage,
		}

		tm.templates[name] = template
		tm.saveTemplatesToFile()

		dialog.ShowInformation("Success", "Template saved", tm.window)
	}, tm.window)
}

// LoadTemplate loads a saved template
func (tm *TemplateManager) LoadTemplate() {
	if len(tm.templates) == 0 {
		dialog.ShowInformation("Info", "No saved templates", tm.window)
		return
	}

	// Create list of template names
	var names []string
	for name := range tm.templates {
		names = append(names, name)
	}

	selectWidget := widget.NewSelect(names, nil)
	selectWidget.PlaceHolder = "Select Template"

	dialog.ShowCustomConfirm("Load Template", "Load", "Cancel", selectWidget, func(load bool) {
		if !load {
			return
		}

		selectedName := selectWidget.Selected
		if selectedName == "" {
			return
		}

		template, exists := tm.templates[selectedName]
		if !exists {
			dialog.ShowError(errors.New("Error: Template not found"), tm.window)
			return
		}

		// Apply template to current watermark config
		appData.Watermark = *template

		dialog.ShowInformation("Success", "模板已Load", tm.window)
	}, tm.window)
}

// DeleteTemplate deletes a saved template
func (tm *TemplateManager) DeleteTemplate() {
	if len(tm.templates) == 0 {
		dialog.ShowInformation("Info", "No saved templates", tm.window)
		return
	}

	// Create list of template names
	var names []string
	for name := range tm.templates {
		names = append(names, name)
	}

	selectWidget := widget.NewSelect(names, nil)
	selectWidget.PlaceHolder = "Select template to delete"

	dialog.ShowCustomConfirm("Delete Template", "Delete", "Cancel", selectWidget, func(shouldDelete bool) {
		if !shouldDelete {
			return
		}

		selectedName := selectWidget.Selected
		if selectedName == "" {
			return
		}

		// Delete template
		delete(tm.templates, selectedName)
		tm.saveTemplatesToFile()

		dialog.ShowInformation("Success", "Template deleted", tm.window)
	}, tm.window)
}

// ExportTemplate exports templates to a file
func (tm *TemplateManager) ExportTemplate() {
	if len(tm.templates) == 0 {
		dialog.ShowInformation("Info", "No saved templates", tm.window)
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, tm.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Convert templates to JSON
		data, err := json.MarshalIndent(tm.templates, "", "  ")
		if err != nil {
			dialog.ShowError(err, tm.window)
			return
		}

		// Write to file
		_, err = writer.Write(data)
		if err != nil {
			dialog.ShowError(err, tm.window)
			return
		}

		dialog.ShowInformation("Success", "Template exported", tm.window)
	}, tm.window)
}

// ImportTemplate imports templates from a file
func (tm *TemplateManager) ImportTemplate() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, tm.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// Read file content
		var data []byte
		_, err = reader.Read(data)
		if err != nil {
			dialog.ShowError(err, tm.window)
			return
		}

		// Parse JSON
		var importedTemplates map[string]*WatermarkConfig
		err = json.Unmarshal(data, &importedTemplates)
		if err != nil {
			dialog.ShowError(errors.New("Error: Invalid template file"), tm.window)
			return
		}

		// Merge with existing templates
		for name, template := range importedTemplates {
			tm.templates[name] = template
		}

		tm.saveTemplatesToFile()
		dialog.ShowInformation("Success", "Template imported", tm.window)
	}, tm.window)
}

// saveTemplatesToFile saves templates to a JSON file
func (tm *TemplateManager) saveTemplatesToFile() {
	// Get user data directory
	userDataDir := fyne.CurrentApp().Storage().RootURI().Path()
	templatesFile := filepath.Join(userDataDir, "templates.json")

	// Create directory if it doesn't exist
	os.MkdirAll(filepath.Dir(templatesFile), 0755)

	// Convert to JSON
	data, err := json.MarshalIndent(tm.templates, "", "  ")
	if err != nil {
		return
	}

	// Write to file
	os.WriteFile(templatesFile, data, 0644)
}

// loadTemplates loads templates from a JSON file
func (tm *TemplateManager) loadTemplates() {
	// Get user data directory
	userDataDir := fyne.CurrentApp().Storage().RootURI().Path()
	templatesFile := filepath.Join(userDataDir, "templates.json")

	// Check if file exists
	if _, err := os.Stat(templatesFile); os.IsNotExist(err) {
		return
	}

	// Read file
	data, err := os.ReadFile(templatesFile)
	if err != nil {
		return
	}

	// Parse JSON
	json.Unmarshal(data, &tm.templates)
}
