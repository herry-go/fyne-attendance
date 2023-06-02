// Package home is the home screen of this app
package home

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/cmd/fyne_demo/data"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"gtk-attendance/utils/ui"
	"path/filepath"
	"strings"
)

// Home page
type Home struct {
	win                    fyne.Window
	pdfFile, pdfFolder     string
	pdfFileNameChanged     chan string
	pageRange, imageFormat string
	imageScale             int
	pdfIsConverting        chan bool
	errorChan              chan string
	outChan                chan string

	confFile, confFolder     string
	confFileNameChanged     chan string
}

// NewHome is used to new it
func NewHome(win fyne.Window) Home {
	p := fyne.CurrentApp().Preferences()
	imgf := p.StringWithFallback("imageFormat", "png")
	imgs := p.IntWithFallback("imageScale", 200)
	return Home{
		pdfFileNameChanged: make(chan string),
		errorChan: make(chan string),
		outChan: make(chan string),
		confFileNameChanged: make(chan string),
		pdfIsConverting:    make(chan bool),
		imageFormat:        imgf,
		imageScale:         imgs,
		win: 				win,
	}
}

func (p *Home) fileInputChanged(s string) {
	ltrChar := "\u202a"
	s = strings.TrimLeft(s, ltrChar)
	if len(s) > 0 && strings.HasSuffix(s, ".xlsx") {
		if dir, file := filepath.Split(s); dir != "" {
			p.pdfFolder = dir
			p.pdfFile = s
			p.pdfFileNameChanged <- file
		} else if p.pdfFolder != "" {
			p.pdfFile = filepath.Join(p.pdfFolder, s)
		}
		p.win.SetTitle(p.pdfFile)
	}
}

func (p *Home) showFileOpen() {
	ui.ShowFileOpen(p.win, func(fname, fpath string, err error) {
		if err != nil {
			ui.ShowError(err, p.win)
		} else if fname != "" {
			p.pdfFileNameChanged <- fname
			p.pdfFile = fpath
			p.pdfFolder = filepath.Dir(p.pdfFile)
			p.win.SetTitle(p.pdfFile)
			FileName = fpath
		}
	}, []string{".xlsx"})
}

func (p *Home) showConfOpen() {
	ui.ShowFileOpen(p.win, func(fname, fpath string, err error) {
		if err != nil {
			ui.ShowError(err, p.win)
		} else if fname != "" {
			p.confFileNameChanged <- fname
			p.confFile = fpath
			p.confFolder = filepath.Dir(p.confFile)
			ConfName = fpath
		}
	}, []string{".json"})
}

func (p *Home) Start() {
	p.pdfIsConverting <- true
	err := Calc(p)
	if err != nil {
		p.errorChan <- err.Error()
	}
	p.outChan <- p.pdfFile
	ui.ShowInformation("prompt","ok", p.win)
	p.pdfIsConverting <- false
}

func (p *Home) Menu(){
	p.win.SetMainMenu(
		fyne.NewMainMenu(
			fyne.NewMenu("File",
				fyne.NewMenuItem("Open", func() { p.showFileOpen() }),
			),
			fyne.NewMenu("Run",
				fyne.NewMenuItem("Start", func() { p.Start()}),
			),
			))
}

func (p *Home) UILayout() *fyne.Container {
	logo := canvas.NewImageFromResource(data.FyneScene)
	logo.SetMinSize(fyne.NewSize(228, 167))




	ltitle := widget.NewLabelWithStyle("file:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	txtFile := widget.NewEntry()
	txtFile.SetPlaceHolder("")
	titleBtn := widget.NewButtonWithIcon("browse", theme.FolderOpenIcon(), p.showFileOpen)

	conf := widget.NewLabelWithStyle("conf:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	confFile := widget.NewEntry()
	confFile.SetPlaceHolder("")
	confBtn := widget.NewButtonWithIcon("browse", theme.FolderOpenIcon(), p.showConfOpen)

	out := widget.NewLabelWithStyle("out:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	outFile := widget.NewEntry()
	outFile.SetPlaceHolder("")

	head := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	title := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), ltitle,titleBtn, txtFile)
	confLay := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), conf,confBtn,  confFile)
	finsh := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), out, outFile)



	box := widget.NewVBox(
		layout.NewSpacer(),
		layout.NewSpacer(),
		widget.NewGroup("Theme",
			fyne.NewContainerWithLayout(layout.NewGridLayout(2),
				widget.NewButton("Dark", func() {
					fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
				}),
				widget.NewButton("Light", func() {
					fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
				}),
			),
		),
	)


	msg := widget.NewEntry()


	l := fyne.NewContainerWithLayout(
		layout.NewVBoxLayout(),
		head,
		title,
		confLay,
		finsh,
		msg,
		layout.NewSpacer(),
		box,
	)

	// 输入显示
	go func() {
		for {
			txtFile.SetText(<-p.pdfFileNameChanged)
		}
	}()

	// 配置显示
	go func() {
		for {
			confFile.SetText(<-p.confFileNameChanged)
		}
	}()

	//输出显示
	go func() {
		for {
			outFile.SetText(<-p.outChan)
		}
	}()


	// 消息显示
	go func() {
		for {
			msg.SetText(<-p.errorChan)
		}
	}()
	go func() {
		for b := range p.pdfIsConverting {
			if b {
				msg.SetText("runing......")
				msg.Disable()
			} else {
				msg.SetText("")
				msg.Enable()
			}
		}
	}()

	return l
}

func (p *Home) Folder() string{
	return p.pdfFolder
}

func (p *Home) SetFile(path string){
	p.pdfFile = path
}

