package parser

import (
	wio "dekr0/hd2_audio_db/io"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func testParseArchiveHeaderCache(t *testing.T) {
	f, err := os.Open("/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/18235e0c9ec0e636")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := wio.NewReader(f, wio.ByteOrder)
	a := Archive{}
	ParseArchiveHeaderCache(&a, r)
	t.Log(a.NumFiles)
	t.Log(a.NumTypes)
	t.Log(a.Unk4Data)
	t.Log(a.AssetTypeCnts)

	r.AbsSeek(0)
	ParseArchiveHeader(&a, r)
	t.Log(a.NumFiles)
	t.Log(a.NumTypes)
	t.Log(a.Unk4Data)
	t.Log(a.AssetTypeCnts)
}

func BenchmarkParseArchiveHeader(b *testing.B) {
	f, err := os.Open("/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/18235e0c9ec0e636")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	for range b.N {
		f.Seek(0, io.SeekStart)
		r := wio.NewReader(f, wio.ByteOrder)
		a := Archive{}
		b.ResetTimer()
		ParseArchiveHeader(&a, r)
	}
}
func BenchmarkParseArchiveHeaderCache(b *testing.B) {
	f, err := os.Open("/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/18235e0c9ec0e636")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	for range b.N {
		f.Seek(0, io.SeekStart)
		r := wio.NewReader(f, wio.ByteOrder)
		a := Archive{}
		b.ResetTimer()
		ParseArchiveHeaderCache(&a, r)
	}
}

func BenchmarkParseAssetHeaders0(b *testing.B) {
	data := os.Getenv("DATA")
	path := filepath.Join(data, "18235e0c9ec0e636")
	MaxParser = 1
	f, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	for range b.N {
		f.Seek(0, io.SeekStart)
		r := wio.NewReader(f, wio.ByteOrder)

		a := Archive{}
		ParseArchiveHeader(&a, r)

		b.ResetTimer()
		ParseAssetHeadersCachePerThread(&a, r)
	}
}

func BenchmarkParseAssetHeaders4(b *testing.B) {
	data := os.Getenv("DATA")
	path := filepath.Join(data, "18235e0c9ec0e636")
	MaxParser = 4
	f, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	for range b.N {
		f.Seek(0, io.SeekStart)
		r := wio.NewReader(f, wio.ByteOrder)

		a := Archive{}
		ParseArchiveHeader(&a, r)

		b.ResetTimer()
		ParseAssetHeadersCachePerThread(&a, r)
	}
}

func BenchmarkParseAssetHeadersInPlace4(b *testing.B) {
	data := os.Getenv("DATA")
	path := filepath.Join(data, "18235e0c9ec0e636")
	MaxParser = 4
	f, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	for range b.N {
		f.Seek(0, io.SeekStart)
		r := wio.NewReader(f, wio.ByteOrder)

		a := Archive{}
		ParseArchiveHeader(&a, r)

		b.ResetTimer()
		ParseAssetHeaders(&a, r)
	}
}
