package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./db/clients.db")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            chat_id INTEGER PRIMARY KEY,
            phone TEXT,
            fio TEXT,
            city TEXT,
            speaker TEXT,
            date TEXT
        )
    `)
	return db, err
}

func UpsertUser(db *sql.DB, chatID int64, phone, fio, city, speaker string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(`
        INSERT INTO users (chat_id, phone, fio, city, speaker, date)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(chat_id) DO UPDATE SET
            phone=COALESCE(NULLIF(excluded.phone, ''), users.phone),
            fio=COALESCE(NULLIF(excluded.fio, ''), users.fio),
            city=COALESCE(NULLIF(excluded.city, ''), users.city),
            speaker=COALESCE(NULLIF(excluded.speaker, ''), users.speaker),
            date=excluded.date
    `, chatID, phone, fio, city, speaker, now)

	log.Printf("Сохраняем (%d, '%s', '%s', '%s', '%s', '%s')", chatID, phone, fio, city, speaker, now)
	return err
}

type User struct {
	ChatID  int64
	Phone   string
	Fio     string
	City    string
	Speaker string
	Date    string
}

func GetAllUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT chat_id, phone, fio, city, speaker, date FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ChatID, &u.Phone, &u.Fio, &u.City, &u.Speaker, &u.Date); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func GetUserByChatID(db *sql.DB, chatID int64) (*User, error) {
	row := db.QueryRow("SELECT chat_id, phone, fio, city, speaker, date FROM users WHERE chat_id = ?", chatID)
	var u User
	if err := row.Scan(&u.ChatID, &u.Phone, &u.Fio, &u.City, &u.Speaker, &u.Date); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
