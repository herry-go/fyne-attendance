// Package main provides various examples of Fyne API capabilities
package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
	home "gtk-attendance/screens"
)

func main() {
	a := app.NewWithID("v1.0.0")
	a.SetIcon(theme.FyneLogo())
	w := a.NewWindow("HM TOOL")

	h := home.NewHome(w)
	h.Menu()
	w.SetContent(h.UILayout())

	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(true)
	w.CenterOnScreen()
	w.SetMaster()
	w.ShowAndRun()

}

