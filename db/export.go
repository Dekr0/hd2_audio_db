package db

import (
	"context"
	"dekr0/hd2_audio_db/internal/complete"
	"dekr0/hd2_audio_db/internal/id"
	"log/slog"
)

func ExportID(ctx context.Context) error {
	db, err := connV("BUILD")
	if err != nil {
		return err
	}
	defer db.Close()

	query := complete.New(db)
	sids, err := query.SourceIdUnique(ctx)
	if err != nil {
		return err
	}
	hids, err := query.HierarchyIdUnique(ctx)
	if err != nil {
		return err
	}
	db.Close()
	
	db, err = conn()
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			slog.Error("Failed to rollback", "error", err)
		}
	}()

	// multi-thread? since it's writting different table. They are independent.
	idQuery := id.New(db).WithTx(tx)
	for _, hid := range hids {
		if err := idQuery.InsertHierarchy(ctx, hid); err != nil {
			return err
		}
	}
	for _, sid := range sids {
		if err := idQuery.InsertSource(ctx, sid); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
