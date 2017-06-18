package main

import (
	"github.com/gorilla/sessions"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var cookieStore = sessions.NewCookieStore([]byte("secret-session-salt"))

type Image struct {
	Name string
	URI  string
	// Add image size to struct?
}

type Data struct {
	Album      []Image
	IsLoggedIn bool
	Username   string
}

func IsImage(file os.FileInfo) bool {
	if file.IsDir() {
		return false
	}

	// To check tha file is an image
	var is func(string) bool = func(ftype string) bool {
		return strings.HasSuffix(strings.ToLower(file.Name()), "."+ftype)
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

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	// Show a list of photos in the uploads directory
	// clicking a photo downloads it
	files, err := ioutil.ReadDir("uploads/")
	if err != nil {
		http.Error(w, "Can't fetch uploads dir.", http.StatusInternalServerError)
		return
	}

	var album []Image = make([]Image, 0, len(files))

	for _, fileInfo := range files {
		if !IsImage(fileInfo) {
			continue
		}

		// Add image to album slice
		album = append(album, Image{
			Name: fileInfo.Name(),
			URI:  "gallery/" + fileInfo.Name(),
		})
	}

	tpl, err := template.ParseFiles("templates/home.html.tmpl")
	if err != nil {
		http.Error(w, "Parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	username, isLoggedIn := CurrentUser(req)

	var data Data = Data{
		Album:      album,
		Username:   username,
		IsLoggedIn: isLoggedIn,
	}

	tpl.Execute(w, data)
}

func GalleryHandler(w http.ResponseWriter, req *http.Request) {
	var resource string = req.URL.String()
	var tokens []string = strings.Split(resource, "/")
	if len(tokens) < 3 {
		// Did not provide an image name
		http.Error(w, "Image not found.", http.StatusNotFound)
		return
	}

	var imageName string = tokens[2]
	file, err := os.Open("uploads/" + imageName)
	if err != nil {
		http.Error(w, "Image not found in gallery.", http.StatusNotFound)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, "File info not found.", http.StatusInternalServerError)
		return
	}

	// Serve any images in the upload folder
	http.ServeContent(w, req, file.Name(), info.ModTime(), file)
}

func UploadsHandler(w http.ResponseWriter, req *http.Request) {
	// /upload form for uploading a new photo (requires a session)
	if !IsLoggedIn(req) {
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	if req.Method == "GET" {
		tpl, err := template.ParseFiles("templates/upload.html.tmpl")
		if err != nil {
			http.Error(w, "Parsing template: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tpl.Execute(w, nil)
		return
	}

	file, fheader, err := req.FormFile("image")
	if err != nil {
		http.Error(w, "Can't handle file "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if fileInfo, _ := os.Stat(fheader.Filename); !IsImage(fileInfo) {
		// Only can upload images
		http.Redirect(w, req, "/upload", http.StatusInternalServerError)
		return
	}

	// Not recomended on prod
	outputFile, err := os.Create("uploads/" + fheader.Filename)
	if err != nil {
		http.Error(
			w,
			"Error uploading file to destination. "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	defer outputFile.Close()

	// Write uplaoded byte stream to output file

	_, err = io.Copy(outputFile, file)
	if err != nil {
		http.Error(w, "Error writing to destination. "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/home", http.StatusFound)
}

func LoginHandler(w http.ResponseWriter, req *http.Request) {
	if IsLoggedIn(req) {
		http.Redirect(w, req, "/home", http.StatusFound)
		return
	}

	// login requires ssl (can hardcode user and password)
	if method := req.Method; method == "GET" {
		tpl, err := template.ParseFiles("templates/login.html.tmpl")
		if err != nil {
			http.Error(w, "Parsing template: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tpl.Execute(w, nil)
		return
	}

	username, authOk := CheckAuth(req)
	if !authOk {
		http.Redirect(w, req, "/login", http.StatusFound)
		return
	}

	session, err := cookieStore.Get(req, "session")
	if err != nil {
		// Posibly session has been tampered
		// TODO: erase session?
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["username"] = username
	session.Save(req, w)

	http.Redirect(w, req, "/home", http.StatusTemporaryRedirect)
}

func LogoutHandler(w http.ResponseWriter, req *http.Request) {
	// logout (deletes session)
	if !IsLoggedIn(req) {
		http.Redirect(w, req, "/home", http.StatusFound)
		return
	}

	// Destroy session
	session, _ := cookieStore.Get(req, "session")
	session.Options.MaxAge = -1
	session.Save(req, w)

	http.Redirect(w, req, "/home", http.StatusFound)
}

func main() {
	// TODO: Create uploads folder if does not exist

	// Static files handler
	http.Handle("/assets/",
		http.StripPrefix("/assets",
			http.FileServer(http.Dir("assets/"))))

	http.HandleFunc("/gallery/", GalleryHandler)
	http.HandleFunc("/upload", UploadsHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/logout", LogoutHandler)

	http.HandleFunc("/home", HomeHandler)

	// use templates and CSS to make it look pretty
	// use a ssl certificate (can be auto signed)
	log.Println("Server up and running on http://localhost:9000/ ...")
	http.ListenAndServe("localhost:9000", nil)
}
