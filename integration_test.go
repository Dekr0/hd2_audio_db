package main

import (
	"context"
	"dekr0/hd2_audio_db/db"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {
	os.Setenv("DATA", "/mnt/d/Program Files/Steam/steamapps/common/Helldivers 2/data")
	data := os.Getenv("DATA")
	os.Setenv("GOOSE_DBSTRING", "build_15637")
	ctx := context.Background()
	if err := db.Generate(ctx, data); err != nil {
		t.Fatal(err)
	}
}
