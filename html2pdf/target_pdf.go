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
	Name          string
	LValues       map[string]lua.LValue
	App           *App
}

func NewTargetPdf(name string, app *App) *TargetPdf {
	return &TargetPdf{
		Name:          name,
		LValues:       map[string]lua.LValue{},
		App:           app,
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

	// parse global options
	globaOptions := &GlobalOptions{}
	if options, ok := tp.LValues["options"]; ok {
		if opttb, ok := options.(*lua.LTable); ok {
			if err := gluamapper.Map(opttb, globaOptions); err != nil {
				return err
			}
		}
	}

	// gloabal options
	if globaOptions.CookieJar != "" {
		pdfg.CookieJar.Set(globaOptions.CookieJar)
	}
	if globaOptions.Copies != "" {
		pdfg.Copies.Set(parseUint(globaOptions.Copies))
	}
	if globaOptions.Dpi != "" {
		pdfg.Dpi.Set(parseUint(globaOptions.Dpi))
	}
	if globaOptions.Grayscale {
		pdfg.Grayscale.Set(globaOptions.Grayscale)
	}
	if globaOptions.ImageDpi != "" {
		pdfg.ImageDpi.Set(parseUint(globaOptions.ImageDpi))
	}
	if globaOptions.ImageQuality != "" {
		pdfg.ImageQuality.Set(parseUint(globaOptions.ImageQuality))
	}
	if globaOptions.Lowquality {
		pdfg.Lowquality.Set(globaOptions.Lowquality)
	}
	if globaOptions.MarginBottom != "" {
		pdfg.MarginBottom.Set(parseUint(globaOptions.MarginBottom))
	}
	if globaOptions.MarginLeft != "" {
		pdfg.MarginLeft.Set(parseUint(globaOptions.MarginLeft))
	}
	if globaOptions.MarginRight != "" {
		pdfg.MarginRight.Set(parseUint(globaOptions.MarginRight))
	}
	if globaOptions.MarginTop != "" {
		pdfg.MarginTop.Set(parseUint(globaOptions.MarginTop))
	}
	if globaOptions.Orientation != "" {
		pdfg.Orientation.Set(globaOptions.Orientation)
	}
	if globaOptions.NoCollate {
		pdfg.NoCollate.Set(globaOptions.NoCollate)
	}
	if globaOptions.PageHeight != "" {
		pdfg.MarginLeft.Set(parseUint(globaOptions.PageHeight))
	}
	if globaOptions.PageSize != "" {
		pdfg.PageSize.Set(globaOptions.PageSize)
	}
	if globaOptions.PageWidth != "" {
		pdfg.PageWidth.Set(parseUint(globaOptions.PageWidth))
	}
	if globaOptions.NoPdfCompression {
		pdfg.NoPdfCompression.Set(globaOptions.NoPdfCompression)
	}
	if globaOptions.Title != "" {
		pdfg.Title.Set(globaOptions.Title)
	}

	// outline options
	if globaOptions.NoOutline {
		pdfg.NoOutline.Set(globaOptions.NoOutline)
	}
	if globaOptions.OutlineDepth != "" {
		pdfg.OutlineDepth.Set(parseUint(globaOptions.OutlineDepth))
	}
	
	// add cover
	cover, err := tp.Cover()
	if err != nil {
		return err
	}
	if cover != nil {
		pdfg.Cover.Input = cover.InputFile()

		if cover.Encoding != "" {
			pdfg.Cover.Encoding.Set(cover.Encoding)
		}
		if cover.PageOffset != "" {
			pdfg.Cover.PageOffset.Set(parseUint(cover.PageOffset))
		}
		if style := cover.UserStyleSheetFile(); style != "" {
			pdfg.Cover.UserStyleSheet.Set(style)
		}
	}

	// add pages
	pages, err := tp.Pages()
	if err != nil {
		return err
	}
	if pages != nil && len(pages) > 0 {
		for _, p := range pages {
			page := wkhtmltopdf.NewPage(p.InputFile())

			if p.Encoding != "" {
				page.Encoding.Set(p.Encoding)
			}
			if p.PageOffset != "" {
				page.PageOffset.Set(parseUint(p.PageOffset))
			}
			if style := p.UserStyleSheetFile(); style != "" {
				page.UserStyleSheet.Set(style)
			}

			pdfg.AddPage(page)
		}
	}

	// add TOC
	toc, err := tp.TOC()
	if err != nil {
		return err
	}

	if cover != nil {
		pdfg.TOC.Include = true

		if toc.DisableDottedLines {
			pdfg.TOC.DisableDottedLines.Set(toc.DisableDottedLines)
		}
		if toc.DisableTocLinks {
			pdfg.TOC.DisableTocLinks.Set(toc.DisableTocLinks)
		}
		if toc.TocHeaderText != "" {
			pdfg.TOC.TocHeaderText.Set(toc.TocHeaderText)
		}
		if toc.TocLevelIndentation != "" {
			pdfg.TOC.TocLevelIndentation.Set(parseUint(toc.TocLevelIndentation))
		}
		if toc.Encoding != "" {
			pdfg.TOC.Encoding.Set(toc.Encoding)
		}
		if toc.PageOffset != "" {
			pdfg.TOC.PageOffset.Set(parseUint(toc.PageOffset))
		}
		if style := toc.UserStyleSheetFile(); style != "" {
			pdfg.TOC.UserStyleSheet.Set(style)
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

func (tp *TargetPdf) TOC() (*TOC, error) {
	toc, ok := tp.LValues["toc"]
	if !ok {
		return nil, nil
	}

	tocTb, ok := toc.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("'%s' invalid data format: toc only support table.", tp.Name)
	}

	ret := &TOC{}
	ret.targetPdf = tp

	maxn := tocTb.MaxN()
	if maxn == 0 { // table
		if err := gluamapper.Map(tocTb, ret); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("'%s' invalid data format: toc can't support array table.", tp.Name)
	}

	return ret, nil
}

func (tp *TargetPdf) CreateTempHTMLfileByContent(content []byte) (string, error) {
	return tp.App.CreateTempHTMLfileByContent(content)
}

func (tp *TargetPdf) CreateTempCSSfileByContent(content []byte) (string, error) {
	return tp.App.CreateTempCSSfileByContent(content)
}


type Cover struct {
	targetPdf    *TargetPdf
	Input        string
	InputContent string

	// page options
	Encoding       string //Set the default text encoding, for input
	UserStyleSheet string //Specify a user style sheet, to load with every page
	UserStyleSheetContent string //Specify a user style sheet, to load with every page
	PageOffset     string // (actually uint)Set the starting page number (default 0)
}

func (p *Cover) InputFile() string {
	var inputfile string

	tp := p.targetPdf

	if p.InputContent != "" {
		t, err := tp.CreateTempHTMLfileByContent([]byte(p.InputContent))
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

func (p *Cover) UserStyleSheetFile() string {
	var inputfile string

	tp := p.targetPdf

	if p.UserStyleSheetContent != "" {
		t, err := tp.CreateTempCSSfileByContent([]byte(p.UserStyleSheetContent))
		if err != nil {
			panic(err)
		}
		inputfile = t
	} else if p.UserStyleSheet != "" {
		inputfile = p.UserStyleSheet
	} else {
		return ""
	}

	return inputfile
}

type Page struct {
	targetPdf    *TargetPdf
	Input        string
	InputContent string

	// page options
	Encoding       string //Set the default text encoding, for input
	UserStyleSheet string //Specify a user style sheet, to load with every page
	UserStyleSheetContent string //Specify a user style sheet, to load with every page
	PageOffset     string // (actually uint)Set the starting page number (default 0)
}

func (p *Page) InputFile() string {
	var inputfile string

	tp := p.targetPdf

	if p.InputContent != "" {
		t, err := tp.CreateTempHTMLfileByContent([]byte(p.InputContent))
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

func (p *Page) UserStyleSheetFile() string {
	var inputfile string

	tp := p.targetPdf

	if p.UserStyleSheetContent != "" {
		t, err := tp.CreateTempCSSfileByContent([]byte(p.UserStyleSheetContent))
		if err != nil {
			panic(err)
		}
		inputfile = t
	} else if p.UserStyleSheet != "" {
		inputfile = p.UserStyleSheet
	} else {
		return ""
	}

	return inputfile
}

type TOC struct {
	targetPdf    *TargetPdf
	DisableDottedLines  bool   //Do not use dotted lines in the toc
	TocHeaderText       string //The header text of the toc (default Table of Contents)
	TocLevelIndentation string // (actually uint) For each level of headings in the toc indent by this length (default 1em)
	DisableTocLinks     bool   //Do not link from toc to sections

	// TODO: suppot following options
	// TocTextSizeShrink   floatOption  //For each level of headings in the toc the font is scaled by this factor
	// XslStyleSheet       string //Use the supplied xsl style sheet for printing the table of content

	// page options
	Encoding       string //Set the default text encoding, for input
	UserStyleSheet string //Specify a user style sheet, to load with every page
	UserStyleSheetContent string //Specify a user style sheet, to load with every page
	PageOffset     string // (actually uint)Set the starting page number (default 0)
}

func (p *TOC) UserStyleSheetFile() string {
	var inputfile string

	tp := p.targetPdf

	if p.UserStyleSheetContent != "" {
		t, err := tp.CreateTempCSSfileByContent([]byte(p.UserStyleSheetContent))
		if err != nil {
			panic(err)
		}
		inputfile = t
	} else if p.UserStyleSheet != "" {
		inputfile = p.UserStyleSheet
	} else {
		return ""
	}

	return inputfile
}

// see also
//   http://wkhtmltopdf.org/usage/wkhtmltopdf.txt
//   https://github.com/SebastiaanKlippert/go-wkhtmltopdf/blob/master/options.go

type GlobalOptions struct {
	CookieJar string // Read and write cookies from and to the supplied cookie jar file
	Copies    string // (actually uint) Number of copies to print into the pdf file (default 1)
	Dpi       string // (actually uint) Change the dpi explicitly (this has no effect on X11 based systems)
	//	ExtendedHelp      bool   // Display more extensive help, detailing less common command switches
	Grayscale bool // PDF will be generated in grayscale
	//	Help              bool   // Display help
	//	HTMLDoc           bool   // Output program html help
	ImageDpi     string // (actually uint) When embedding images scale them down to this dpi (default 600)
	ImageQuality string // (actually uint) When jpeg compressing images use this quality (default 94)
	//	License           bool   // Output license information and exit
	Lowquality bool // Generates lower quality pdf/ps. Useful to shrink the result document space
	//	ManPage           bool   // Output program man page
	MarginBottom     string // (actually uint) Set the page bottom margin
	MarginLeft       string // (actually uint) Set the page left margin (default 10mm)
	MarginRight      string // (actually uint) Set the page right margin (default 10mm)
	MarginTop        string // (actually uint) Set the page top margin
	Orientation      string // Set orientation to Landscape or Portrait (default Portrait)
	NoCollate        bool   // Do not collate when printing multiple copies (default collate)
	PageHeight       string // (actually uint) Page height
	PageSize         string // Set paper size to: A4, Letter, etc. (default A4)
	PageWidth        string // (actually uint) Page width
	NoPdfCompression bool   // Do not use lossless compression on pdf objects
	//	Quiet             bool   // Be less verbose
	//	ReadArgsFromStdin bool   // Read command line arguments from stdin
	//	Readme            bool   // Output program readme
	Title string // The title of the generated pdf file (The title of the first document is used if not specified)
	//	Version           bool   // Output version information and exit

	// outlineOptions

	NoOutline    bool   //Do not put an outline into the pdf
	OutlineDepth string // (actually uint) Set the depth of the outline (default 4)
}

func parseUint(str string) uint {
	v, err := strconv.ParseUint(str, 10, 0)
	if err != nil {
		panic(fmt.Sprintf("detected invalid parameter (uint expected): %v", err))
	}

	return uint(v)
}
