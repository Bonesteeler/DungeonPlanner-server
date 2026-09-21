package ui

import (
	"dungeonplannerserver-dashboard/internal/queries"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewLoginUI() fyne.CanvasObject {
	username := widget.NewEntry()
	username.SetPlaceHolder("Username")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	loginButton := widget.NewButton("Login", func() {
		response, err := queries.Login(username.Text, password.Text)
		if err != nil {
			fmt.Println("Login failed:", err)
			return
		}
		fmt.Println("Login successful:", response, " err:", err)
	})

  card := widget.NewCard("", "", container.NewVBox(
		username,
		password,
		loginButton,
  ))

	return container.NewVBox(
		card,
	)
}
