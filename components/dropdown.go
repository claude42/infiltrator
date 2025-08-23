package components

import "github.com/gdamore/tcell/v2"

type Dropdown struct {
	*Select

	open bool
}

func NewDropdown(options []string, key tcell.Key, do func(int)) *Dropdown {
	return &Dropdown{Select: NewSelect(options, key, do)}
}

func (d *Dropdown) Render(updateScreen bool) {
	if !d.visible {
		return
	}

	d.Select.Render(false)

	if d.open {
		d.renderDropDownBox()
	}

	if updateScreen {
		Screen.Show()
	}
}

func (d *Dropdown) renderDropDownBox() {
	style := d.CurrentStyler.Style()
	selectedStyle := style.Reverse(false)

	boxX, boxY, width, height := d.dropdownBoxDimensions()

	Screen.SetContent(boxX, boxY, tcell.RuneULCorner, nil, style)
	DrawChars(boxX+1, boxY, width-2, tcell.RuneHLine, style)
	Screen.SetContent(boxX+width-1, boxY, tcell.RuneURCorner, nil, style)

	for i := range d.Options {
		y := boxY + 1 + i*2
		Screen.SetContent(boxX, y, tcell.RuneVLine, nil, style)
		var textStyle tcell.Style
		if i == d.selected {
			textStyle = selectedStyle
		} else {
			textStyle = style
		}
		centerText(boxX+1, width-2, y, d.Options[i], ' ', textStyle)
		Screen.SetContent(boxX+width-1, y, tcell.RuneVLine, nil, style)

		Screen.SetContent(boxX, y+1, tcell.RuneLTee, nil, style)
		DrawChars(boxX+1, y+1, width-2, tcell.RuneHLine, style)
		Screen.SetContent(boxX+width-1, y+1, tcell.RuneRTee, nil, style)
	}

	Screen.SetContent(boxX, boxY+height-1, tcell.RuneLLCorner, nil, style)
	Screen.SetContent(boxX+width-1, boxY+height-1, tcell.RuneLRCorner, nil, style)
}

func (d *Dropdown) dropdownBoxDimensions() (boxX, boxY, width, height int) {
	_, screenHeight := Screen.Size()

	width = d.width
	height = 2*len(d.Options) + 1

	boxX = d.x
	boxY = d.y - 1
	// if there is not enough space below, try to render it above
	if boxY+height > screenHeight {
		boxY = d.y - height + 1
	}

	return
}

func (d *Dropdown) HandleEvent(ev tcell.Event) bool {
	if !d.IsActive() {
		return false
	}

	switch tev := ev.(type) {
	case *tcell.EventKey:
		if tev.Key() == d.key {
			d.open = !d.open
			RenderAll(true)
			return true
		}

		if !d.open {
			return false
		}

		switch tev.Key() {
		case tcell.KeyUp:
			d.PreviousOption()
			d.Render(true)
			return true
		case tcell.KeyDown:
			d.NextOption()
			d.Render(true)
			return true
		case tcell.KeyEnter:
			d.open = false
			if d.do != nil {
				d.do(d.selected)
			}
			RenderAll(true)
			return true
		case tcell.KeyEscape:
			d.open = false
			RenderAll(true)
			return true
		default:
			d.open = false
			RenderAll(true)
			return false
		}
	}

	return false
}
