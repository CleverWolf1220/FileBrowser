package filemanager

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"golang.org/x/net/webdav"
)

var (
	// ErrDuplicated occurs when you try to create a user that already exists.
	ErrDuplicated = errors.New("Duplicated user")
)

// FileManager is a file manager instance. It should be creating using the
// 'New' function and not directly.
type FileManager struct {
	*User
	Assets *assets

	// BaseURL is the path where the GUI will be accessible. It musn't end with
	// a trailing slash and mustn't contain PrefixURL, if set. Despite being
	// exported, it should only be manipulated using SetBaseURL function.
	BaseURL string

	// WebDavURL is the path where the WebDAV will be accessible. It can be set to ""
	// in order to override the GUI and only use the WebDAV. It musn't end with
	// a trailing slash. Despite being exported, it should only be manipulated
	// using SetWebDavURL function.
	WebDavURL string

	// PrefixURL is a part of the URL that is trimmed from the http.Request.URL before
	// it arrives to our handlers. It may be useful when using FileManager as a middleware
	// such as in caddy-filemanager plugin. It musn't end with a trailing slash.
	PrefixURL string

	// Users is a map with the different configurations for each user.
	Users map[string]*User

	// BeforeSave is a function that is called before saving a file.
	BeforeSave Command

	// AfterSave is a function that is called before saving a file.
	AfterSave Command
}

// Command is a command function.
type Command func(r *http.Request, m *FileManager, u *User) error

// User contains the configuration for each user. It should be created
// using NewUser on a File Manager instance.
type User struct {
	// scope is the physical path the user has access to.
	scope string

	// fileSystem is the virtual file system the user has access.
	fileSystem webdav.FileSystem

	// handler handles incoming requests to the WebDAV backend.
	handler *webdav.Handler

	// Rules is an array of access and deny rules.
	Rules []*Rule `json:"-"`

	// TODO: this MUST be done in another way
	StyleSheet string `json:"-"`

	// These indicate if the user can perform certain actions.
	AllowNew      bool // Create files and folders
	AllowEdit     bool // Edit/rename files
	AllowCommands bool // Execute commands

	// Commands is the list of commands the user can execute.
	Commands []string
}

// assets are the static and front-end assets, such as JS, CSS and HTML templates.
type assets struct {
	requiredJS *rice.Box // JS that is always required to have in order to be usable.
	Templates  *rice.Box
	CSS        *rice.Box
	JS         *rice.Box
}

// Rule is a dissalow/allow rule.
type Rule struct {
	// Regex indicates if this rule uses Regular Expressions or not.
	Regex bool

	// Allow indicates if this is an allow rule. Set 'false' to be a disallow rule.
	Allow bool

	// Path is the corresponding URL path for this rule.
	Path string

	// Regexp is the regular expression. Only use this when 'Regex' was set to true.
	Regexp *regexp.Regexp
}

// New creates a new File Manager instance with the needed
// configuration to work.
func New(scope string) *FileManager {
	m := &FileManager{
		User: &User{
			AllowCommands: true,
			AllowEdit:     true,
			AllowNew:      true,
			Commands:      []string{},
			Rules:         []*Rule{},
		},
		Users:      map[string]*User{},
		BeforeSave: func(r *http.Request, m *FileManager, u *User) error { return nil },
		AfterSave:  func(r *http.Request, m *FileManager, u *User) error { return nil },
		Assets: &assets{
			Templates:  rice.MustFindBox("./_assets/templates"),
			CSS:        rice.MustFindBox("./_assets/css"),
			requiredJS: rice.MustFindBox("./_assets/js"),
		},
	}

	m.SetScope(scope, "")
	m.SetBaseURL("/")
	m.SetWebDavURL("/webdav")

	return m
}

// AbsoluteURL returns the actual URL where
// File Manager interface can be accessed.
func (m FileManager) AbsoluteURL() string {
	return m.PrefixURL + m.BaseURL
}

// AbsoluteWebDavURL returns the actual URL
// where WebDAV can be accessed.
func (m FileManager) AbsoluteWebDavURL() string {
	return m.PrefixURL + m.WebDavURL
}

// SetBaseURL updates the BaseURL of a File Manager
// object.
func (m *FileManager) SetBaseURL(url string) {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	url = "/" + url
	m.BaseURL = strings.TrimSuffix(url, "/")
}

// SetWebDavURL updates the WebDavURL of a File Manager
// object and updates it's main handler.
func (m *FileManager) SetWebDavURL(url string) {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")

	m.WebDavURL = m.BaseURL + "/" + url

	// update base user webdav handler
	m.handler = &webdav.Handler{
		Prefix:     m.WebDavURL,
		FileSystem: m.fileSystem,
		LockSystem: webdav.NewMemLS(),
	}

	// update other users' handlers to match
	// the new URL
	for _, u := range m.Users {
		u.handler = &webdav.Handler{
			Prefix:     m.WebDavURL,
			FileSystem: u.fileSystem,
			LockSystem: webdav.NewMemLS(),
		}
	}
}

// SetScope updates a user scope and its virtual file system.
// If the user string is blank, it will change the base scope.
func (m *FileManager) SetScope(scope string, username string) error {
	var u *User

	if username == "" {
		u = m.User
	} else {
		var ok bool
		u, ok = m.Users[username]
		if !ok {
			return errors.New("Inexistent user")
		}
	}

	u.scope = strings.TrimSuffix(scope, "/")
	u.fileSystem = webdav.Dir(u.scope)

	u.handler = &webdav.Handler{
		Prefix:     m.WebDavURL,
		FileSystem: u.fileSystem,
		LockSystem: webdav.NewMemLS(),
	}

	return nil
}

// NewUser creates a new user on a File Manager struct
// which inherits its configuration from the main user.
func (m *FileManager) NewUser(username string) error {
	if _, ok := m.Users[username]; ok {
		return ErrDuplicated
	}

	m.Users[username] = &User{
		scope:         m.User.scope,
		fileSystem:    m.User.fileSystem,
		handler:       m.User.handler,
		Rules:         m.User.Rules,
		AllowNew:      m.User.AllowNew,
		AllowEdit:     m.User.AllowEdit,
		AllowCommands: m.User.AllowCommands,
		Commands:      m.User.Commands,
	}

	return nil
}

// Allowed checks if the user has permission to access a directory/file.
func (u User) Allowed(url string) bool {
	var rule *Rule
	i := len(u.Rules) - 1

	for i >= 0 {
		rule = u.Rules[i]

		if rule.Regex {
			if rule.Regexp.MatchString(url) {
				return rule.Allow
			}
		} else if strings.HasPrefix(url, rule.Path) {
			return rule.Allow
		}

		i--
	}

	return true
}
