package html2pdf

import (
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/kohkimakimoto/html2pdf/support/color"
	"github.com/kohkimakimoto/html2pdf/support/gluamapper"
	"github.com/kohkimakimoto/loglv"
	"github.com/yuin/gopher-lua"
	"log"
	"strconv"
)

type TargetPdf struct {
	Name    string
	LValues map[string]lua.LValue
	App     *App
	Options *PdfOptions
}

func NewTargetPdf(name string, app *App) *TargetPdf {
	return &TargetPdf{
		Name:    name,
		LValues: map[string]lua.LValue{},
		App:     app,
		Options: &PdfOptions{},
	}
}

func (tp *TargetPdf) Run() error {
	log.Print(color.FgBold(fmt.Sprintf("==> Processing: %s", tp.Name)))
	log.Print(fmt.Sprintf("    output_file: %s", tp.OutputFile()))

	wkhtmltopdf.SetPath(tp.App.WkhtmltopdfCmd)
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}

	if options, ok := tp.LValues["options"]; ok {
		if opttb, ok := options.(*lua.LTable); ok {
			if err := gluamapper.Map(opttb, tp.Options); err != nil {
				return err
			}
		}
	}

	// setup gloabal options
	if tp.Options.CookieJar != "" {
		pdfg.CookieJar.Set(tp.Options.CookieJar)
	}
	if tp.Options.Copies != "" {
		pdfg.Copies.Set(parseUint(tp.Options.Copies))
	}
	if tp.Options.Dpi != "" {
		pdfg.Dpi.Set(parseUint(tp.Options.Dpi))
	}
	if tp.Options.Grayscale {
		pdfg.Grayscale.Set(tp.Options.Grayscale)
	}
	if tp.Options.ImageDpi != "" {
		pdfg.ImageDpi.Set(parseUint(tp.Options.ImageDpi))
	}
	if tp.Options.ImageQuality != "" {
		pdfg.ImageQuality.Set(parseUint(tp.Options.ImageQuality))
	}
	if tp.Options.Lowquality {
		pdfg.Lowquality.Set(tp.Options.Lowquality)
	}
	if tp.Options.MarginBottom != "" {
		pdfg.MarginBottom.Set(parseUint(tp.Options.MarginBottom))
	}
	if tp.Options.MarginLeft != "" {
		pdfg.MarginLeft.Set(parseUint(tp.Options.MarginLeft))
	}
	if tp.Options.MarginRight != "" {
		pdfg.MarginRight.Set(parseUint(tp.Options.MarginRight))
	}
	if tp.Options.MarginTop != "" {
		pdfg.MarginTop.Set(parseUint(tp.Options.MarginTop))
	}
	if tp.Options.Orientation != "" {
		pdfg.Orientation.Set(tp.Options.Orientation)
	}
	if tp.Options.NoCollate {
		pdfg.NoCollate.Set(tp.Options.NoCollate)
	}
	if tp.Options.PageHeight != "" {
		pdfg.MarginLeft.Set(parseUint(tp.Options.PageHeight))
	}
	if tp.Options.PageSize != "" {
		pdfg.PageSize.Set(tp.Options.PageSize)
	}
	if tp.Options.PageWidth != "" {
		pdfg.PageWidth.Set(parseUint(tp.Options.PageWidth))
	}
	if tp.Options.NoPdfCompression {
		pdfg.NoPdfCompression.Set(tp.Options.NoPdfCompression)
	}
	if tp.Options.Title != "" {
		pdfg.Title.Set(tp.Options.Title)
	}

	// add cover
	cover, err := tp.Cover()
	if err != nil {
		return err
	}
	if cover != nil {
		pdfg.Cover.Input = cover.InputFile()
	}

	// add pages
	pages, err := tp.Pages()
	if err != nil {
		return err
	}
	if pages != nil && len(pages) > 0 {
		for _, p := range pages {
			page := wkhtmltopdf.NewPage(p.InputFile())
			pdfg.AddPage(page)
		}
	}

	if loglv.IsDebug() {
		log.Printf("    (Debug) wkhtmltopdf args: %s", pdfg.Args())
	}

	err = pdfg.Create()
	if err != nil {
		return err
	}

	err = pdfg.WriteFile(tp.OutputFile())
	if err != nil {
		return err
	}

	return nil
}

func (tp *TargetPdf) OutputFile() string {
	if dist, ok := toString(tp.LValues["output_file"]); ok {
		return dist
	}

	return tp.Name
}

