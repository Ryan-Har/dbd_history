package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var funcMap = template.FuncMap{
	"toJson": toJSON,
}

var templates = template.Must(
	template.New("").Funcs(funcMap).ParseGlob("templates/*.html"),
)

//var templates = template.Must(template.ParseGlob("templates/*.html"))

type Entry struct {
	ID        int
	Timestamp string // already formatted
	Character string
}

type HistoryPageData struct {
	Entries []Entry
	Count   int
}

type CharStat struct {
	Character string
	Count     int
	Image     string
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
	http.HandleFunc("/stats", handleStats)
	http.HandleFunc("GET /images/{id}", handleCharacterImage)

	//admin routes
	http.HandleFunc("GET /admin", handleAdmin)
	http.HandleFunc("POST /admin/characters", handleAddCharacter)
	http.HandleFunc("POST /admin/characters/{id}/update", handleUpdateCharacter)
	http.HandleFunc("POST /admin/characters/{id}/delete", handleDeleteCharacter)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER,
		character TEXT
	);
	CREATE TABLE IF NOT EXISTS characters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		image BLOB NOT NULL,
		image_type TEXT NOT NULL
	);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	characters, err := getCharacters()
	if err != nil {
		log.Println("Failed to retrieve characters:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	templates.ExecuteTemplate(w, "index.html", characters)
}

func handleCharacterImage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	var imageData []byte
	var imageType string

	err = db.QueryRow(`
        SELECT image, image_type
        FROM characters
        WHERE id = ?
    `, id).Scan(&imageData, &imageType)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Image retrieval error:", err)
		}

		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", imageType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")

	w.Write(imageData)
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

func handleStats(w http.ResponseWriter, r *http.Request) {

	characters, err := getCharacters()
	if err != nil {
		log.Println("Failed to retrieve characters:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Build a map of character counts from DB
	rows, err := db.Query("SELECT character, COUNT(*) as cnt FROM entries GROUP BY character")
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	counts := make(map[string]int)
	totalGames := 0
	for rows.Next() {
		var name string
		var cnt int
		if err := rows.Scan(&name, &cnt); err != nil {
			http.Error(w, "DB Error", http.StatusInternalServerError)
			return
		}
		counts[name] = cnt
		totalGames += cnt
	}

	// Build full stats list including unplayed characters
	var stats []CharStat
	for _, c := range characters {
		stats = append(stats, CharStat{
			Character: c.Name,
			Count:     counts[c.Name],
			Image:     c.Image,
		})
	}

	// Sort most → least
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	leastPlayed := stats[len(stats)-1].Character
	leastPlayedImage := ""
	for _, c := range characters {
		if c.Name == leastPlayed {
			leastPlayedImage = c.Image
			break
		}
	}

	// Last played (character + timestamp)
	var lastCharacter string
	var lastUnix int64
	err = db.QueryRow("SELECT character, timestamp FROM entries ORDER BY timestamp DESC LIMIT 1").
		Scan(&lastCharacter, &lastUnix)
	lastTime := ""
	if err == nil {
		lastTime = time.Unix(lastUnix, 0).Local().Format("02 Jan 2006 15:04")
	}
	lastPlayedImage := ""
	for _, c := range characters {
		if c.Name == lastCharacter {
			lastPlayedImage = c.Image
			break
		}
	}

	// Time since Twins
	var twinsUnix int64
	twinsLastTime := "Never"
	err = db.QueryRow(
		"SELECT timestamp FROM entries WHERE character = ? ORDER BY timestamp DESC LIMIT 1",
		"Twins",
	).Scan(&twinsUnix)
	if err == nil {
		twinsLastTime = time.Unix(twinsUnix, 0).Local().Format("02 Jan 2006 15:04")
	}

	// Get Twins image from characters map
	twinsImage := ""
	for _, c := range characters {
		if c.Name == "Twins" {
			twinsImage = c.Image
			break
		}
	}

	data := struct {
		Stats            []CharStat
		LastPlayed       string
		LastPlayedAt     string
		LastPlayedImage  string
		LeastPlayed      string
		LeastPlayedImage string
		TwinsLastTime    string
		TwinsImage       string
		TotalGames       int
	}{
		Stats:            stats,
		LastPlayed:       lastCharacter,
		LastPlayedAt:     lastTime,
		LastPlayedImage:  lastPlayedImage,
		LeastPlayed:      leastPlayed,
		LeastPlayedImage: leastPlayedImage,
		TwinsLastTime:    twinsLastTime,
		TwinsImage:       twinsImage,
		TotalGames:       totalGames,
	}

	templates.ExecuteTemplate(w, "stats.html", data)
}

func toJSON(v interface{}) template.JS {
	b, _ := json.Marshal(v)
	return template.JS(b)
}
