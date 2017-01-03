package handlers

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/file"
	"github.com/hacdias/caddy-filemanager/frontmatter"
	"github.com/spf13/hugo/parser"
)

// Editor contains the information for the editor page
type Editor struct {
	Class       string
	Mode        string
	Content     string
	FrontMatter *frontmatter.Content
}

// GetEditor gets the editor based on a FileInfo struct
func GetEditor(i *file.Info) (*Editor, error) {
	// Create a new editor variable and set the mode
	editor := new(Editor)
	editor.Mode = strings.TrimPrefix(filepath.Ext(i.Name), ".")

	switch editor.Mode {
	case "md", "markdown", "mdown", "mmark":
		editor.Mode = "markdown"
	case "asciidoc", "adoc", "ad":
		editor.Mode = "asciidoc"
	case "rst":
		editor.Mode = "rst"
	case "html", "htm":
		editor.Mode = "html"
	case "js":
		editor.Mode = "javascript"
	case "go":
		editor.Mode = "golang"
	}

	var page parser.Page
	var err error

	// Handle the content depending on the file extension
	switch editor.Mode {
	case "json", "toml", "yaml":
		// Defines the class and declares an error
		editor.Class = "frontmatter-only"

		// Checks if the file already has the frontmatter rune and parses it
		if frontmatter.HasRune(i.Content) {
			editor.FrontMatter, _, err = frontmatter.Pretty(i.Content)
		} else {
			editor.FrontMatter, _, err = frontmatter.Pretty(frontmatter.AppendRune(i.Content, editor.Mode))
		}

		// Check if there were any errors
		if err == nil {
			break
		}

		fallthrough
	case "markdown", "asciidoc", "rst":
		if frontmatter.HasRune(i.Content) {
			// Starts a new buffer and parses the file using Hugo's functions
			buffer := bytes.NewBuffer(i.Content)
			page, err = parser.ReadFrom(buffer)
			editor.Class = "complete"

			if err == nil {
				// Parses the page content and the frontmatter
				editor.Content = strings.TrimSpace(string(page.Content()))
				editor.FrontMatter, _, err = frontmatter.Pretty(page.FrontMatter())

				if err == nil {
					break
				}
			}
		}

		fallthrough
	default:
		editor.Class = "content-only"
		editor.Content = i.StringifyContent()
	}

	return editor, nil
}
