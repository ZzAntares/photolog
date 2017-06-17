package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Image struct {
	Name string
	URI  string
	// Add size to struct?
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

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	// Show a list of photos in the uploads directory
	// clicking a photo downloads it
	tpl, err := template.ParseFiles("templates/home.html.tmpl")
	if err != nil {
		http.Error(w, "Parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
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

	tpl.Execute(w, album)
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
}

func LoginHandler(w http.ResponseWriter, req *http.Request) {
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

	io.WriteString(w, "POST isn't supported yet, but I got your details!")
	fmt.Println(req.PostForm)
}

func LogoutHandler(w http.ResponseWriter, req *http.Request) {
	// logout (deletes session)
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
