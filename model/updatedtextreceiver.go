package model

type UpdatedTextReceiver interface {
	UpdateText(text string) error
}
