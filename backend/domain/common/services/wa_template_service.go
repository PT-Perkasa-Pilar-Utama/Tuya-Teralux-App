package services

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"sensio/domain/common/utils"
)

type WATemplateData struct {
	CustomerName     string
	BuildingName     string
	RoomName         string
	DateStr          string
	BookingTime      string
	Password         string
	RemainingMinutes string
	CompanyName      string
}

type WATemplateService struct {
	templateDir string
}

func NewWATemplateService() *WATemplateService {
	tmplDir := utils.GetAssetPath("templates/wa")
	return &WATemplateService{
		templateDir: tmplDir,
	}
}

func (s *WATemplateService) SetTemplateDir(dir string) {
	s.templateDir = dir
}

func (s *WATemplateService) RenderTemplate(templateName string, data *WATemplateData) (string, error) {
	if data == nil {
		data = &WATemplateData{}
	}

	if data.CustomerName == "" {
		data.CustomerName = "Bapak/Ibu"
	}

	if data.CompanyName == "" {
		data.CompanyName = "[Nama Perusahaan]"
	}

	tmplPath := filepath.Join(s.templateDir, templateName+".md")

	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		return "", fmt.Errorf("template %s not found in %s", templateName, s.templateDir)
	}

	parsedTemplate, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	var buf bytes.Buffer
	if err := parsedTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (s *WATemplateService) GetAvailableTemplates() []string {
	var templates []string
	entries, err := os.ReadDir(s.templateDir)
	if err != nil {
		return templates
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if len(name) > 3 && name[len(name)-3:] == ".md" {
				templates = append(templates, name[:len(name)-3])
			}
		}
	}
	return templates
}