func (tp *TargetPdf) Pages() ([]*Page, error) {
	ret := []*Page{}

	pages, ok := tp.LValues["pages"]
	if !ok {
		return ret, nil
	}
	pagesTb, ok := pages.(*lua.LTable)
	if !ok {
		return ret, nil
	}

	maxn := pagesTb.MaxN()
	if maxn == 0 { // table
		p := &Page{}
		p.targetPdf = tp
		if err := gluamapper.Map(pagesTb, p); err != nil {
			return nil, err
		}

		ret = append(ret, p)
	} else {
		// array
		pagesTb.ForEach(func(k, v lua.LValue) {
			if lp, ok := v.(*lua.LTable); ok {
				p := &Page{}
				p.targetPdf = tp

				if err := gluamapper.Map(lp, p); err != nil {
					panic(err)
				}

				ret = append(ret, p)
			} else {
				panic(fmt.Sprintf("'%s' invalid data format: pages can't support nested array table.", tp.Name))
			}
		})
	}

	return ret, nil
}

func (tp *TargetPdf) Cover() (*Cover, error) {
	cover, ok := tp.LValues["cover"]
	if !ok {
		return nil, nil
	}

	coverTb, ok := cover.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("'%s' invalid data format: cover only support table.", tp.Name)
	}

	ret := &Cover{}
	ret.targetPdf = tp

	maxn := coverTb.MaxN()
	if maxn == 0 { // table
		if err := gluamapper.Map(coverTb, ret); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("'%s' invalid data format: cover can't support array table.", tp.Name)
	}

	return ret, nil
}

func (tp *TargetPdf) convertPage(src map[string]interface{}) (*Page, error) {
	page := &Page{}
	page.targetPdf = tp

	return page, nil
}

func (tp *TargetPdf) SendContentToTempHTMLfile(content []byte) (string, error) {
	return tp.App.SendContentToTempHTMLfile(content)
}

type Page struct {
	targetPdf    *TargetPdf
	Input        string
	InputContent string
}

func (p *Page) InputFile() string {
	var inputfile string

	tp := p.targetPdf

	if p.InputContent != "" {
		t, err := tp.SendContentToTempHTMLfile([]byte(p.InputContent))
		if err != nil {
			panic(err)
		}
		inputfile = t
	} else if p.Input != "" {
		inputfile = p.Input
	} else {
		panic(fmt.Sprintf("'%s': page must have 'input' or 'input_content'.", tp.Name))
	}

	return inputfile
}

type Cover struct {
	targetPdf    *TargetPdf
	Input        string
	InputContent string
}

func (p *Cover) InputFile() string {
	var inputfile string

	tp := p.targetPdf

	if p.InputContent != "" {
		t, err := tp.SendContentToTempHTMLfile([]byte(p.InputContent))
		if err != nil {
			panic(err)
		}
		inputfile = t
	} else if p.Input != "" {
		inputfile = p.Input
	} else {
		panic(fmt.Sprintf("'%s': page must have 'input' or 'input_content'.", tp.Name))
	}

	return inputfile
}

type PdfOptions struct {
	CookieJar         string // Read and write cookies from and to the supplied cookie jar file
	Copies            string // (actually uint) Number of copies to print into the pdf file (default 1)
	Dpi               string // (actually uint) Change the dpi explicitly (this has no effect on X11 based systems)
//	ExtendedHelp      bool   // Display more extensive help, detailing less common command switches
	Grayscale         bool   // PDF will be generated in grayscale
//	Help              bool   // Display help
//	HTMLDoc           bool   // Output program html help
	ImageDpi          string   // (actually uint) When embedding images scale them down to this dpi (default 600)
	ImageQuality      string   // (actually uint) When jpeg compressing images use this quality (default 94)
//	License           bool   // Output license information and exit
	Lowquality        bool   // Generates lower quality pdf/ps. Useful to shrink the result document space
//	ManPage           bool   // Output program man page
	MarginBottom      string   // (actually uint) Set the page bottom margin
	MarginLeft        string   // (actually uint) Set the page left margin (default 10mm)
	MarginRight       string   // (actually uint) Set the page right margin (default 10mm)
	MarginTop         string   // (actually uint) Set the page top margin
	Orientation       string // Set orientation to Landscape or Portrait (default Portrait)
	NoCollate         bool   // Do not collate when printing multiple copies (default collate)
	PageHeight        string   // (actually uint) Page height
	PageSize          string // Set paper size to: A4, Letter, etc. (default A4)
	PageWidth         string   // (actually uint) Page width
	NoPdfCompression  bool   // Do not use lossless compression on pdf objects
//	Quiet             bool   // Be less verbose
//	ReadArgsFromStdin bool   // Read command line arguments from stdin
//	Readme            bool   // Output program readme
	Title             string // The title of the generated pdf file (The title of the first document is used if not specified)
//	Version           bool   // Output version information and exit
}

func parseUint(str string) uint {
	v, err := strconv.ParseUint(str, 10, 0)
	if err != nil {
		panic(fmt.Sprintf("detected invalid parameter (uint expected): %v", err))
	}

	return uint(v)
}
