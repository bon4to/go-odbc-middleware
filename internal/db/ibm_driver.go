package db

import (
	"database/sql"
	"fmt"

	_ "github.com/ibmdb/go_ibm_db"
)

func makeIBMString(Host string, Port string, SourceName string, User string, Password string) string {
	dsn := fmt.Sprintf(
		"HOSTNAME=%s;PORT=%s;DATABASE=%s;UID=%s;PWD=%s",
		Host, Port, SourceName, User, Password,
	)
	return dsn
}

func openIBMConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("go_ibm_db", dsn)
	return db, err
}
