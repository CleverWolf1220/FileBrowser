package file

import (
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/utils/errors"
)

// Info contains the information about a particular file or directory
type Info struct {
	Name        string
	Size        int64
	URL         string
	ModTime     time.Time
	Mode        os.FileMode
	IsDir       bool
	Path        string // Relative path to Caddyfile
	VirtualPath string // Relative path to u.FileSystem
	Mimetype    string
	Content     []byte
	Type        string
	UserAllowed bool // Indicates if the user has enough permissions
}

// GetInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func GetInfo(url *url.URL, c *config.Config, u *config.User) (*Info, int, error) {
	var err error

	i := &Info{URL: c.PrefixURL + url.Path}
	i.VirtualPath = strings.Replace(url.Path, c.BaseURL, "", 1)
	i.VirtualPath = strings.TrimPrefix(i.VirtualPath, "/")
	i.VirtualPath = "/" + i.VirtualPath

	i.Path = u.Scope + i.VirtualPath
	i.Path = strings.Replace(i.Path, "\\", "/", -1)
	i.Path = filepath.Clean(i.Path)

	info, err := os.Stat(i.Path)
	if err != nil {
		return i, errors.ErrorToHTTPCode(err, false), err
	}

	i.Name = info.Name()
	i.ModTime = info.ModTime()
	i.Mode = info.Mode()
	i.IsDir = info.IsDir()
	i.Size = info.Size()

	return i, 0, nil
}

// RetrieveFileType obtains the mimetype and a simplified internal Type
// using the first 512 bytes from the file.
func (i *Info) RetrieveFileType() error {
	i.Mimetype = mime.TypeByExtension(filepath.Ext(i.Name))

	if i.Mimetype == "" {
		err := i.Read()
		if err != nil {
			return err
		}

		i.Mimetype = http.DetectContentType(i.Content)
	}

	i.Type = simplifyMediaType(i.Mimetype)
	return nil
}

// Reads the file.
func (i *Info) Read() error {
	if len(i.Content) != 0 {
		return nil
	}

	var err error
	i.Content, err = ioutil.ReadFile(i.Path)
	if err != nil {
		return err
	}
	return nil
}

// StringifyContent returns the string version of Raw
func (i Info) StringifyContent() string {
	return string(i.Content)
}

// HumanSize returns the size of the file as a human-readable string
// in IEC format (i.e. power of 2 or base 1024).
func (i Info) HumanSize() string {
	return humanize.IBytes(uint64(i.Size))
}

// HumanModTime returns the modified time of the file as a human-readable string.
func (i Info) HumanModTime(format string) string {
	return i.ModTime.Format(format)
}

// CanBeEdited checks if the extension of a file is supported by the editor
func (i Info) CanBeEdited() bool {
	if i.Type == "text" {
		return true
	}

	// If the type isn't text (and is blob for example), it will check some
	// common types that are mistaken not to be text.
	extensions := [...]string{
		".md", ".markdown", ".mdown", ".mmark",
		".asciidoc", ".adoc", ".ad",
		".rst",
		".json", ".toml", ".yaml", ".csv", ".xml", ".rss", ".conf", ".ini",
		".tex", ".sty",
		".css", ".sass", ".scss",
		".js",
		".html",
		".txt", ".rtf",
		".sh", ".bash", ".ps1", ".bat", ".cmd",
		".php", ".pl", ".py",
		"Caddyfile",
		".c", ".cc", ".h", ".hh", ".cpp", ".hpp", ".f90",
		".f", ".bas", ".d", ".ada", ".nim", ".cr", ".java", ".cs", ".vala", ".vapi",
	}

	for _, extension := range extensions {
		if strings.HasSuffix(i.Name, extension) {
			return true
		}
	}

	return false
}

func simplifyMediaType(name string) string {
	if strings.HasPrefix(name, "video") {
		return "video"
	}

	if strings.HasPrefix(name, "audio") {
		return "audio"
	}

	if strings.HasPrefix(name, "image") {
		return "image"
	}

	if strings.HasPrefix(name, "text") {
		return "text"
	}

	if strings.HasPrefix(name, "application/javascript") {
		return "text"
	}

	return "blob"
}
