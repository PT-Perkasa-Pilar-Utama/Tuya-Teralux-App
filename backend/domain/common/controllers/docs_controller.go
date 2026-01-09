package controllers

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type DocsController struct {
	DocsRoot string
}

func NewDocsController() *DocsController {
	// Assuming docs are in backend/docs relative to execution
	// Adjust path if needed based on where binary runs
	return &DocsController{
		DocsRoot: "./docs",
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
		// Default to first md file or README if root requested
		// For now, let's redirect to manual/installation_guide.md if exists
		if _, err := os.Stat(filepath.Join(ctrl.DocsRoot, "manual", "installation_guide.md")); err == nil {
			c.Redirect(http.StatusFound, "/docs/manual/installation_guide.md")
			return
		}
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

	// Build Sidebar Tree
	tree, err := ctrl.buildFileTree(ctrl.DocsRoot)
	if err != nil {
		utils.LogError("Failed to build file tree: %v", err)
		c.String(http.StatusInternalServerError, "Failed to build file tree")
		return
	}

	var content template.HTML
	var title string = "Documentation"

	if !info.IsDir() {
		if strings.HasSuffix(info.Name(), ".md") {
			// Read file content
			data, err := os.ReadFile(fullPath)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to read file")
				return
			}

			// Convert Markdown to HTML
			extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
			p := parser.NewWithExtensions(extensions)
			doc := p.Parse(data)

			htmlFlags := html.CommonFlags | html.HrefTargetBlank
			opts := html.RendererOptions{Flags: htmlFlags}
			renderer := html.NewRenderer(opts)

			content = template.HTML(markdown.Render(doc, renderer))
			title = strings.TrimSuffix(info.Name(), ".md")
		} else {
			// Serve raw file for assets (images, etc.)
			c.File(fullPath)
			return
		}
	} else {
		content = template.HTML("<h3>Select a document from the sidebar</h3>")
	}

	// Render HTML
	// We'll use an inline template for simplicity, but normally this goes to a file
	tmplStr := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Teralux Docs - {{ .Title }}</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/5.2.0/github-markdown.min.css">
    <style>
        :root {
            --bg-color: #ffffff;
            --text-color: #24292f;
            --sidebar-bg: #f6f8fa;
            --sidebar-border: #d0d7de;
            --link-color: #0969da;
            --hover-bg: #e6f0ff;
            --active-bg: #ddf4ff;
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --bg-color: #0d1117;
                --text-color: #c9d1d9;
                --sidebar-bg: #010409;
                --sidebar-border: #30363d;
                --link-color: #58a6ff;
                --hover-bg: #1f6feb; /* Darker blue for hover */
                --active-bg: #1f6feb;
            }
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji";
            margin: 0;
            display: flex;
            height: 100vh;
            background-color: var(--bg-color);
            color: var(--text-color);
        }

        .sidebar {
            width: 280px;
            background-color: var(--sidebar-bg);
            border-right: 1px solid var(--sidebar-border);
            overflow-y: auto;
            padding: 24px 16px;
            position: fixed;
            left: 0;
            top: 0;
            bottom: 0;
            flex-shrink: 0;
            display: flex;
            flex-direction: column;
        }

        .sidebar h2 {
            font-size: 1.25rem;
            margin-top: 0;
            margin-bottom: 24px;
            padding-bottom: 16px;
            border-bottom: 1px solid var(--sidebar-border);
            display: flex;
            align-items: center;
            gap: 8px;
            color: var(--text-color);
        }

        .content {
            flex-grow: 1;
            overflow-y: auto;
            padding: 48px 48px 120px 48px; /* Added extra bottom padding */
            display: flex;
            justify-content: center;
        }

        .markdown-body {
            box-sizing: border-box;
            min-width: 200px;
            max-width: 980px;
            width: 100%;
            margin: 0 auto;
            background-color: transparent !important; /* Let body bg shine through for consistency */
        }

        ul { list-style-type: none; padding-left: 0; margin: 0; }
        .sidebar ul ul { padding-left: 16px; margin-top: 4px; }
        
        li { margin-bottom: 2px; }

        .file-link { 
            text-decoration: none; 
            color: var(--text-color); 
            display: block; 
            padding: 6px 10px; 
            border-radius: 6px;
            font-size: 14px;
            line-height: 1.5;
            transition: background 0.1s ease;
        }
        
        .file-link:hover { 
            background-color: var(--sidebar-border); /* Subtle hover */
            text-decoration: none;
        }

        .file-link.active {
            background-color: var(--active-bg);
            color: var(--text-color);
            font-weight: 600;
        }

        .dir-name { 
            font-weight: 600; 
            margin-top: 16px; 
            margin-bottom: 8px; 
            display: block; 
            color: var(--text-color);
            font-size: 13px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            opacity: 0.8;
            padding-left: 10px;
            cursor: pointer;
            user-select: none;
        }
        
        .dir-name:hover {
            opacity: 1;
        }

        /* Scrollbar styling */
        ::-webkit-scrollbar { width: 8px; height: 8px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: var(--sidebar-border); border-radius: 4px; }
        ::-webkit-scrollbar-thumb:hover { background: #8b949e; }

        /* Alerts */
        .markdown-alert {
            padding: 8px 16px;
            margin-bottom: 16px;
            border-left: 0.25em solid;
            background-color: var(--bg-color);
        }
        
        .markdown-alert-title {
            display: flex;
            align-items: center;
            font-weight: 600;
            margin-bottom: 4px;
            margin-top: 0;
            font-size: 14px;
        }

        .markdown-alert-title svg {
            fill: currentColor;
            margin-right: 8px;
        }

        /* Note */
        .markdown-alert-note { border-color: #0969da; }
        .markdown-alert-note .markdown-alert-title { color: #0969da; }
        
        /* Tip */
        .markdown-alert-tip { border-color: #1a7f37; }
        .markdown-alert-tip .markdown-alert-title { color: #1a7f37; }

        /* Important */
        .markdown-alert-important { border-color: #8250df; }
        .markdown-alert-important .markdown-alert-title { color: #8250df; }

        /* Warning */
        .markdown-alert-warning { border-color: #bf8700; }
        .markdown-alert-warning .markdown-alert-title { color: #bf8700; }

        /* Caution */
        .markdown-alert-caution { border-color: #cf222e; }
        .markdown-alert-caution .markdown-alert-title { color: #cf222e; }
        
        /* Dark mode overrides */
        @media (prefers-color-scheme: dark) {
            .markdown-alert-note { border-color: #1f6feb; }
            .markdown-alert-note .markdown-alert-title { color: #2f81f7; }
            .markdown-alert-tip { border-color: #238636; }
            .markdown-alert-tip .markdown-alert-title { color: #3fb950; }
            .markdown-alert-important { border-color: #8957e5; }
            .markdown-alert-important .markdown-alert-title { color: #a371f7; }
            .markdown-alert-warning { border-color: #9e6a03; }
            .markdown-alert-warning .markdown-alert-title { color: #d29922; }
            .markdown-alert-caution { border-color: #da3633; }
            .markdown-alert-caution .markdown-alert-title { color: #f85149; }
        }
    </style>
    <script>
        document.addEventListener('DOMContentLoaded', () => {
            document.querySelectorAll('blockquote').forEach(bq => {
                const p = bq.querySelector('p');
                if (p && p.textContent.trim().startsWith('[!')) {
                    const match = p.textContent.trim().match(/^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]/i);
                    if (match) {
                        const type = match[1].toLowerCase();
                        bq.classList.add('markdown-alert', 'markdown-alert-' + type);
                        
                        const icons = {
                            note: '<svg class="octicon" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>',
                            tip: '<svg class="octicon" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M8 1.5c-2.363 0-4 1.69-4 3.75 0 .984.424 1.625.984 2.304l.214.253c.223.264.47.556.673.848.284.411.537.896.621 1.49a.75.75 0 0 1-1.484.211c-.04-.282-.163-.547-.37-.847a8.456 8.456 0 0 0-.542-.68c-.084-.1-.173-.205-.268-.32C3.201 7.75 2.5 6.766 2.5 5.25 2.5 2.31 4.863 0 8 0s5.5 2.31 5.5 5.25c0 1.516-.701 2.5-1.328 3.259-.095.115-.184.22-.268.319-.207.245-.383.453-.541.681-.208.3-.33.565-.37.847a.75.75 0 0 1-1.485-.212c.084-.593.337-1.078.621-1.489.203-.292.45-.584.673-.848.075-.088.147-.173.213-.253.561-.679.985-1.32.985-2.304 0-2.06-1.637-3.75-4-3.75ZM5.75 12h4.5a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1 0-1.5ZM6 15.25a.75.75 0 0 1 .75-.75h2.5a.75.75 0 0 1 0 1.5h-2.5a.75.75 0 0 1-.75-.75Z"></path></svg>',
                            important: '<svg class="octicon checked" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M0 1.75C0 .784.784 0 1.75 0h12.5C15.216 0 16 .784 16 1.75v9.5A1.75 1.75 0 0 1 14.25 13H8.06l-2.573 2.573A1.458 1.458 0 0 1 3 14.543V13H1.75A1.75 1.75 0 0 1 0 11.25Zm1.75-.25a.25.25 0 0 0-.25.25v9.5c0 .138.112.25.25.25h2a.75.75 0 0 1 .75.75v2.19l2.72-2.72a.75.75 0 0 1 .53-.22h6.5a.25.25 0 0 0 .25-.25v-9.5a.25.25 0 0 0-.25-.25Zm7 2.25v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>',
                            warning: '<svg class="octicon" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 9a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"></path></svg>',
                            caution: '<svg class="octicon" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path d="M4.47.22A.749.749 0 0 1 5 0h6c.199 0 .389.079.53.22l4.25 4.25c.141.14.22.331.22.53v6a.749.749 0 0 1-.22.53l-4.25 4.25A.749.749 0 0 1 11 16H5a.749.749 0 0 1-.53-.22L.22 11.53A.749.749 0 0 1 0 11V5c0-.199.079-.389.22-.53Zm.84 1.28L1.5 5.31v5.38l3.81 3.81h5.38l3.81-3.81V5.31L10.69 1.5ZM8 4a.75.75 0 0 1 .75.75v3.5a.75.75 0 0 1-1.5 0v-3.5A.75.75 0 0 1 8 4Zm0 8a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"></path></svg>'
                        };

                        const titleHTML = '<p class="markdown-alert-title">' + 
                            (icons[type] || '') + 
                            '<span>' + match[1] + '</span></p>';
                        
                        let content = p.innerHTML;
                        content = content.replace(match[0], '').trim();
                        if (content.startsWith('<br>')) content = content.substring(4).trim();
                        
                        p.innerHTML = content;
                        bq.insertAdjacentHTML('afterbegin', titleHTML);
                    }
                }
            });
        });
        
        // Collapsible folders with localStorage persistence
        document.addEventListener('DOMContentLoaded', () => {
            const savedStates = JSON.parse(localStorage.getItem('docsFolderStates') || '{}');
            
            document.querySelectorAll('.dir-name').forEach(dirName => {
                const folderName = dirName.textContent.trim();
                const isCollapsed = savedStates[folderName] === false;
                
                dirName.addEventListener('click', function() {
                    const ul = this.nextElementSibling;
                    if (ul && ul.tagName === 'UL') {
                        const currentlyHidden = ul.style.display === 'none';
                        ul.style.display = currentlyHidden ? 'block' : 'none';
                        this.textContent = currentlyHidden 
                            ? '▼ ' + this.textContent.replace(/^[▼▶] /, '')
                            : '▶ ' + this.textContent.replace(/^[▼▶] /, '');
                        
                        // Save state
                        const states = JSON.parse(localStorage.getItem('docsFolderStates') || '{}');
                        states[this.textContent.replace(/^[▼▶] /, '')] = currentlyHidden;
                        localStorage.setItem('docsFolderStates', JSON.stringify(states));
                    }
                });
                
                // Restore state
                if (!dirName.textContent.match(/^[▼▶] /)) {
                    dirName.textContent = (isCollapsed ? '▶ ' : '▼ ') + dirName.textContent;
                }
                const ul = dirName.nextElementSibling;
                if (ul && ul.tagName === 'UL' && isCollapsed) {
                    ul.style.display = 'none';
                }
            });
        });
        
        // Resizable sidebar with localStorage persistence
        document.addEventListener('DOMContentLoaded', () => {
            const sidebar = document.querySelector('.sidebar');
            const resizer = document.querySelector('.resizer');
            const content = document.querySelector('.content');
            let isResizing = false;
            
            // Restore saved width
            const savedWidth = localStorage.getItem('docsSidebarWidth');
            if (savedWidth) {
                const width = parseInt(savedWidth);
                sidebar.style.width = width + 'px';
                resizer.style.left = width + 'px';
                content.style.marginLeft = width + 'px';
            }
            
            resizer.addEventListener('mousedown', (e) => {
                isResizing = true;
                document.body.style.cursor = 'col-resize';
                document.body.style.userSelect = 'none';
            });
            
            document.addEventListener('mousemove', (e) => {
                if (!isResizing) return;
                const newWidth = e.clientX;
                if (newWidth > 150 && newWidth < 600) {
                    sidebar.style.width = newWidth + 'px';
                    resizer.style.left = newWidth + 'px';
                    content.style.marginLeft = newWidth + 'px';
                }
            });
            
            document.addEventListener('mouseup', () => {
                if (isResizing) {
                    // Save width
                    localStorage.setItem('docsSidebarWidth', sidebar.style.width.replace('px', ''));
                }
                isResizing = false;
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
            });
        });
    </script>
</head>
<body>
    <div class="sidebar">
        <h2>📚 Teralux Docs</h2>
        {{ template "tree" .Tree }}
    </div>
    <div class="resizer" style="width: 4px; background-color: var(--sidebar-border); cursor: col-resize; position: fixed; left: 280px; top: 0; bottom: 0; z-index: 10; transition: background-color 0.2s;"></div>
    <div class="content">
        <article class="markdown-body">
            {{ .Content }}
        </article>
    </div>
</body>
</html>

{{ define "tree" }}
<ul>
    {{ range .Children }}
        <li>
            {{ if .IsDir }}
                <span class="dir-name">{{ .Name }}</span>
                {{ template "tree" . }}
            {{ else }}
                <a href="/docs/{{ .Path }}" class="file-link">{{ .Name }}</a>
            {{ end }}
        </li>
    {{ end }}
</ul>
{{ end }}
`
	t, err := template.New("docs").Parse(tmplStr)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error")
		return
	}

	data := struct {
		Title   string
		Content template.HTML
		Tree    *FileNode
	}{
		Title:   title,
		Content: content,
		Tree:    tree,
	}

	t.Execute(c.Writer, data)
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
			// Find or create child
			var child *FileNode
			for _, c := range current.Children {
				if c.Name == part {
					child = c
					break
				}
			}
			if child == nil {
				child = &FileNode{
					Name:     part,
					Path:     relPath,
					IsDir:    d.IsDir() && i == len(parts)-1, // Only mark as dir if it's the current entry
					Children: []*FileNode{},
				}
				// If it's a directory in the middle of path, it is a directory
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
