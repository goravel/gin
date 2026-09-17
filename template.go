package gin

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	stdpath "path"
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

// viewTier is the precedence class of a view source. Lower-numbered tiers take
// precedence, in the order tierApp, tierDir, tierFS.
type viewTier int

const (
	tierApp viewTier = iota
	tierDir
	tierFS
)

// viewSource is a filesystem that contributes templates at a given precedence tier.
type viewSource struct {
	fsys fs.FS
	tier viewTier
	// label identifies the source in warnings: the directory path for tierApp
	// and tierDir, "fs[i]" (i being the LoadViewsFromFS registration index) for
	// tierFS.
	label string
}

// pathOf renders name, a slash-separated path inside the source, for display.
func (s viewSource) pathOf(name string) string {
	if s.tier == tierFS {
		return s.label + "/" + name
	}
	return filepath.Join(s.label, filepath.FromSlash(name))
}

// viewDefines tracks the template name each precedence tier has already claimed,
// so the first source to define a name wins.
type viewDefines struct {
	app map[string]string
	pkg map[string]string
}

// claim records templateName for source and reports whether source may contribute
// it. A name already claimed by the application is dropped silently; one already
// claimed by an earlier package source is dropped with a warning.
func (d *viewDefines) claim(source viewSource, name, templateName string) bool {
	fullPath := source.pathOf(name)

	if source.tier == tierApp {
		d.app[templateName] = fullPath
		return true
	}
	if _, ok := d.app[templateName]; ok {
		return false
	}
	if prevFile, ok := d.pkg[templateName]; ok {
		if LogFacade != nil {
			LogFacade.Warningf("view collision: %q defined in %q and %q, using first", templateName, prevFile, fullPath)
		}
		return false
	}

	d.pkg[templateName] = fullPath
	return true
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

	defines := &viewDefines{app: make(map[string]string), pkg: make(map[string]string)}
	loaded := false

	for _, source := range viewSources() {
		contributed, err := loadSource(instance, source, leftDelim, defines)
		if err != nil {
			return nil, err
		}
		loaded = loaded || contributed
	}

	// No views found (neither app resources/views nor registered package views or
	// filesystems): return (nil, nil) so callers keep their current renderer. For
	// the deferred default-template path this means the template set is compiled
	// exactly once at first serve — views registered later are not picked up.
	if !loaded {
		return nil, nil
	}

	return &render.HTMLProduction{Template: instance}, nil
}

func DefaultTemplate() (*render.HTMLProduction, error) {
	return NewTemplate(RenderOptions{})
}

// viewSources returns every existing template source in precedence order: the
// application's resources/views, then directories registered via LoadViewsFrom,
// then filesystems registered via LoadViewsFromFS, each in registration order.
func viewSources() []viewSource {
	var sources []viewSource

	if dir := path.Resource("views"); file.Exists(dir) {
		sources = append(sources, viewSource{fsys: os.DirFS(dir), tier: tierApp, label: dir})
	}

	viewFacade := ViewFacade
	if viewFacade == nil && App != nil {
		viewFacade = App.MakeView()
	}
	if viewFacade == nil {
		return sources
	}

	for _, dir := range viewFacade.RegisteredViews() {
		if file.Exists(dir) {
			sources = append(sources, viewSource{fsys: os.DirFS(dir), tier: tierDir, label: dir})
		}
	}

	// Labels keep the registration index, so a skipped filesystem still consumes
	// its slot and warnings point at the position the package registered.
	for i, fsys := range viewFacade.RegisteredViewFS() {
		if fsys == nil {
			if LogFacade != nil {
				LogFacade.Warningf("view source fs[%d] is nil, skipping", i)
			}
			continue
		}
		if _, err := fs.Stat(fsys, "."); err != nil {
			if LogFacade != nil {
				LogFacade.Warningf("view source fs[%d] is unreadable, skipping: %v", i, err)
			}
			continue
		}
		sources = append(sources, viewSource{fsys: fsys, tier: tierFS, label: fmt.Sprintf("fs[%d]", i)})
	}

	return sources
}

// loadSource walks source and parses every template it contributes into
// instance, reporting whether it contributed any. Each file is read once and
// parsed immediately, so only one file's content is held at a time. A file with
// no define block is claimed under its base name, the name it is parsed as, so it
// cannot override a same-named template from a higher-precedence source.
func loadSource(instance *template.Template, source viewSource, leftDelim string, defines *viewDefines) (bool, error) {
	contributed := false

	err := fs.WalkDir(source.fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		content, err := fs.ReadFile(source.fsys, name)
		if err != nil {
			return err
		}
		text := string(content)

		templateName := extractDefineName(text, leftDelim)
		if templateName == "" {
			templateName = stdpath.Base(name)
		}
		if !defines.claim(source, name, templateName) {
			return nil
		}

		// Mirrors html/template.ParseFS: every file becomes an associated template
		// named after its base. The content is parsed directly instead of going
		// through ParseFS, which would re-read the file and — because it resolves
		// names through fs.Glob — mangle or fail on any name containing "[", "*"
		// or "?".
		if _, err := instance.New(stdpath.Base(name)).Parse(text); err != nil {
			return err
		}
		contributed = true

		return nil
	})
	if err != nil {
		return false, err
	}

	return contributed, nil
}
