package db

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dekr0/hd2_audio_db/internal/database"
	wio "dekr0/hd2_audio_db/io"
	"dekr0/hd2_audio_db/parser"

	_ "github.com/mattn/go-sqlite3"
)

var MaxArchiveReader = 4
var MaxBankParser = 0
var HircMetric = 0

func conn() (*sql.DB, error) {
	p := os.Getenv("GOOSE_DBSTRING")
	db, err := sql.Open("sqlite3", p)
	return db, err
}

func WriteArchives(ctx context.Context, data string) error {
	f, err := os.Open(data)
	if err != nil {
		return err
	}

	timeout, cancel := context.WithTimeout(ctx, time.Second * 8)
	defer cancel()

	sem := make(chan struct{}, MaxArchiveReader)

	var w sync.WaitGroup

	for {
		entries, err := f.ReadDir(1024)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		select {
		case <- timeout.Done():
			return timeout.Err()
		case sem <- struct{}{}:
			w.Add(1)
			go writeArchives(&w, ctx, data, entries)
		default:
			writeArchives(nil, ctx, data, entries)
		}
	}

	w.Wait()

	return nil
}

func writeArchives(
	w *sync.WaitGroup,
	ctx context.Context,
	data string,
	entries []os.DirEntry,
) {
	if w != nil {
		defer w.Done()
	}
	c, err := conn()
	if err != nil {
		slog.Error("Failed to create database connection")
		panic(err)
	}
	defer c.Close()

	q := database.New(c)

	tx, err := c.Begin()
	if err != nil {
		slog.Error("Failed to start a transaction")
		panic(err)
	}

	withTx := q.WithTx(tx)
	for _, entry := range entries {
		if entry.IsDir() { continue }

		ext := filepath.Ext(entry.Name())
		if strings.Compare(ext, ".stream") == 0 { continue }
		if strings.Compare(ext, ".gpu_resources") == 0 { continue }
		if strings.Compare(ext, ".ini") == 0 { continue }
		if strings.Compare(ext, ".data") == 0 { continue }
		if strings.Contains(ext, "patch") { continue }

		stat, err := os.Lstat(filepath.Join(data, entry.Name()))
		if err != nil {
			slog.Error("Failed to obtain date of modified", "archive", entry.Name(), "error", err)
			continue
		}

		p := database.InsertArchiveParams{
			Aid: entry.Name(), 
			Tags: "",
			Categories: "",
			DateModified: stat.ModTime().Format(time.UnixDate),
		}
		if err := withTx.InsertArchive(ctx, p); err != nil {
			slog.Error("Failed to insert archive", "archive", entry.Name())
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Failed to commit archive insertion transaction")
		panic(err)
	}

	if err := c.Close(); err != nil {
		slog.Error("Failed to close database connection")
		panic(err)
	}
}

func Generate(ctx context.Context, data string) error {
	c, err := conn()
	if err != nil {
		return err
	}
	defer c.Close()

	timeout, cancel := context.WithTimeout(ctx, time.Second * 32)
	defer cancel()

	q := database.New(c)
	archives, err := q.GetAllArchive(ctx)
	if err != nil {
		return err
	}
	c.Close()

	assetInsert := []database.InsertAssetParams{}
	bankInsert := []database.InsertSoundbankParams{}
	hircInsert := []database.InsertHierarchyParams{}
	soundInsert := []database.InsertSoundParams{}
	for _, archive := range archives {
		select {
		case <- timeout.Done():
			return ctx.Err()
		default:
			localAssetInsert, localBankInsert, localHircInsert, localSoundInsert := gather(
				filepath.Join(data, archive.Aid),
				archive.Aid,
			)
			assetInsert = append(assetInsert, localAssetInsert...)
			bankInsert = append(bankInsert, localBankInsert...)
			hircInsert = append(hircInsert, localHircInsert...)
			soundInsert = append(soundInsert, localSoundInsert...)
		}
	}
	fmt.Println(HircMetric)

	c, err = conn()
	if err != nil {
		return err
	}
	defer c.Close()

	tx, err := c.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	qTx := database.New(c).WithTx(tx)
	for _, a := range assetInsert {
		if err := qTx.InsertAsset(ctx, a); err != nil {
			panic(err)
		}
	}
	for _, b := range bankInsert {
		if err := qTx.InsertSoundbank(ctx, b); err != nil {
			panic(err)
		}
	}
	for _, h := range hircInsert {
		if err := qTx.InsertHierarchy(ctx, h); err != nil {
			panic(err)
		}
	}
	for _, s := range soundInsert {
		if err := qTx.InsertSound(ctx, s); err != nil {
			panic(err)
		}
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}

	return c.Close()
}

func gather(p string, aid string) (
	assetInsert []database.InsertAssetParams,
	bankInsert []database.InsertSoundbankParams,
	hircInsert []database.InsertHierarchyParams,
	soundInsert []database.InsertSoundParams,
) {
	a := parser.Archive{}

	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}
	defer f.Close()

	r := wio.NewReader(f, wio.ByteOrder)

	parseHeader(&a, r)
	assetInsert = make([]database.InsertAssetParams, len(a.Headers))
	bankInsert = make([]database.InsertSoundbankParams, len(a.SoundBnks))
	for i, a := range a.Headers {
		assetInsert[i].Aid = aid
		assetInsert[i].Fid = int64(a.FileID)
		if a.FileID == 0 {
			slog.Error("Detected zero file ID", "aid", aid, "fid", a.FileID, "tid", a.TypeID)
			panic("Panic")
		}
		assetInsert[i].Tid = int64(a.TypeID)
		assetInsert[i].DataOffset = int64(a.DataOffset)
		assetInsert[i].StreamFileOffset = int64(a.StreamOffset)
		assetInsert[i].GpuRsrcOffset = int64(a.GPURsrcOffset)
		assetInsert[i].Unknown01 = int64(a.UnknownU64A)
		assetInsert[i].Unknown02 = int64(a.UnknownU64B)
		assetInsert[i].DataSize = int64(a.DataSize)
		assetInsert[i].StreamSize = int64(a.StreamSize)
		assetInsert[i].GpuRsrcSize = int64(a.StreamSize)
		assetInsert[i].Unknown03 = int64(a.UnknownU32A)
		assetInsert[i].Unknown04 = int64(a.UnknownU32B)
	}

	bankInsert, hircInsert, soundInsert = parseBanks(&a, bankInsert, r, aid, p)

	return assetInsert, bankInsert, hircInsert, soundInsert
}

