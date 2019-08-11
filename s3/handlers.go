package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func mountRoutes() {
	http.HandleFunc("/train", handleTrain)
}

func handleTrain(w http.ResponseWriter, r *http.Request) {
	fileName, err := uploadFile(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"success\":\"true\", \"dataset\":\"" + fileName + "\"}"))
}

func uploadFile(w http.ResponseWriter, r *http.Request) (string, error) {
	// Reference: https://tutorialedge.net/golang/go-file-upload-tutorial/
	log.Println("Uploading file...")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	// FormFile returns the first file for the given key `dataset`,
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, meta, err := r.FormFile("dataset")
	if err != nil {
		return "", err
	}
	defer file.Close()

	log.Printf("Uploaded File: %+v\n", meta.Filename)
	log.Printf("File Size: %+v\n", meta.Size)
	log.Printf("MIME Header: %+v\n", meta.Header)

	// Create a temporary file within our temp directory that follows
	// a particular naming pattern
	os.MkdirAll(filepath.Join(".", "datasets"), os.ModePerm)
	tempFile, err := ioutil.TempFile("datasets", "dataset-train-"+meta.Filename)
	if err != nil {
		log.Println("Cannot create temp file:", err)
		return "", err
	}
	defer tempFile.Close()

	// Read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Cannot read temp file:", err)
		return "", err
	}

	// Write this byte array to our temporary file
	_, err = tempFile.Write(fileBytes)
	if err != nil {
		log.Println("Cannot write to temp file:", err)
	}

	// Get file's name and get out
	fileName := filepath.Base(tempFile.Name())

	return fileName, nil
}
