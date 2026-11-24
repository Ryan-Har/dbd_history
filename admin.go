package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	characters, err := getCharacters()
	if err != nil {
		http.Error(w, "Failed to retrieve characters", http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "admin.html", characters); err != nil {
		log.Println("Admin template error:", err)
	}
}

func handleAddCharacter(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "Upload too large or invalid form", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))

	if name == "" {
		http.Error(w, "Character name is required", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Character image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(io.LimitReader(file, (5<<20)+1))
	if err != nil || len(imageData) == 0 || len(imageData) > 5<<20 {
		http.Error(w, "Invalid image size", http.StatusBadRequest)
		return
	}

	imageType := http.DetectContentType(imageData)

	allowedTypes := map[string]bool{
		"image/webp": true,
		"image/png":  true,
		"image/jpeg": true,
		"image/gif":  true,
	}

	if !allowedTypes[imageType] {
		http.Error(w, "Unsupported image format", http.StatusBadRequest)
		return
	}

	err = addCharacter(name, imageData, imageType)
	if err != nil {
		log.Println("Failed to add character:", err)
		http.Error(w, "Failed to save character", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func handleDeleteCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid character ID", http.StatusBadRequest)
		return
	}

	err = deleteCharacter(id)
	if err != nil {
		log.Println("Failed to delete character:", err)
		http.Error(w, "Failed to delete character", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func handleUpdateCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid character ID", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "Upload too large or invalid form", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "Character name is required", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Character image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(io.LimitReader(file, (5<<20)+1))
	if err != nil || len(imageData) == 0 || len(imageData) > 5<<20 {
		http.Error(w, "Invalid image size", http.StatusBadRequest)
		return
	}

	imageType := http.DetectContentType(imageData)

	allowedTypes := map[string]bool{
		"image/webp": true,
		"image/png":  true,
		"image/jpeg": true,
		"image/gif":  true,
	}

	if !allowedTypes[imageType] {
		http.Error(w, "Unsupported image format", http.StatusBadRequest)
		return
	}

	err = updateCharacter(id, name, imageData, imageType)
	if err != nil {
		log.Println("Failed to update character:", err)
		http.Error(w, "Failed to update character", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
