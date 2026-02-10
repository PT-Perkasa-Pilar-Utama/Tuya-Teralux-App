package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type DocsController struct {
	DocsRoot string
	md       goldmark.Markdown
}

func NewDocsController() *DocsController {
	// Initialize Goldmark with professional extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(), // Allow raw HTML in markdown if needed
		),
	)

	return &DocsController{
		DocsRoot: "./docs",
		md:       md,
	}
}

type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
}

func (ctrl *DocsController) ServeDocs(c *gin.Context) {
	requestPath := c.Param("path")
	if requestPath == "" || requestPath == "/" {
		if _, err := os.Stat(filepath.Join(ctrl.DocsRoot, "manual", "installation_guide.md")); err == nil {
			c.Redirect(http.StatusFound, "/docs/manual/installation_guide.md")
			return
		}
	}

	// Build Sidebar Tree
	tree, err := ctrl.buildFileTree(ctrl.DocsRoot)
	if err != nil {
		utils.LogError("Failed to build file tree: %v", err)
		c.String(http.StatusInternalServerError, "Failed to build file tree")
		return
	}

	// Serve Static Assets
	if strings.HasPrefix(requestPath, "/assets/") {
		// Map /docs/assets/... to ./assets/docs/...
		assetPath := filepath.Join("assets", "docs", strings.TrimPrefix(requestPath, "/assets/"))
		c.File(assetPath)
		return
	}

	fullPath := filepath.Join(ctrl.DocsRoot, requestPath)

	// Security: Prevent directory traversal
	rel, err := filepath.Rel(ctrl.DocsRoot, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		c.String(http.StatusForbidden, "Access Forbidden")
		return
	}

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File Not Found")
		return
	}

	var content template.HTML
	var title string = "Documentation"

	if !info.IsDir() {
		if strings.HasSuffix(info.Name(), ".md") {
			data, err := os.ReadFile(fullPath)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to read file")
				return
			}

			// Transcribe Markdown to HTML using Goldmark
			var buf bytes.Buffer
			if err := ctrl.md.Convert(data, &buf); err != nil {
				c.String(http.StatusInternalServerError, "Failed to render markdown")
				return
			}

			content = template.HTML(buf.String())
			title = strings.TrimSuffix(info.Name(), ".md")
		} else {
			c.File(fullPath)
			return
		}
	} else {
		content = template.HTML("<h3>Select a document from the sidebar</h3>")
	}

	// Render using professional View folder
	tmpl := template.New("layout.html").Funcs(template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	})

	tmpl, err = tmpl.ParseGlob("/home/farismnrr/Documents/Programs/teralux_app/backend/views/docs/*.html")
	if err != nil {
		utils.LogError("Template loading error: %v", err)
		c.String(http.StatusInternalServerError, "Template error")
		return
	}

	data := struct {
		Title       string
		Content     template.HTML
		Tree        *FileNode
		CurrentPath string
	}{
		Title:       title,
		Content:     content,
		Tree:        tree,
		CurrentPath: strings.TrimPrefix(requestPath, "/"),
	}

	if err := tmpl.ExecuteTemplate(c.Writer, "layout.html", data); err != nil {
		utils.LogError("Template execution error: %v", err)
		c.String(http.StatusInternalServerError, "Rendering error")
	}
}

func (ctrl *DocsController) buildFileTree(root string) (*FileNode, error) {
	node := &FileNode{
		Name:     "Docs",
		IsDir:    true,
		Children: []*FileNode{},
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}

		// Skip swagger and assets folder
		if d.IsDir() {
			if d.Name() == "swagger" || d.Name() == "assets" {
				return filepath.SkipDir
			}
		}

		relPath, _ := filepath.Rel(root, path)
		parts := strings.Split(relPath, string(os.PathSeparator))

		current := node
		for i, part := range parts {
			var child *FileNode
			for _, c := range current.Children {
				if c.Name == part {
					child = c
					break
				}
			}
			if child == nil {
				// We join parts up to current index for the correct relative path
				childPath := filepath.Join(parts[:i+1]...)
				child = &FileNode{
					Name:     part,
					Path:     childPath,
					IsDir:    d.IsDir() && i == len(parts)-1,
					Children: []*FileNode{},
				}
				if i < len(parts)-1 {
					child.IsDir = true
				}

				current.Children = append(current.Children, child)
			}
			current = child
		}
		return nil
	})

	return node, err
}
