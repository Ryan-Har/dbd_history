package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var templates = template.Must(template.ParseGlob("templates/*.html"))

type Entry struct {
	ID        int
	Timestamp string // already formatted
	Character string
}

type HistoryPageData struct {
	Entries []Entry
	Count   int
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./db/characters.db")
	if err != nil {
		log.Fatal(err)
	}
	initDB()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/submit", handleSubmit)
	http.HandleFunc("/history", handleHistory)
	http.HandleFunc("/delete", handleDelete)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER,
		character TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", characters)
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	character := r.FormValue("character")
	_, err := db.Exec("INSERT INTO entries (timestamp, character) VALUES (?, ?)", time.Now().Unix(), character)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/history", http.StatusSeeOther)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, timestamp, character FROM entries ORDER BY timestamp DESC")
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var history []Entry
	for rows.Next() {
		var id int
		var timestampUnix int64
		var character string

		if err := rows.Scan(&id, &timestampUnix, &character); err != nil {
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}

		formatted := time.Unix(timestampUnix, 0).Format("02 Jan 2006 15:04")
		history = append(history, Entry{
			ID:        id,
			Timestamp: formatted,
			Character: character,
		})
	}

	data := HistoryPageData{
		Entries: history,
		Count:   len(history),
	}

	templates.ExecuteTemplate(w, "history.html", data)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM entries WHERE id = ?", id)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/history", http.StatusSeeOther)
}