func gatherInPlace(p string) () {
	a := parser.Archive{}

	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}
	defer f.Close()

	r := wio.NewReader(f, wio.ByteOrder)

	parseHeader(&a, r)
	parseBanksInPlace(&a, r, p)
}

func parseHeader(a *parser.Archive, r *wio.Reader) {
	parser.ParseArchiveHeader(a, r)
	parser.ParseAssetHeaders(a, r)
}

type ShareRsrc struct {
	m           sync.Mutex
	hircInsert  []database.InsertHierarchyParams
	soundInsert []database.InsertSoundParams
}

func parseBanks(
	a *parser.Archive,
	bankInsert []database.InsertSoundbankParams,
	r *wio.Reader,
	aid string,
	p string,
) (
	[]database.InsertSoundbankParams,
	[]database.InsertHierarchyParams,
	[]database.InsertSoundParams,
) {
	sem := make(chan struct{}, MaxBankParser)

	shareRsrc := ShareRsrc{
		hircInsert: []database.InsertHierarchyParams{}, 
		soundInsert: []database.InsertSoundParams{},
	}

	var w sync.WaitGroup

	for i, b := range a.SoundBnks {
		var path string = ""
		for _, w := range a.Deps {
			if a.Headers[w].FileID == a.Headers[b].FileID {
				r.AbsSeekUnsafe(uint(a.Headers[w].DataOffset))
				data := make([]byte, a.Headers[w].DataSize, a.Headers[w].DataSize)
				if err := r.Read(data); err != nil {
					slog.Error(
						"Failed to read data of wwise dependency",
						"path", p,
						"fid", a.Headers[w].FileID,
					)
					panic(err)
				}
				path = string(
					bytes.ReplaceAll(
						bytes.ReplaceAll(data[5:], []byte{'\u0000'}, []byte{}),
						[]byte{'/'},
						[]byte{'_'},
					),
				)
				break
			}
		}

		h := &a.Headers[b]
		bankInsert[i].Aid = aid
		bankInsert[i].Fid = int64(h.FileID)
		bankInsert[i].Path = path
		bankInsert[i].Name = ""
		bankInsert[i].Categories = ""

		select {
		case sem <- struct{}{}:
			w.Add(1)
			go parseBank(&shareRsrc, &w, aid, h.FileID, p, h)
		default:
			r.AbsSeekUnsafe(uint(h.DataOffset + 16))
			hirc := parser.ParseBank(r, h.DataOffset + uint64(h.DataSize))
			if hirc == nil {
				slog.Warn("Missing hierarchy", "path", p, "aid", aid, "fid", h.FileID)
				continue
			}
			HircMetric += int(hirc.Header)

			Fid := int64(h.FileID)
			hircInsert := make([]database.InsertHierarchyParams, len(hirc.Hierarchy))
			for i, h := range hirc.Hierarchy {
				hircInsert[i] = database.InsertHierarchyParams{
					Aid: aid,
					Fid: Fid,
					Hid: int64(h.ID),
					Type: parser.HircTypeName[h.Type],
					Parent: int64(h.Parent),
					Label: "",
					Tags: "",
					Description: "",
				}
			}
			soundInsert := make([]database.InsertSoundParams, len(hirc.Sound))
			for i, s := range hirc.Sound {
				soundInsert[i] = database.InsertSoundParams{
					Aid: aid,
					Fid: Fid,
					Hid: int64(hirc.Hierarchy[i].ID),
					Sid: int64(s.SourceID),
				}
			}
			shareRsrc.m.Lock()
			shareRsrc.hircInsert = append(shareRsrc.hircInsert, hircInsert...)
			shareRsrc.soundInsert = append(shareRsrc.soundInsert, soundInsert...)
			shareRsrc.m.Unlock()
		}
	}
	w.Wait()

	return bankInsert, shareRsrc.hircInsert, shareRsrc.soundInsert
}

