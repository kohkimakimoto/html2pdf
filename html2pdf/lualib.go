package html2pdf

import (
	"fmt"
	"github.com/cjoudrey/gluahttp"
	"github.com/kohkimakimoto/gluaenv"
	"github.com/kohkimakimoto/gluafs"
	"github.com/kohkimakimoto/gluatemplate"
	"github.com/kohkimakimoto/gluayaml"
	"github.com/kohkimakimoto/gluamarkdown"
	"github.com/kohkimakimoto/loglv"
	gluajson "github.com/layeh/gopher-json"
	"github.com/yuin/gluare"
	"github.com/yuin/gopher-lua"
	"log"
	"net/http"
)

func (app *App) openLibs() {
	L := app.LState

	loadLTargetPdfClass(L)

	L.SetGlobal("pdf", L.NewFunction(app.fnPdf))
	L.PreloadModule("html2pdf", app.luaModuleLoader)

	// buit-in packages
	L.PreloadModule("json", gluajson.Loader)
	L.PreloadModule("fs", gluafs.Loader)
	L.PreloadModule("yaml", gluayaml.Loader)
	L.PreloadModule("template", gluatemplate.Loader)
	L.PreloadModule("markdown", gluamarkdown.Loader)
	L.PreloadModule("env", gluaenv.Loader)
	L.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)
	L.PreloadModule("re", gluare.Loader)
}

func (app *App) luaModuleLoader(L *lua.LState) int {
	tb := L.NewTable()
	L.SetFuncs(tb, map[string]lua.LGFunction{
		"pdf": app.fnPdf,
	})

	L.Push(tb)

	return 1
}

func (app *App) fnPdf(L *lua.LState) int {
	name := L.CheckString(1)

	if L.GetTop() == 1 {
		// object or DSL style
		r := app.registerTargetPdf(L, name)
		L.Push(newLTargetPdf(L, r))

		return 1
	} else if L.GetTop() == 2 {
		// function style
		tb := L.CheckTable(2)
		r := app.registerTargetPdf(L, name)
		setupTargetPdf(r, tb)
		L.Push(newLTargetPdf(L, r))

		return 1
	}

	return 0
}

func (app *App) registerTargetPdf(L *lua.LState, name string) *TargetPdf {
	tp := NewTargetPdf(name, app)

	if loglv.IsDebug() {
		log.Printf("    (Debug) registering pdf '%s'", tp.Name)
	}

	// set default attributes
	app.RegisterTargetPdf(tp)

	return tp
}

func (app *App) RegisterTargetPdf(tp *TargetPdf) {
	app.Targetpdfs = append(app.Targetpdfs, tp)
}

// Lua TargetPdf Class
const lTargetPdfClass = "TargetPdf*"

func loadLTargetPdfClass(L *lua.LState) {
	mt := L.NewTypeMetatable(lTargetPdfClass)

	L.SetField(mt, "__call", L.NewFunction(targetPdfCall))
	L.SetField(mt, "__index", L.NewFunction(targetPdfIndex))
	L.SetField(mt, "__newindex", L.NewFunction(targetPdfNewindex))
}

func updateTargetPdf(tp *TargetPdf, key string, value lua.LValue) {
	tp.LValues[key] = value
}

func setupTargetPdf(r *TargetPdf, attributes *lua.LTable) {
	attributes.ForEach(func(k, v lua.LValue) {
		if kstr, ok := toString(k); ok {
			updateTargetPdf(r, kstr, v)
		} else {
			panic(fmt.Sprintf("'%s' An key must be string", r.Name))
		}
	})
}

func newLTargetPdf(L *lua.LState, r *TargetPdf) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = r
	L.SetMetatable(ud, L.GetTypeMetatable(lTargetPdfClass))

	return ud
}

func checkTargetPdf(L *lua.LState) *TargetPdf {
	ud := L.CheckUserData(1)
	if result, ok := ud.Value.(*TargetPdf); ok {
		return result
	}
	L.ArgError(1, "TargetPdf expected")

	return nil
}

func targetPdfCall(L *lua.LState) int {
	r := checkTargetPdf(L)
	tb := L.CheckTable(2)

	setupTargetPdf(r, tb)

	return 0
}

func targetPdfIndex(L *lua.LState) int {
	tp := checkTargetPdf(L)
	index := L.CheckString(2)

	v, ok := tp.LValues[index]
	if v == nil || !ok {
		v = lua.LNil
	}

	L.Push(v)
	return 1
}

func targetPdfNewindex(L *lua.LState) int {
	tp := checkTargetPdf(L)
	index := L.CheckString(2)
	value := L.CheckAny(3)

	updateTargetPdf(tp, index, value)

	return 0
}
