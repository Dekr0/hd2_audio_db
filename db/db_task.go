package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path"

	"os"
	"strings"

	"dekr0/hd2_audio_db/internal/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"dekr0/hd2_audio_db/log"
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

/*
 Read the data from all files supported by 
*/
func UpdateSoundLabelsFromFileNoStruct(
    pathArgs string, ctx context.Context,
) error {
    paths := strings.Split(pathArgs, ";")

    db, err := getDBConn()
    if err != nil {
        return err
    }
    defer db.Close()

    dbQueries := database.New(db)

    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    queryWithTx := dbQueries.WithTx(tx)

    for _, p := range paths {
        var data LabelFile

        buf, err := os.ReadFile(p)
        if err != nil {
            log.DefaultLogger.Warn("Error opening file", "file_path", p, "error", err)
            continue
        }

        err = json.Unmarshal(buf, &data)
        if err != nil {
            log.DefaultLogger.Warn("Error parsing file", "file_path", p, "error", err)
            continue
        }

        for label, sourceId := range data.Sources {
            err := queryWithTx.UpdateSoundLabelBySourceId(
                ctx,
                database.UpdateSoundLabelBySourceIdParams{
                    Label: label,
                    WwiseShortID: sourceId,
                },
            )
            if err != nil {
                tx.Rollback()
                return err
            }
        }
    }

    err = tx.Commit()
    if err != nil {
        tx.Rollback()
        return err
    }

    return nil
}

func UpdateSoundLabelsFromFolderNoStruct(
    pathArg string, ctx context.Context,
) error {
    jsonFiles, err := os.ReadDir(pathArg)
    if err != nil {
        return err
    }

    db, err := getDBConn()
    if err != nil {
        return err
    }
    defer db.Close()

    dbQueries := database.New(db)

    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }

    queryWithTx := dbQueries.WithTx(tx)

    for _, f := range jsonFiles {
        var data LabelFile

        if !strings.HasSuffix(f.Name(), ".json") {
            continue
        }

        filename := path.Join(pathArg, f.Name())

        buf, err := os.ReadFile(filename)
        if err != nil {
            log.DefaultLogger.Warn(
                "Error opening file", 
                "file_path", filename, 
                "error", err,
            )
            continue
        }

        err = json.Unmarshal(buf, &data)
        if err != nil {
            log.DefaultLogger.Warn(
                "Error parsing file", 
                "file_path", filename, 
                "error", err,
            )
            continue
        }

        for label, sourceId := range data.Sources {
            err := queryWithTx.UpdateSoundLabelBySourceId(
                ctx,
                database.UpdateSoundLabelBySourceIdParams{
                    Label: label,
                    WwiseShortID: sourceId,
                },
            )
            if err != nil {
                tx.Rollback()
                return err
            }
        }
    }

    err = tx.Commit()
    if err != nil {
        tx.Rollback()
        return err
    }

    return nil
}
