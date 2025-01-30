package db

import (
	"database/sql"
	"errors"

	"os"

	"dekr0/hd2_audio_db/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

/*
A helper function that find the primary key for a record that contain the
string representation of a Wwise hierarchy object type. Error only arise when
a type doesn't exist or *t is nil

[parameter]
rows - rows of database record about Wwise hierarchy object types
*t - a string representation of a Wwiser hierarchy object type

[return]
string - primary key for a record whose string representaion of a Wwise.
hierarchy object type matches the one provided by *t. Empty string when an
error occurs.
error - trivial
*/
func getHierarchyObjectTypePrimaryKey(
    rows []database.HierarchyObjectType, t string,
) (string, error) {
    for _, row := range rows {
        if row.Type == t {
            return row.DbID, nil
        }
    }
    return "", errors.New("Type " + t + "does not exist")
}

/*
A helper function for generate a Database primary key using UUID4

[return]
string - Return Empty string when an error occurs
error - trivial
*/
func genDBID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	
	idS := id.String()
	if idS == "" {
		return "", errors.New("Empty string when generate UUID4 string")
	}

	return idS, nil
}

func getDBConn() (*sql.DB, error) {
	godotenv.Load()

	dbPath := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
    return db, err
}

