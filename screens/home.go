// Package home is the home screen of this app
package home

import (
	"fyne.io/fyne/cmd/fyne_demo/data"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gtk-attendance/utils/ui"
	"os"
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
	if p.confFile == "" || p.pdfFile == "" {
		ui.ShowInformation("提示","请上传考勤文件和配置文件！", p.win)
		return
	}

	p.pdfIsConverting <- true
	err := Calc(p)
	if err != nil {
		p.errorChan <- err.Error()
	}
	p.outChan <- p.pdfFile
	ui.ShowInformation("提示","ok", p.win)
	p.pdfIsConverting <- false
}

func (p *Home) Menu(){
	p.win.SetMainMenu(
		fyne.NewMainMenu(
			fyne.NewMenu("文件",
				fyne.NewMenuItem("打开", func() { p.showFileOpen() }),
			),
			fyne.NewMenu("运行",
				fyne.NewMenuItem("启动", func() { p.Start()}),
			),
			))
}

func (p *Home) UILayout() *fyne.Container {
	logo := canvas.NewImageFromResource(data.FyneScene)
	logo.SetMinSize(fyne.NewSize(228, 167))




	ltitle := widget.NewLabelWithStyle("考勤文件:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	txtFile := widget.NewEntry()
	txtFile.SetPlaceHolder("请选择考勤文件")
	txtFile.Resize(fyne.NewSize(10, 200))
	titleBtn := widget.NewButtonWithIcon("选择文件", theme.FolderOpenIcon(), p.showFileOpen)

	conf := widget.NewLabelWithStyle("假期配置:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	confFile := widget.NewEntry()
	confFile.SetPlaceHolder("请选择假期配置")
	confBtn := widget.NewButtonWithIcon("选择文件", theme.FolderOpenIcon(), p.showConfOpen)

	out := widget.NewLabelWithStyle("计算结果:", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	outFile := widget.NewEntry()
	outFile.SetPlaceHolder("计算结果显示")
	outBtn := widget.NewButtonWithIcon("计  算", theme.HomeIcon(), p.Start)

	head := container.New(layout.NewHBoxLayout())
	title := container.New(layout.NewGridLayout(3), ltitle,titleBtn, txtFile)
	confLay := container.New(layout.NewGridLayout(3), conf,confBtn,  confFile)
	finsh := container.New(layout.NewGridLayout(3), out, outBtn, outFile)



	box := container.NewVBox(
		layout.NewSpacer(),
		layout.NewSpacer(),
		container.New(layout.NewGridLayout(2),
			widget.NewButton("Dark", func() {
				os.Setenv("FYNE_THEME","dark")
				fyne.CurrentApp().Settings().UpdateTheme()
			}),
			widget.NewButton("Light", func() {
				os.Setenv("FYNE_THEME","light")
				fyne.CurrentApp().Settings().UpdateTheme()
			}),
		),
	)


	msg := widget.NewEntry()


	l := container.New(
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

