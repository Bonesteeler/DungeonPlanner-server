package ui

import (
	"dungeonplannerserver-dashboard/internal/queries"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewLoginUI() fyne.CanvasObject {


	return container.NewVBox(
		NewLoginCard(),
		NewSignupCard(),
	)
}

func NewLoginCard() fyne.Widget {
	email := widget.NewEntry()
	email.SetPlaceHolder("Email")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	loginButton := widget.NewButton("Login", func() {
		response, err := queries.Login(email.Text, password.Text)
		if err != nil {
			fmt.Println("Login failed:", err)
			return
		}
		fmt.Println("Login successful:", response, " err:", err)
	})

  loginCard := widget.NewCard("", "", container.NewVBox(
		email,
		password,
		loginButton,
  ))
	return loginCard
}

func NewSignupCard() fyne.Widget {
	email := widget.NewEntry()
	email.SetPlaceHolder("Email")

	username := widget.NewEntry()
	username.SetPlaceHolder("Username")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	signupButton := widget.NewButton("Sign Up", func() {
		response, err := queries.Signup(email.Text, username.Text, password.Text)
		if err != nil {
			fmt.Println("Signup failed:", err)
			return
		}
		fmt.Println("Signup successful:", response, " err:", err)
	})

  signupCard := widget.NewCard("", "", container.NewVBox(
		email,
		username,
		password,
		signupButton,
  ))
	return signupCard
}