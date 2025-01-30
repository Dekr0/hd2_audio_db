package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"dekr0/hd2_audio_db/internal/database"
	"dekr0/hd2_audio_db/log"
)

func UpdateSoundLabelsFromFolder(
    dir string, ctx context.Context,
) error {
    labelFiles, err := os.ReadDir(dir)
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

    for _, labelFile := range labelFiles {
        var data LabelFileVersion

        if labelFile.IsDir() {
            continue
        }

        labelFilePath := path.Join(dir, labelFile.Name())

        buf, err := os.ReadFile(labelFilePath)
        if err != nil {
            log.DefaultLogger.Warn("Error opening file", 
                "file_path", labelFilePath,
                "error", err)
            continue
        }

        err = json.Unmarshal(buf, &data)
        if err != nil {
            log.DefaultLogger.Warn("Error parsing file", 
                "file_path", labelFilePath,
                "error", err)
            continue
        }

        if data.Version == 1 {
            if err = updateUsingVersionOne(buf, *queryWithTx, ctx); err != nil {
                log.DefaultLogger.Warn("Error importing label", 
                    "file_path", labelFilePath, 
                    "error", err)
            }
        } else if data.Version == 2 {
            if err = updateUsingVersionTwo(buf, *queryWithTx, ctx); err != nil {
                log.DefaultLogger.Warn("Error importing label", 
                    "file_path", labelFilePath, 
                    "error", err)
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

func updateUsingVersionOne(
    buf []byte, queryWithTx database.Queries, ctx context.Context) error {

    var data LabelFileVersion1

    err := json.Unmarshal(buf, &data)
    if err != nil {
        return err
    }
    
    for label, sourceID := range data.Sources {
        if err = queryWithTx.UpdateLabelByWwiseShortId(ctx,
            database.UpdateLabelByWwiseShortIdParams{
                Label: label,
                WwiseShortID: sourceID,
            },
        ); err != nil {
            return err
        }
    }

    return nil
}

func updateUsingVersionTwo(
    buf []byte, queryWithTx database.Queries, ctx context.Context) error {

    var data LabelFileVersion2

    err := json.Unmarshal(buf, &data)
    if err != nil {
        return err
    }
    
    for _, container := range data.Containers {
        if container.ContainerID != "" {
            if err = queryWithTx.UpdateInfoByWwiseObjectId(ctx,
                database.UpdateInfoByWwiseObjectIdParams{
                    Label: container.Name,
                    Description: container.Desc,
                    WwiseObjectID: container.ContainerID,
                },
                ); err != nil {
                return err
            }
        }
        for i, source := range container.Sources {
            if err = queryWithTx.UpdateInfoByWwiseShortId(ctx,
                database.UpdateInfoByWwiseShortIdParams{
                    Label: fmt.Sprintf("%s_%i", container.Name, i), 
                    Description: container.Desc,
                    WwiseShortID: source,
                },
            ); err != nil {
                return err
            }
        }
    }

    for _, source := range data.Sources {
        if err = queryWithTx.UpdateLabelByWwiseShortId(ctx,
            database.UpdateLabelByWwiseShortIdParams{
                Label: source.Name,
                WwiseShortID: source.Id,
            },
        ); err != nil {
            return err
        }
    }

    return nil
}
