// Package main provides various examples of Fyne API capabilities
package main

import (
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"gtk-attendance/fonts"
	home "gtk-attendance/screens"
)

func main() {
	a := app.NewWithID("v1.0.0")
	a.SetIcon(theme.FyneLogo())
	a.Settings().SetTheme(&fonts.MyTheme{})
	w := a.NewWindow("考勤助手")

	h := home.NewHome(w)
	h.Menu()
	w.SetContent(h.UILayout())

	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(false)
	w.CenterOnScreen()
	w.SetMaster()
	w.ShowAndRun()

}

