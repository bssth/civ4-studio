package editor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"sync"
)

// ProgressBar is a custom progress bar with text label
type ProgressBar struct {
	sync.Mutex
	block    *fyne.Container
	progress *widget.ProgressBarInfinite
	label    *widget.Label
}

func (p *ProgressBar) GetBlock() *fyne.Container {
	return p.block
}

func (p *ProgressBar) Start(loadingText string) {
	p.Lock()
	defer p.Unlock()
	p.label.SetText(loadingText)
	p.block.Show()
	p.progress.Start()
}

func (p *ProgressBar) Stop() {
	p.Lock()
	defer p.Unlock()
	p.progress.Stop()
	p.block.Hide()
}

func GuiProgressBar() *ProgressBar {
	progress := widget.NewProgressBarInfinite()
	progressText := widget.NewLabel("Loading...")
	progressBlock := container.NewVBox(container.NewCenter(progressText), progress)
	progressBlock.Hide()
	return &ProgressBar{progress: progress, label: progressText, block: progressBlock}
}

// FormField is a common interface for all form fields
type FormField interface {
	GetKey() string
	GetValue() string
}

// SelectEntry is a form field with predefined values.
type SelectEntry struct {
	Key        string
	entry      *widget.SelectEntry
	valueLabel *widget.Label
}

func (s *SelectEntry) GetKey() string {
	return s.Key
}

func (s *SelectEntry) GetValue() string {
	return s.entry.Text
}

func GuiSelectEntry(c *fyne.Container, key string, label string, value string, variants []string, onChange func(string)) *SelectEntry {
	valueLabel := widget.NewLabel(GetLangString(value))
	e := widget.NewSelectEntry(variants)
	e.OnChanged = func(s string) {
		onChange(s)
		valueLabel.SetText(GetLangString(s))
	}
	e.Text = value
	c.Add(container.NewGridWithColumns(3, widget.NewLabel(label), e, valueLabel))
	return &SelectEntry{key, e, valueLabel}
}

type Entry struct {
	Key   string
	entry *widget.Entry
}

func (e *Entry) GetKey() string {
	return e.Key
}

func (e *Entry) GetValue() string {
	return e.entry.Text
}

func GuiTextField(c *fyne.Container, key string, label string, value string, onChange func(string)) *Entry {
	additionalLabel := widget.NewLabel(GetLangString(key))
	e := widget.NewEntry()
	e.OnChanged = func(s string) {
		onChange(s)
	}
	e.Text = value
	c.Add(container.NewGridWithColumns(3, widget.NewLabel(label), e, additionalLabel))
	return &Entry{key, e}
}

func GuiCheckbox(c *fyne.Container, label string, value bool, onChange func(bool)) {
	cb := widget.NewCheck(GetLangString(label), onChange)
	cb.Checked = value
	c.Add(cb)
}

func GuiNoMapLoaded(c *fyne.Container) {
	c.Add(container.NewCenter(canvas.NewText("To start editing, open a map or create a new one", color.White)))
}
