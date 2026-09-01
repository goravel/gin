package gin

import (
	"fmt"
	"html/template"
	"io/fs"
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

var (
	defineRe = regexp.MustCompile(`\{\{\s*define\s+"([^"]+)"`)
)

// common type for view sources, wrapping an fs.FS and providing collision warning display.
type viewSource struct {
	fsys          fs.FS
	display       func(name string) string
	isAppResource bool
}

func extractDefineName(content string, leftDelim string) string {
	var re *regexp.Regexp
	if leftDelim == "" || leftDelim == "{{" {
		re = defineRe
	} else {
		re = regexp.MustCompile(regexp.QuoteMeta(leftDelim) + `\s*define\s+"([^"]+)"`)
	}
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func NewTemplate(options RenderOptions) (*render.HTMLProduction, error) {
	instance := template.New("")
	if options.Delims != nil {
		instance.Delims(options.Delims.Left, options.Delims.Right)
	}
	if options.FuncMap != nil {
		instance.Funcs(options.FuncMap)
	}

	leftDelim := "{{"
	if options.Delims != nil {
		leftDelim = options.Delims.Left
	}

	appDefines := make(map[string]string)
	pkgDefines := make(map[string]string)
	loaded := false

	// Precedence: application views > package directories > package filesystems,
	// each in registration order.
	for _, source := range viewSources() {
		files, err := collectTemplates(source, leftDelim, appDefines, pkgDefines)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			continue
		}
		if _, err := instance.ParseFS(source.fsys, files...); err != nil {
			return nil, err
		}
		loaded = true
	}

	if !loaded {
		return nil, nil
	}

	return &render.HTMLProduction{Template: instance}, nil
}

func DefaultTemplate() (*render.HTMLProduction, error) {
	return NewTemplate(RenderOptions{})
}

// returns every existing template source in precedence order.
func viewSources() []viewSource {
	var sources []viewSource

	if dir := path.Resource("views"); file.Exists(dir) {
		sources = append(sources, dirSource(dir, true))
	}

	if ViewFacade == nil {
		return sources
	}

	for _, dir := range ViewFacade.RegisteredViews() {
		if file.Exists(dir) {
			sources = append(sources, dirSource(dir, false))
		}
	}

	for i, fsys := range ViewFacade.RegisteredViewFS() {
		if _, err := fs.Stat(fsys, "."); err != nil {
			continue
		}
		sources = append(sources, viewSource{
			fsys: fsys,
			display: func(name string) string {
				return fmt.Sprintf("fs[%d]/%s", i, name)
			},
		})
	}

	return sources
}

func dirSource(dir string, isAppResource bool) viewSource {
	return viewSource{
		fsys: os.DirFS(dir),
		display: func(name string) string {
			return filepath.Join(dir, filepath.FromSlash(name))
		},
		isAppResource: isAppResource,
	}
}

func collectTemplates(source viewSource, leftDelim string, appDefines, pkgDefines map[string]string) ([]string, error) {
	var files []string

	err := fs.WalkDir(source.fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		content, readErr := fs.ReadFile(source.fsys, name)
		if readErr != nil {
			return readErr
		}

		defineName := extractDefineName(string(content), leftDelim)
		if defineName == "" {
			files = append(files, name)
			return nil
		}

		fullPath := source.display(name)
		if source.isAppResource {
			appDefines[defineName] = fullPath
			files = append(files, name)
			return nil
		}

		if _, ok := appDefines[defineName]; ok {
			return nil
		}
		if prevFile, ok := pkgDefines[defineName]; ok {
			if LogFacade != nil {
				LogFacade.Warningf("view collision: %q defined in %q and %q, using first", defineName, prevFile, fullPath)
			}
			return nil
		}

		pkgDefines[defineName] = fullPath
		files = append(files, name)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
