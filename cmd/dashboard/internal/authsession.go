package internal

import (

)

var accessToken  string
var refreshToken string

func SaveTokenPair(access string, refresh string) {
	accessToken = access
	refreshToken = refresh
}