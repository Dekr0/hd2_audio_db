package main

import (
	"context"
	"testing"
)

func TestUpdateHelldiverSoundbanks(t *testing.T) {
	DATA = "D:\\Program Files\\Steam\\steamapps\\common\\Helldivers 2\\data"
	err := extractAllSoundbanks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
