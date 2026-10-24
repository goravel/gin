package gin

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin/render"
	contractsview "github.com/goravel/framework/contracts/view"
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

// claim records templateNames for source and reports whether source may contribute
// the file. A name already claimed by the application is dropped silently; one
// already claimed by an earlier package source is dropped with a warning. Nothing
// is recorded unless every name is free, so a skipped file never blocks a later one.
func (d *viewDefines) claim(source viewSource, name string, templateNames ...string) bool {
	fullPath := source.pathOf(name)

	if source.tier == tierApp {
		for _, templateName := range templateNames {
			d.app[templateName] = fullPath
		}
		return true
	}
	for _, templateName := range templateNames {
		if _, ok := d.app[templateName]; ok {
			return false
		}
	}
	for _, templateName := range templateNames {
		if prevFile, ok := d.pkg[templateName]; ok {
			if LogFacade != nil {
				LogFacade.Warningf("view collision: %q defined in %q and %q, using first", templateName, prevFile, fullPath)
			}
			return false
		}
	}

	for _, templateName := range templateNames {
		d.pkg[templateName] = fullPath
	}
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

// DefaultTemplate compiles every registered view source with the functions
// registered on the view facade through View.AddFuncs, which is also where
// ApplicationBuilder.WithViewFuncs puts the application's functions.
// Custom delimiters still require the http.drivers.<driver>.template callback.
func DefaultTemplate() (*render.HTMLProduction, error) {
	options := RenderOptions{}
	if viewFacade := resolveViewFacade(); viewFacade != nil {
		options.FuncMap = viewFacade.Funcs()
	}

	return NewTemplate(options)
}

// resolveViewFacade returns the view facade ServiceProvider.Boot assigned, or
// resolves it from the container when the driver is used before Boot.
func resolveViewFacade() contractsview.View {
	if ViewFacade != nil {
		return ViewFacade
	}
	if App != nil {
		return App.MakeView()
	}

	return nil
}

// viewSources returns every existing template source in precedence order: the
// application's resources/views, then directories registered via LoadViewsFrom,
// then filesystems registered via LoadViewsFromFS, each in registration order.
// A path that is missing or is not a directory is skipped: walking it through
// os.DirFS would fail the whole compile.
func viewSources() []viewSource {
	var sources []viewSource

	if dir := path.Resource("views"); isDir(dir) {
		sources = append(sources, viewSource{fsys: os.DirFS(dir), tier: tierApp, label: dir})
	}

	viewFacade := resolveViewFacade()
	if viewFacade == nil {
		return sources
	}

	for _, dir := range viewFacade.RegisteredViews() {
		if isDir(dir) {
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

func isDir(dir string) bool {
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// loadSource walks source and parses every template it contributes into
// instance, reporting whether it contributed any. Each file is read once and
// parsed immediately, so only one file's content is held at a time. Every file is
// parsed under its slash-separated path relative to the source root, so files
// sharing a base name in different directories stay distinct. A file with no
// define block is claimed under that path, so it cannot override a template of
// the same path from a higher-precedence source. Errors name the source they came
// from: os.DirFS reports paths relative to the directory, and an fs.FS has no
// path of its own.
func loadSource(instance *template.Template, source viewSource, leftDelim string, defines *viewDefines) (bool, error) {
	contributed := false

	err := fs.WalkDir(source.fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk view source %s: %w", source.pathOf(name), err)
		}
		if d.IsDir() {
			return nil
		}

		content, err := fs.ReadFile(source.fsys, name)
		if err != nil {
			return fmt.Errorf("failed to read view %s: %w", source.pathOf(name), err)
		}
		text := string(content)

		// A file contributes two names: the one its define block declares, if any,
		// and its own, the slash-separated path relative to the source root. Both
		// must be free in every higher-precedence source, otherwise the whole file
		// is skipped: html/template keeps one template per name, so parsing the
		// file would replace the template already registered under its path. The
		// define name is checked first so a collision is reported by the name
		// templates are rendered with.
		var templateNames []string
		if defineName := extractDefineName(text, leftDelim); defineName != "" && defineName != name {
			templateNames = append(templateNames, defineName)
		}
		templateNames = append(templateNames, name)
		if !defines.claim(source, name, templateNames...) {
			return nil
		}

		// Unlike html/template.ParseFS, which names each file after its base, the
		// file is named after its relative path. The content is parsed directly
		// instead of going through ParseFS, which would re-read the file and —
		// because it resolves names through fs.Glob — mangle or fail on any name
		// containing "[", "*" or "?".
		if _, err := instance.New(name).Parse(text); err != nil {
			return fmt.Errorf("failed to parse view %s: %w", source.pathOf(name), err)
		}
		contributed = true

		return nil
	})
	if err != nil {
		return false, err
	}

	return contributed, nil
}
