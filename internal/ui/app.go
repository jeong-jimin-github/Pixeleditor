package ui

import (
	"fmt"
	"image/color"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/jeong-jimin-github/Pixeleditor/internal/doc"
)

type App struct {
	fy       fyne.App
	win      fyne.Window
	doc      *doc.Document
	view     *PixelView
	scroll   *container.Scroll
	colorBox *canvas.Rectangle
	zoomLbl  *widget.Label
	status   *widget.Label
	palBox   *fyne.Container
	penBtn   *widget.Button
	eraserBtn *widget.Button
	fillBtn  *widget.Button
	pipetteBtn *widget.Button
}

func Run() {
	a := app.NewWithID("com.jeongjimin.pixeleditor")
	w := a.NewWindow("Pixel Editor Pro - Palette & Save")
	w.Resize(fyne.NewSize(1200, 850))

	ui := &App{
		fy:  a,
		win: w,
		doc: doc.New(doc.DefaultWidth, doc.DefaultHeight),
	}
	ui.build()
	ui.bindShortcuts()
	w.SetContent(ui.layout())
	ui.rebuildPalette()
	ui.refreshChrome()
	w.Canvas().Focus(ui.view)
	w.ShowAndRun()
}

func (a *App) build() {
	a.view = newPixelView(a.doc)
	a.view.win = a.win
	a.view.onChange = a.refreshChrome
	a.view.onHover = func(x, y int) { a.updateStatus(x, y) }

	a.colorBox = canvas.NewRectangle(a.doc.Brush)
	a.colorBox.SetMinSize(fyne.NewSize(64, 32))
	a.colorBox.CornerRadius = 4

	a.zoomLbl = widget.NewLabel("10x")
	a.status = widget.NewLabel("64x64 | pen")
	a.palBox = container.NewGridWithColumns(4)

	a.penBtn = widget.NewButton("Pen (P)", func() { a.setTool(doc.ToolPen) })
	a.eraserBtn = widget.NewButton("Eraser (E)", func() { a.setTool(doc.ToolEraser) })
	a.fillBtn = widget.NewButton("Fill (F)", func() { a.setTool(doc.ToolFill) })
	a.pipetteBtn = widget.NewButton("Pipette (I)", func() { a.setTool(doc.ToolPipette) })
	a.syncToolButtons()
}

func (a *App) layout() fyne.CanvasObject {
	symCheck := widget.NewCheck("Symmetry X", func(v bool) {
		a.doc.Symmetry = v
		a.view.Refresh()
	})
	lockCheck := widget.NewCheck("Lock Palette", func(v bool) {
		a.doc.LockPalette = v
	})
	gridCheck := widget.NewCheck("Grid", func(v bool) {
		a.doc.ShowGrid = v
		a.view.Refresh()
	})
	gridCheck.SetChecked(true)

	gapEntry := widget.NewEntry()
	gapEntry.SetText("1")
	gapEntry.OnChanged = func(s string) {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || n < 1 {
			return
		}
		if n > 128 {
			n = 128
		}
		a.doc.GridGap = n
		a.view.Refresh()
	}

	brushSlider := widget.NewSlider(1, 8)
	brushSlider.Step = 1
	brushSlider.Value = 1
	brushLbl := widget.NewLabel("1")
	brushSlider.OnChanged = func(v float64) {
		a.doc.BrushSize = int(v)
		brushLbl.SetText(strconv.Itoa(int(v)))
		a.view.Refresh()
	}

	colorTap := widget.NewButton("Pick color", a.chooseColor)

	refreshBtn := widget.NewButton("Refresh", a.rebuildPalette)

	sidebarTop := container.NewVBox(
		widget.NewLabelWithStyle("Tools", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		a.penBtn, a.eraserBtn, a.fillBtn, a.pipetteBtn,
		symCheck,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Color", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		a.colorBox,
		colorTap,
		container.NewBorder(nil, nil, widget.NewLabel("Brush"), brushLbl, brushSlider),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Palette", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		refreshBtn,
		lockCheck,
	)
	palScroll := container.NewVScroll(a.palBox)
	sidebar := container.NewBorder(sidebarTop, nil, nil, nil, palScroll)
	left := newFixedWidth(168, sidebar)

	openBtn := widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), a.openImage)
	saveBtn := widget.NewButtonWithIcon("Save (Ctrl+S)", theme.DocumentSaveIcon(), a.quickSave)
	saveAsBtn := widget.NewButton("Save As...", a.saveImageAs)
	newBtn := widget.NewButtonWithIcon("New", theme.DocumentIcon(), a.newImage)
	undoBtn := widget.NewButtonWithIcon("Undo", theme.ContentUndoIcon(), a.undo)
	redoBtn := widget.NewButtonWithIcon("Redo", theme.ContentRedoIcon(), a.redo)
	clearBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), a.clear)

	zoomOut := widget.NewButtonWithIcon("", theme.ZoomOutIcon(), func() { a.changeZoom(-1) })
	zoomIn := widget.NewButtonWithIcon("", theme.ZoomInIcon(), func() { a.changeZoom(1) })

	toolbar := container.NewHBox(
		newBtn, openBtn, saveBtn, saveAsBtn,
		widget.NewSeparator(),
		undoBtn, redoBtn, clearBtn,
		widget.NewSeparator(),
		gridCheck, widget.NewLabel("Gap"), gapEntry,
		widget.NewSeparator(),
		zoomOut, a.zoomLbl, zoomIn,
	)

	a.scroll = container.NewScroll(a.view)
	a.view.scroll = a.scroll
	canvasArea := container.NewStack(
		canvas.NewRectangle(color.RGBA{0x30, 0x30, 0x30, 255}),
		a.scroll,
	)

	right := container.NewBorder(toolbar, a.status, nil, nil, canvasArea)
	return container.NewBorder(nil, nil, left, nil, right)
}

