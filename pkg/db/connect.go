package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Open(dbPath string) (*sql.DB, error) {
	main, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// if os.Getenv("WAL_MODE") == "true" {
	// 	_, err = main.Exec("PRAGMA journal_mode=wal")
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	print("WAL mode enabled")
	// }

	return main, nil
}
