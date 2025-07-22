package ui

type ResizableRenderer interface {
	Resize(x, y, width, height int)
	Render(updateScreen bool)
}