func (a *App) bindShortcuts() {
	c := a.win.Canvas()
	mod := fyne.KeyModifierShortcutDefault
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: mod}, func(fyne.Shortcut) { a.quickSave() })
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: mod}, func(fyne.Shortcut) { a.undo() })
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyY, Modifier: mod}, func(fyne.Shortcut) { a.redo() })
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: mod}, func(fyne.Shortcut) { a.newImage() })
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: mod}, func(fyne.Shortcut) { a.openImage() })

	c.SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if _, ok := c.Focused().(*widget.Entry); ok {
			return
		}
		switch ev.Name {
		case fyne.KeyP:
			a.setTool(doc.ToolPen)
		case fyne.KeyE:
			a.setTool(doc.ToolEraser)
		case fyne.KeyF:
			a.setTool(doc.ToolFill)
		case fyne.KeyI:
			a.setTool(doc.ToolPipette)
		case fyne.KeyG:
			a.doc.ShowGrid = !a.doc.ShowGrid
			a.view.Refresh()
		case fyne.KeyPlus, fyne.KeyEqual:
			a.changeZoom(1)
		case fyne.KeyMinus:
			a.changeZoom(-1)
		}
	})
}

func (a *App) setTool(t doc.Tool) {
	a.doc.Tool = t
	a.syncToolButtons()
	a.updateStatus(a.view.hoverX, a.view.hoverY)
}

func (a *App) syncToolButtons() {
	set := func(b *widget.Button, on bool) {
		if on {
			b.Importance = widget.HighImportance
		} else {
			b.Importance = widget.MediumImportance
		}
		b.Refresh()
	}
	set(a.penBtn, a.doc.Tool == doc.ToolPen)
	set(a.eraserBtn, a.doc.Tool == doc.ToolEraser)
	set(a.fillBtn, a.doc.Tool == doc.ToolFill)
	set(a.pipetteBtn, a.doc.Tool == doc.ToolPipette)
}

func (a *App) changeZoom(delta int) {
	a.doc.ChangeZoom(delta)
	a.view.Refresh()
	a.scroll.Refresh()
	a.refreshChrome()
}

func (a *App) refreshChrome() {
	a.colorBox.FillColor = a.doc.Brush
	a.colorBox.Refresh()
	a.zoomLbl.SetText(fmt.Sprintf("%dx", a.doc.Zoom))
	a.syncToolButtons()
	a.updateStatus(a.view.hoverX, a.view.hoverY)
	a.setTitle()
}

func (a *App) setTitle() {
	if a.doc.FilePath != "" {
		a.win.SetTitle("Pixel Editor Pro - " + a.doc.FilePath)
		return
	}
	a.win.SetTitle("Pixel Editor Pro - Palette & Save")
}

func (a *App) updateStatus(x, y int) {
	coord := ""
	if a.doc.In(x, y) {
		c := a.doc.At(x, y)
		coord = fmt.Sprintf("%d, %d  rgba(%d,%d,%d,%d)  |  ", x, y, c.R, c.G, c.B, c.A)
	}
	a.status.SetText(fmt.Sprintf("%s%dx%d  |  %s  |  zoom %dx  |  space/middle-drag pan, wheel zoom",
		coord, a.doc.Width, a.doc.Height, a.doc.Tool, a.doc.Zoom))
}

