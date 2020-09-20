package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	Sender  string
	Text    string
	Time    time.Time
	TimeRaw int64
}

func (msg Message) String() string {
	return fmt.Sprintf("[%v] %s: %s", msg.TimeStr(), msg.Sender, msg.Text)
}

func (msg Message) TimeStr() string {
	return msg.Time.Format("2006-01-02 15:04")
}

var baseDate = time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)

func LoadMessages(f func(sms Message)) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	fn := filepath.Join(homeDir, "Library/Messages/chat.db")
	db, err := sql.Open("sqlite3", fn)
	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT h.uncanonicalized_id, m.text, m.date FROM message m INNER JOIN handle h ON m.handle_id = h.rowid WHERE m.service = 'SMS' AND h.uncanonicalized_id IS NOT NULL AND m.text IS NOT NULL ORDER BY m.date DESC LIMIT 10000")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var m Message
	for rows.Next() {
		var date int64
		err := rows.Scan(&m.Sender, &m.Text, &date)
		if err != nil {
			return err
		}
		m.Time = baseDate.Add(time.Duration(date))
		m.TimeRaw = date
		f(m)
	}
	return rows.Err()
}
