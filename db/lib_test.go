package db

import (
	"context"
	"dekr0/hd2_audio_db/internal/database"
	wio "dekr0/hd2_audio_db/io"
	"dekr0/hd2_audio_db/parser"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func getAllArchivePath() []string {
	data := os.Getenv("DATA")

	c, err := conn()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 4)
	defer cancel()

	q := database.New(c)
	archives, err := q.GetAllArchive(ctx)
	if err != nil {
		panic(err)
	}

	path := make([]string, len(archives), len(archives))
	for i, archive := range archives {
		path[i] = filepath.Join(data, archive.Aid)
	}

	return path
}

func useDiscard() {
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				io.Discard,
				&slog.HandlerOptions{AddSource: true},
			),
		),
	)
}

func useStderr() {
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(
				os.Stderr,
				&slog.HandlerOptions{AddSource: true, Level: slog.LevelError},
			),
		),
	)
}

func TestWriteArchives(t *testing.T) {
	useStderr()
	data := os.Getenv("DATA")
	timeout, cancel := context.WithTimeout(context.Background(), time.Second * 4)
	defer cancel()
	if err := WriteArchives(timeout, data); err != nil {
		t.Fatal(err)
	}
}

func TestGatherSync(t *testing.T) {
	useStderr()
	MaxBankParser = 0
	for _, p := range getAllArchivePath() {
		gather(p, "")
	}
}

func TestGatherParallel(t *testing.T) {
	useStderr()
	MaxBankParser = 4
	for _, p := range getAllArchivePath() {
		gather(p, "")
	}
}

func testGatherInPlaceSync(t *testing.T) {
	useStderr()
	MaxBankParser = 0
	for _, p := range getAllArchivePath() {
		gatherInPlace(p)
	}
}

func testGatherInPlaceParallel(t *testing.T) {
	useStderr()
	MaxBankParser = 4
	for _, p := range getAllArchivePath() {
		gatherInPlace(p)
	}
}

func TestGatherEdgeCaseSync(t *testing.T) {
	useStderr()
	MaxBankParser = 0
	gather("/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/9ba626afa44a3aa3", "")
}

func TestGatherEdgeCaseParallel(t *testing.T) {
	useStderr()
	MaxBankParser = 0
	gather("/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/9ba626afa44a3aa3", "")
}

func testGatherInPlaceEdgeCaseSync(t *testing.T) {
	useStderr()
	MaxBankParser = 0
	gatherInPlace("/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/1a9e1e280e0112d8")
}

func testGatherInPlaceEdgeCaseParallel(t *testing.T) {
	useStderr()
	MaxBankParser = 0
	gatherInPlace("/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9")
}

func TestGenerate(t *testing.T) {
	useStderr()
	MaxBankParser = 0
	parser.MaxParser = 1
	if err := Generate(context.Background(), os.Getenv("DATA")); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkGather0(b *testing.B) {
	useDiscard()
	MaxBankParser = 0

	p := "/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9"
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	defer f.Close()
	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)
	parseHeader(&a, r)
	for range b.N {
		r.AbsSeekUnsafe(0)
		b.ResetTimer()
		parseBanks(&a, []database.InsertSoundbankParams{}, r, "", p)
	}
}

func BenchmarkGather4(b *testing.B) {
	useDiscard()
	MaxBankParser = 4

	p := "/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9"
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	defer f.Close()
	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)
	parseHeader(&a, r)
	for range b.N {
		r.AbsSeekUnsafe(0)
		b.ResetTimer()
		parseBanks(&a, []database.InsertSoundbankParams{}, r, "", p)
	}
}

func BenchmarkGather6(b *testing.B) {
	useDiscard()
	MaxBankParser = 6

	p := "/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9"
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	defer f.Close()
	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)
	parseHeader(&a, r)
	for range b.N {
		r.AbsSeekUnsafe(0)
		b.ResetTimer()
		parseBanks(&a, make([]database.InsertSoundbankParams, len(a.SoundBnks)), r, "", p)
	}
}

func BenchmarkGather8(b *testing.B) {
	useDiscard()
	MaxBankParser = 8

	p := "/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9"
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	defer f.Close()
	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)
	parseHeader(&a, r)
	for range b.N {
		r.AbsSeekUnsafe(0)
		b.ResetTimer()
		parseBanks(&a, []database.InsertSoundbankParams{}, r, "", p)
	}
}

func benchmarkGatherInPlace0(b *testing.B) {
	useDiscard()
	MaxBankParser = 0

	p := "/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9"
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	defer f.Close()
	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)
	parseHeader(&a, r)
	for range b.N {
		r.AbsSeekUnsafe(0)
		b.ResetTimer()
		parseBanksInPlace(&a, r, p)
	}
}

func benchmarkGatherInPlace4(b *testing.B) {
	useDiscard()
	MaxBankParser = 4

	p := "/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9"
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	defer f.Close()
	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)
	parseHeader(&a, r)
	for range b.N {
		r.AbsSeekUnsafe(0)
		b.ResetTimer()
		parseBanksInPlace(&a, r, p)
	}
}

func benchmarkGatherInPlace6(b *testing.B) {
	useDiscard()
	MaxBankParser = 6

	p := "/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9"
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	defer f.Close()
	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)
	parseHeader(&a, r)
	for range b.N {
		r.AbsSeekUnsafe(0)
		b.ResetTimer()
		parseBanksInPlace(&a, r, p)
	}
}

func benchmarkGatherInPlace8(b *testing.B) {
	useDiscard()
	MaxBankParser = 8

	p := "/mnt/D/Program Files/Steam/steamapps/common/Helldivers 2/data/e75f556a740e00c9"
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	defer f.Close()
	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)
	parseHeader(&a, r)
	for range b.N {
		r.AbsSeekUnsafe(0)
		b.ResetTimer()
		parseBanksInPlace(&a, r, p)
	}
}