func (a *App) rebuildPalette() {
	colors := a.doc.Palette()
	items := make([]fyne.CanvasObject, 0, len(colors))
	limit := len(colors)
	if limit > 512 {
		limit = 512
	}
	for i := 0; i < limit; i++ {
		c := colors[i]
		items = append(items, newSwatch(c, fyne.NewSize(28, 22), func() {
			a.doc.Brush = c
			a.doc.Tool = doc.ToolPen
			a.refreshChrome()
		}))
	}
	if len(items) == 0 {
		a.palBox.Objects = []fyne.CanvasObject{widget.NewLabel("No colors yet")}
	} else {
		a.palBox.Objects = []fyne.CanvasObject{container.NewGridWithColumns(4, items...)}
	}
	a.palBox.Refresh()
}

func (a *App) chooseColor() {
	if a.doc.LockPalette {
		dialog.ShowInformation("Locked", "Palette is locked.\nPlease select a color from the Palette list.", a.win)
		return
	}
	dialog.ShowColorPicker("Color", "Choose brush color", func(c color.Color) {
		if c == nil {
			return
		}
		r, g, b, _ := c.RGBA()
		a.doc.Brush = color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255}
		a.refreshChrome()
	}, a.win)
}

func (a *App) undo() {
	if a.doc.Undo() {
		a.view.Refresh()
		a.scroll.Refresh()
		a.rebuildPalette()
		a.refreshChrome()
	}
}

func (a *App) redo() {
	if a.doc.Redo() {
		a.view.Refresh()
		a.scroll.Refresh()
		a.rebuildPalette()
		a.refreshChrome()
	}
}

func (a *App) clear() {
	a.doc.Clear()
	a.view.Refresh()
	a.rebuildPalette()
	a.refreshChrome()
}

func (a *App) newImage() {
	wEntry := widget.NewEntry()
	hEntry := widget.NewEntry()
	wEntry.SetText(strconv.Itoa(a.doc.Width))
	hEntry.SetText(strconv.Itoa(a.doc.Height))
	dialog.ShowForm("New Image", "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Width", wEntry),
		widget.NewFormItem("Height", hEntry),
	}, func(ok bool) {
		if !ok {
			return
		}
		w, err1 := strconv.Atoi(strings.TrimSpace(wEntry.Text))
		h, err2 := strconv.Atoi(strings.TrimSpace(hEntry.Text))
		if err1 != nil || err2 != nil || w < 1 || h < 1 {
			dialog.ShowError(fmt.Errorf("width and height must be positive integers"), a.win)
			return
		}
		if w > 2048 || h > 2048 {
			dialog.ShowError(fmt.Errorf("maximum size is 2048x2048"), a.win)
			return
		}
		a.doc.Reset(w, h)
		a.view.Refresh()
		a.scroll.Refresh()
		a.rebuildPalette()
		a.refreshChrome()
	}, a.win)
}

func (a *App) openImage() {
	fd := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		if rc == nil {
			return
		}
		defer rc.Close()
		img, err := doc.Decode(rc)
		if err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		a.doc.LoadImage(img, pathFromURI(rc.URI()))
		a.view.Refresh()
		a.scroll.Refresh()
		a.rebuildPalette()
		a.refreshChrome()
	}, a.win)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".bmp"}))
	fd.Show()
}

func (a *App) quickSave() {
	if a.doc.FilePath != "" {
		if err := a.doc.Save(a.doc.FilePath); err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		dialog.ShowInformation("Saved", "Saved to "+a.doc.FilePath, a.win)
		return
	}
	a.saveImageAs()
}

func (a *App) saveImageAs() {
	fd := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		if uc == nil {
			return
		}
		defer uc.Close()
		path := pathFromURI(uc.URI())
		if path != "" && filepath.Ext(path) == "" {
			path += ".png"
		}
		if err := doc.Encode(uc, a.doc.Img, path); err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		a.doc.FilePath = path
		a.refreshChrome()
		dialog.ShowInformation("Saved", "Image saved.", a.win)
	}, a.win)
	fd.SetFileName("sprite.png")
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
	fd.Show()
}

func pathFromURI(u fyne.URI) string {
	if u == nil {
		return ""
	}
	p := u.Path()
	if runtime.GOOS == "windows" {
		p = strings.TrimPrefix(p, "/")
		p = filepath.FromSlash(p)
	}
	return p
}

type fixedWidth struct {
	width float32
}

func newFixedWidth(width float32, obj fyne.CanvasObject) fyne.CanvasObject {
	return container.New(&fixedWidth{width: width}, obj)
}

func (f *fixedWidth) MinSize(objects []fyne.CanvasObject) fyne.Size {
	h := float32(0)
	for _, o := range objects {
		s := o.MinSize()
		if s.Height > h {
			h = s.Height
		}
	}
	return fyne.NewSize(f.width, h)
}

func (f *fixedWidth) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(f.width, size.Height))
		o.Move(fyne.NewPos(0, 0))
	}
}
