package gin

import (
	"html/template"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin/render"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

type Delims struct {
	Left  string
	Right string
}

type RenderOptions struct {
	Delims  *Delims
	FuncMap template.FuncMap
}

var defineRe = regexp.MustCompile(`\{\{\s*define\s+"([^"]+)"`)

func extractDefineName(content string) string {
	matches := defineRe.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func NewTemplate(options RenderOptions, extraViews []string) (*render.HTMLProduction, error) {
	instance := template.New("")
	if options.Delims != nil {
		instance.Delims(options.Delims.Left, options.Delims.Right)
	}
	if options.FuncMap != nil {
		instance.Funcs(options.FuncMap)
	}

	appDefines := make(map[string]string)
	pkgDefines := make(map[string]string)
	var files []string

	dir := path.Resource("views")
	if file.Exists(dir) {
		if err := filepath.Walk(dir, func(fullPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				content, readErr := os.ReadFile(fullPath)
				if readErr != nil {
					return readErr
				}
				name := extractDefineName(string(content))
				if name != "" {
					appDefines[name] = fullPath
				}
				files = append(files, fullPath)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	for _, dir := range extraViews {
		if !file.Exists(dir) {
			continue
		}
		if err := filepath.Walk(dir, func(fullPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				content, readErr := os.ReadFile(fullPath)
				if readErr != nil {
					return readErr
				}
				name := extractDefineName(string(content))
				if name == "" {
					files = append(files, fullPath)
					return nil
				}
				if _, ok := appDefines[name]; ok {
					return nil
				}
				if prevFile, ok := pkgDefines[name]; ok {
					LogFacade.Warningf("view collision: %q defined in %q and %q, using first", name, prevFile, fullPath)
					return nil
				}
				pkgDefines[name] = fullPath
				files = append(files, fullPath)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	if len(files) == 0 {
		return nil, nil
	}

	tmpl := template.Must(instance.ParseFiles(files...))

	return &render.HTMLProduction{Template: tmpl}, nil
}

func DefaultTemplate(extraViews []string) (*render.HTMLProduction, error) {
	return NewTemplate(RenderOptions{}, extraViews)
}