func parseBank(
	s *ShareRsrc, w *sync.WaitGroup,
	aid string, fid uint64, p string,
	h *parser.AssetHeader,
) {
	defer w.Done()
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}
	r := wio.NewReader(f, wio.ByteOrder)
	r.AbsSeekUnsafe(uint(h.DataOffset + 16))
	hirc := parser.ParseBank(r, h.DataOffset + uint64(h.DataSize))
	if hirc == nil {
		slog.Warn("Missing hierarchy", "path", p, "aid", aid, "fid", fid)
		return
	}
	Fid := int64(fid)
	hircInsert := make([]database.InsertHierarchyParams, len(hirc.Hierarchy))
	for i, h := range hirc.Hierarchy {
		hircInsert[i] = database.InsertHierarchyParams{
			Aid: aid,
			Fid: Fid,
			Hid: int64(h.ID),
			Type: parser.HircTypeName[h.Type],
			Parent: int64(h.Parent),
			Label: "",
			Tags: "",
			Description: "",
		}
	}
	soundInsert := make([]database.InsertSoundParams, len(hirc.Sound))
	for i, s := range hirc.Sound {
		soundInsert[i] = database.InsertSoundParams{
			Aid: aid,
			Fid: Fid,
			Hid: int64(hirc.Hierarchy[i].ID),
			Sid: int64(s.SourceID),
		}
	}
	s.m.Lock()
	s.hircInsert = append(s.hircInsert, hircInsert...)
	s.soundInsert = append(s.soundInsert, soundInsert...)
	s.m.Unlock()
}

func parseBanksInPlace(a *parser.Archive, r *wio.Reader, p string) {
	sem := make(chan struct{}, MaxBankParser)
	var w sync.WaitGroup

	for _, b := range a.SoundBnks {
		var path string = ""
		for _, w := range a.Deps {
			if a.Headers[w].FileID == a.Headers[b].FileID {
				r.AbsSeekUnsafe(uint(a.Headers[w].DataOffset))
				data := make([]byte, a.Headers[w].DataSize, a.Headers[w].DataSize)
				if err := r.Read(data); err != nil {
					slog.Error(
						"Failed to read data of wwise dependency",
						"path", p,
						"fid", a.Headers[w].FileID,
					)
					panic(err)
				}
				path = string(
					bytes.ReplaceAll(
						bytes.ReplaceAll(data[5:], []byte{'\u0000'}, []byte{}),
						[]byte{'/'},
						[]byte{'_'},
					),
				)
				break
			}
		}
		h := &a.Headers[b]
		select {
		case sem <- struct{}{}:
			w.Add(1)
			go parseBankInPlace(&w, p, h)
		default:
			r.AbsSeekUnsafe(uint(h.DataOffset + 16))

			data := make([]byte, h.DataSize - 16, h.DataSize - 16)
			r.ReadUnsafe(data)

			ir := wio.NewInPlaceReader(data, wio.ByteOrder)
			hirc := parser.ParseBankInPlace(ir, h.DataOffset + uint64(h.DataSize))
			if hirc == nil {
				slog.Warn("Missing hierarchy", "path", p, "bank", path)
			}
		}
	}
	w.Wait()
}

func parseBankInPlace(w *sync.WaitGroup, p string, h *parser.AssetHeader) {
	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}

	r := wio.NewReader(f, wio.ByteOrder)

	r.AbsSeekUnsafe(uint(h.DataOffset + 16))

	data := make([]byte, h.DataSize - 16, h.DataSize - 16)
	r.ReadUnsafe(data)
	f.Close()

	ir := wio.NewInPlaceReader(data, wio.ByteOrder)
	parser.ParseBankInPlace(ir, h.DataOffset + uint64(h.DataSize))
	w.Done()
}
