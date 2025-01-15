package db

import (
	"encoding/json"
	"os"
	"testing"
)

func TestUpdateHelldiverSoundbanks(t *testing.T) {
}

func TestLabelReading(t *testing.T) {
    type Data struct {
        Sources map[string]string `json:"sources"`
    }

    buf, err := os.ReadFile("./label/weapons_stratagems/safe/120mm.json")
    if err != nil {
        t.Fatal(err)
    }

    var data Data

    err = json.Unmarshal(buf, &data)
    if err != nil {
        t.Fatal(err)
    }

    for k, v := range data.Sources {
        t.Logf("Label: %s. Id: %s\n", k, v)
    }
}
