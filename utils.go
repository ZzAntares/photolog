package main

import (
	"net/http"
	"os"
	"strings"
)

func IsImage(file os.FileInfo) bool {
	if file.IsDir() {
		return false
	}

	return HasImageName(file.Name())
}

func HasImageName(name string) bool {
	var is func(string) bool = func(ftype string) bool {
		return strings.HasSuffix(strings.ToLower(name), "."+ftype)
	}

	return is("jpg") || is("png") || is("jpeg")
}

func IsLoggedIn(req *http.Request) bool {
	session, err := cookieStore.Get(req, "session")
	if err != nil {
		// Session tampered
		return false
	}

	_, ok := session.Values["username"]
	return ok
}

func CurrentUser(req *http.Request) (string, bool) {
	session, err := cookieStore.Get(req, "session")
	if err != nil {
		return "", false
	}

	user, ok := session.Values["username"]
	if user != nil {
		var username string = user.(string)

		return username, ok
	}

	return "", ok // TODO: A username shouldn't be allowed to be empty string
}

func CheckAuth(req *http.Request) (username string, authenticated bool) {
	username, password := req.FormValue("username"), req.FormValue("password")

	if username == "demo" && password == "demo" {
		return username, true
	}

	return "", false
}
