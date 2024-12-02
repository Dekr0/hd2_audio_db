package main

import (
	"context"
	"testing"
)

func TestUpdateHelldiverSoundbanks(t *testing.T) {
	DATA = "D:\\Program Files\\Steam\\steamapps\\common\\Helldivers 2\\data"
	err := extractHelldiverSoundbanks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
