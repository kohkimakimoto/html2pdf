package main

import (
	"flag"
	"fmt"
	"github.com/kohkimakimoto/html2pdf/html2pdf"
	"github.com/kohkimakimoto/html2pdf/support/color"
	"os"
)

func main() {
	os.Exit(realMain())
}

func realMain() (status int) {
	defer func() {
		if err := recover(); err != nil {
			printError(err)
			status = 1
		}
	}()

	// parse flags...
	var optLogLevel, optVarJson, optVarJsonFile string
	var optVersion bool

	flag.StringVar(&optLogLevel, "l", "info", "")
	flag.StringVar(&optLogLevel, "log-level", "info", "")
	flag.StringVar(&optVarJson, "var", "", "")
	flag.StringVar(&optVarJsonFile, "var-file", "", "")

	flag.BoolVar(&optVersion, "v", false, "")
	flag.BoolVar(&optVersion, "version", false, "")

	flag.Usage = func() {
		fmt.Println(`Usage: ` + html2pdf.Name + ` [OPTIONS...] [SCRIPT_FILE]

  ` + html2pdf.Name + ` -- ` + html2pdf.Usage + `
  version ` + html2pdf.Version + ` (` + html2pdf.CommitHash + `)

Options:
  -l, -log-level=LEVEL       Log level (quiet|error|warning|info|debug). Default is 'info'.
  -h, -help                  Show help
  -v, -version               Print the version
  -var=JSON                  JSON string to input variables.
  -var-file=JSON_FILE        JSON file to input variables.
`)
	}
	flag.Parse()

	if optVersion {
		// show version
		fmt.Println(html2pdf.Name + " version " + html2pdf.Version + " (" + html2pdf.CommitHash + ")")
		return 0
	}

	if flag.NArg() == 0 {
		// show usage
		flag.Usage()
		return 0
	}

	// specify the script file. parse flags again for using flags after the recipe file.
	scriptFile := flag.Arg(0)
	indexOfScript := (len(os.Args) - flag.NArg())
	flag.CommandLine.Parse(os.Args[indexOfScript+1:])

	// finished parsing flags, start initializing app.
	app := html2pdf.NewApp()
	// defer app.Close()

	app.LogLevel = optLogLevel

	if err := app.Init(); err != nil {
		printError(err)
		status = 1
	}

	if optVarJsonFile != "" {
		if err := app.LoadVariableFromJSONFile(optVarJsonFile); err != nil {
			printError(err)
			return 1
		}
	}

	if optVarJson != "" {
		if err := app.LoadVariableFromJSON(optVarJson); err != nil {
			printError(err)
			return 1
		}
	}
	if err := app.LoadScriptFile(scriptFile); err != nil {
		printError(err)
		return 1
	}

	if err := app.Run(); err != nil {
		printError(err)
		return 1
	}

	return status
}

func printError(err interface{}) {
	fmt.Fprintf(os.Stderr, color.FgRB(html2pdf.Name+" aborted!\n"))
	fmt.Fprintf(os.Stderr, color.FgRB("%v\n", err))
}
