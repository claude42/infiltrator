package ui

type PanelSelection struct {
	ModalImpl
}

func NewPanelSelection(content string, orientation Orientation) *PanelSelection {
	p := &PanelSelection{}
	p.SetContent(content, orientation)
	p.ModalImpl.Resize(0, 0, 0, 0)

	return p
}
