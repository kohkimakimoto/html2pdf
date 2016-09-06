package html2pdf

import (
	"encoding/json"
	"github.com/kohkimakimoto/html2pdf/resource"
	"github.com/kohkimakimoto/loglv"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type App struct {
	LState         *lua.LState
	LogLevel       string
	variable       map[string]interface{}
	Cachedir       string
	CacheBindir    string
	CacheTmpdir    string
	WkhtmltopdfCmd string
	Targetpdfs     []*TargetPdf
	Tmpfiles       []string
}

func NewApp() *App {
	L := lua.NewState()

	cachedir := filepath.Join(os.TempDir(), "html2pdf_cache")
	cacheBindir := filepath.Join(cachedir, "bin")
	cacheTmpdir := filepath.Join(cachedir, "tmp")

	var wk string
	if runtime.GOOS == "windows" {
		wk = filepath.Join(cacheBindir, "wkhtmltopdf.exe")
	} else {
		wk = filepath.Join(cacheBindir, "wkhtmltopdf")
	}

	app := &App{
		LState: L,
		variable: map[string]interface{}{
			"GOARCH": runtime.GOARCH,
			"GOOS":   runtime.GOOS,
		},
		Cachedir:       cachedir,
		CacheBindir:    cacheBindir,
		CacheTmpdir:    cacheTmpdir,
		WkhtmltopdfCmd: wk,
		Targetpdfs:     []*TargetPdf{},
		Tmpfiles:       []string{},
	}

	L.SetGlobal("var", toLValue(L, app.variable))

	return app
}

func (app *App) Close() {
	app.LState.Close()
	for _, f := range app.Tmpfiles {
		os.Remove(f)
	}
}

func (app *App) Init() error {
	// It intends not to output timestamp with log.
	log.SetFlags(0)
	// support leveled logging.
	loglv.Init()
	// output to stdout
	loglv.SetOutput(os.Stdout)

	if app.LogLevel == "" {
		app.LogLevel = "info"
	}

	err := loglv.SetLevelByString(app.LogLevel)
	if err != nil {
		return err
	}

	// load lua libraries.
	app.openLibs()

	return nil
}

func (app *App) LoadVariableFromJSON(v string) error {
	variable := app.variable
	err := json.Unmarshal([]byte(v), &variable)
	if err != nil {
		return err
	}

	L := app.LState
	L.SetGlobal("var", toLValue(L, app.variable))

	return nil
}

func (app *App) LoadVariableFromJSONFile(jsonFile string) error {
	b, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return err
	}

	variable := app.variable
	err = json.Unmarshal(b, &variable)
	if err != nil {
		return err
	}

	L := app.LState
	L.SetGlobal("var", toLValue(L, app.variable))

	return nil
}

func (app *App) LoadRecipe(recipeContent string) error {
	if err := app.LState.DoString(recipeContent); err != nil {
		return err
	}

	return nil
}

func (app *App) LoadScriptFile(recipeFile string) error {
	if err := app.LState.DoFile(recipeFile); err != nil {
		return err
	}

	return nil
}

// see also http://stackoverflow.com/questions/5776125/wkhtmltopdf-command-fails
func (app *App) CreateTempHTMLfileByContent(content []byte) (string, error) {
	tmpFile, err := ioutil.TempFile(app.CacheTmpdir, "")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write(content)
	if err != nil {
		return "", err
	}

	err = tmpFile.Chmod(0600)
	if err != nil {
		return "", err
	}

	name := tmpFile.Name()
	name2 := name + ".html"
	if err := os.Rename(name, name2); err != nil {
		return "", err
	}

	if loglv.IsDebug() {
		log.Printf("    (Debug) Created tmpfile: %s", name2)
	}

	app.Tmpfiles = append(app.Tmpfiles, name2)

	return name2, nil
}

func (app *App) CreateTempCSSfileByContent(content []byte) (string, error) {
	tmpFile, err := ioutil.TempFile(app.CacheTmpdir, "")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write(content)
	if err != nil {
		return "", err
	}

	err = tmpFile.Chmod(0600)
	if err != nil {
		return "", err
	}

	name := tmpFile.Name()
	name2 := name + ".css"
	if err := os.Rename(name, name2); err != nil {
		return "", err
	}

	if loglv.IsDebug() {
		log.Printf("    (Debug) Created tmpfile: %s", name2)
	}

	app.Tmpfiles = append(app.Tmpfiles, name2)

	return name2, nil
}

func (app *App) Run() error {
	log.Printf("==> Starting " + Name + "...")

	if loglv.IsDebug() {
		log.Printf("    (Debug) Log level '%s'", loglv.LvString())
	}

	// create cache directory
	if _, err := os.Stat(app.Cachedir); os.IsNotExist(err) {
		defaultUmask := Umask(0)
		os.MkdirAll(app.Cachedir, 0777)
		Umask(defaultUmask)

		if loglv.IsDebug() {
			log.Printf("    (Debug) created dir = %s", app.Cachedir)
		}
	}

	if _, err := os.Stat(app.CacheTmpdir); os.IsNotExist(err) {
		defaultUmask := Umask(0)
		os.MkdirAll(app.CacheTmpdir, 0777)
		Umask(defaultUmask)

		if loglv.IsDebug() {
			log.Printf("    (Debug) created dir = %s", app.CacheTmpdir)
		}
	}

	if _, err := os.Stat(app.CacheBindir); os.IsNotExist(err) {
		defaultUmask := Umask(0)
		os.MkdirAll(app.CacheBindir, 0777)
		Umask(defaultUmask)

		if loglv.IsDebug() {
			log.Printf("    (Debug) created dir = %s", app.CacheBindir)
		}
	}

	// check wkhtmltopdf command
	if _, err := os.Stat(app.WkhtmltopdfCmd); os.IsNotExist(err) {
		assetName := "wkhtmltopdf"
		if runtime.GOOS == "windows" {
			assetName = "wkhtmltopdf.exe"
		}

		b, err := resource.Asset(assetName)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(app.WkhtmltopdfCmd, b, 0777)
		if err != nil {
			return err
		}

		if loglv.IsDebug() {
			log.Printf("    (Debug) outputed wkhtmltopdf = %s", app.WkhtmltopdfCmd)
		}
	}

	if loglv.IsDebug() {
		log.Printf("    (Debug) wkhtmltopdf command: %s", app.WkhtmltopdfCmd)
	}

	log.Printf("==> Loaded %d pdf config.", len(app.Targetpdfs))

	for _, tp := range app.Targetpdfs {
		err := tp.Run()
		if err != nil {
			return err
		}
	}

	log.Print("==> Complete!")
	return nil
}
