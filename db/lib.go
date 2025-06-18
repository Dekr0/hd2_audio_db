package db

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"dekr0/hd2_audio_db/internal/database"
	wio "dekr0/hd2_audio_db/io"
	"dekr0/hd2_audio_db/parser"

	_ "github.com/mattn/go-sqlite3"
)

var MaxDirReader = 4
var MaxArchiveReder = 2
var MaxBankParser = 4
var MaxBankWriter = 4
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

	sem := make(chan struct{}, MaxDirReader)

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

	timeout, cancel := context.WithTimeout(ctx, time.Second * 360)
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
				if err := r.ReadFull(data); err != nil {
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
		if path == "" {
			path = fmt.Sprintf("bank_%d_%s", a.Headers[b].FileID, aid)
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
				if err := r.ReadFull(data); err != nil {
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
			r.ReadFullUnsafe(data)

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
	r.ReadFullUnsafe(data)
	f.Close()

	ir := wio.NewInPlaceReader(data, wio.ByteOrder)
	parser.ParseBankInPlace(ir, h.DataOffset + uint64(h.DataSize))
	w.Done()
}

func ExportAllSoundbank(ctx context.Context, data string, dest string) error {
	stat, err := os.Lstat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(dest, 0777); err != nil {
				return err
			}
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is a file", dest)
	}

	f, err := os.Open(data)
	if err != nil {
		return err
	}

	timeout, cancel := context.WithTimeout(ctx, time.Second * 8)
	defer cancel()
	var w sync.WaitGroup

	sem := make(chan struct{}, MaxArchiveReder)
	for {
		entries, err := f.ReadDir(1024)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() { continue }

			ext := filepath.Ext(entry.Name())
			if strings.Compare(ext, ".stream") == 0 { continue }
			if strings.Compare(ext, ".gpu_resources") == 0 { continue }
			if strings.Compare(ext, ".ini") == 0 { continue }
			if strings.Compare(ext, ".data") == 0 { continue }
			if strings.Contains(ext, "patch") { continue }

			select {
			case <- timeout.Done():
				return timeout.Err()
			case sem <- struct{}{}:
				w.Add(1)
				go exportSoundbanks(&w, ctx, filepath.Join(data, entry.Name()), dest)
			default:
				exportSoundbanks(nil, ctx, filepath.Join(data, entry.Name()), dest)
			}
		}
	}
	w.Wait()
	return nil
}

func exportSoundbanks(
	w *sync.WaitGroup,
	ctx context.Context,
	p string,
	dest string,
) {
	if w != nil {
		defer w.Done()
	}

	f, err := os.Open(p)
	if err != nil {
		slog.Error("Failed to open archive", "path", p)
		panic(err)
	}
	defer f.Close()

	a := parser.Archive{}
	r := wio.NewReader(f, wio.ByteOrder)

	parseHeader(&a, r)

	var ww sync.WaitGroup
	sem := make(chan struct{}, MaxBankWriter)

	var bh *parser.AssetHeader
	var wh *parser.AssetHeader
	for i, b := range a.SoundBnks {
		bh = &a.Headers[b]

		var path string = fmt.Sprintf("Unknown %d.bnk", i)

		for _, w := range a.Deps {
			wh = &a.Headers[w]

			if wh.FileID == bh.FileID {
				r.AbsSeekUnsafe(uint(wh.DataOffset))

				data := make([]byte, wh.DataSize, wh.DataSize)
				if err := r.ReadFull(data); err != nil {
					slog.Error(
						"Failed to read data of wwise dependency",
						"path", p,
						"fid", wh.FileID,
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
				path = strings.ReplaceAll(path, "content_audio_", "")
				path += ".bnk"

				break
			}
		}


		sf, err := os.Open(p)
		if err != nil {
			slog.Error("Failed to open archive", "path", p)
			panic(err)
		}

		sr := wio.NewReader(
			io.NewSectionReader(
				sf,
				int64(bh.DataOffset + 16),
				int64(bh.DataSize - 16),
			),
			wio.ByteOrder,
		)

		select {
		case <- ctx.Done():
			return
		case sem <- struct{}{}:
			ww.Add(1)
			go exportSoundbank(
				&ww, ctx, 
				sf, sr, 
				filepath.Base(p), bh.FileID, filepath.Join(dest, path),
			)
		default:
			exportSoundbank(
				nil, ctx,
				sf, sr,
				filepath.Base(p), bh.FileID, filepath.Join(dest, path),
			)
		}
	}
	ww.Wait()
}

func exportSoundbank(
	w *sync.WaitGroup,
	ctx context.Context,
	i *os.File,
	r *wio.Reader,
	aid string,
	fid uint64,
	p string,
) {
	if w != nil { defer w.Done() }
	if i != nil { defer i.Close() }

	f, err := os.Create(p)
	if err != nil {
		slog.Error(
			"Failed to create file",
			"path", p,
			"aid", aid,
			"fid", fid,
			"error", err,
		)
		return
	}
	writer := bufio.NewWriterSize(f, 4096)
	defer f.Close()

	// BKHD
	data := make([]byte, 12, 12)
	_, err = r.Read(data)
	if err != nil {
		slog.Error(
			"Failed to read BKHD",
			"path", p,
			"aid", aid,
			"fid", fid,
			"error", err,
		)
		return
	}
	data[0x08] = 0x8D
	data[0x09] = 0x00
	data[0x0A] = 0x00
	data[0x0B] = 0x00
	_, err = writer.Write(data)
	if err != nil {
		slog.Error(
			"Failed to write all bytes",
			"path", p,
			"aid", aid, 
			"fid", fid,
			"error", err,
		)
		return
	}
	if err := writer.Flush(); err != nil {
		slog.Error(
			"Failed to flush bytes",
			"path", p,
			"aid", aid, 
			"fid", fid,
			"error", err,
		)
		return
	}

	data = slices.Grow(data, 4096 - 12)
	for i := range cap(data) {
		if i < len(data) {
			data[i] = 0
		} else {
			data = append(data, 0)
		}
	}
	for {
		select {
		case <- ctx.Done():
			return
		default:
		}
		nread, err := r.Read(data)
		if nread < len(data) {
			fmt.Print()
		}
		if err != nil {
			if err == io.EOF {
				if nread > 0 {
					_, err = writer.Write(data[:nread:nread])
					slog.Error(
						"Failed to write all bytes",
						"path", p,
						"aid", aid,
						"fid", fid,
						"error", err,
					)
				}
				return
			}
			slog.Error(
				"Failed to write all bytes",
				"path", p,
				"aid", aid,
				"fid", fid,
				"error", err,
			)
			return
		}
		_, err = f.Write(data[:nread:nread])
		if err != nil {
			slog.Error(
				"Failed to write all bytes",
				"path", p,
				"aid", aid,
				"fid", fid,
				"error", err,
			)
		}
	}
}

type Result struct {
	sels []byte
	err    error
}

func ExportSoundbanksTUI(ctx context.Context, data string, dest string) error {
	stat, err := os.Lstat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(dest, 0777); err != nil {
				return err
			}
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is a file", dest)
	}

	fzf := "fzf"
	if runtime.GOOS == "Windows" {
		fzf = "fzf.exe"
	}

	cmd := exec.CommandContext(ctx, fzf, "-m")
	p, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer cmd.Cancel()
	defer p.Close()

	c := make(chan *Result, 1) 
	go func() {
		result, err := cmd.Output()
		c <- &Result{result, err}
	}()

	f, err := os.Open(data)
	if err != nil {
		return err
	}

	var r *Result
	for r == nil {
		r = poll(ctx, c)
		if r != nil {
			break
		}

		entries, err := f.ReadDir(1024)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		r = writeEntryToFzF(ctx, c, cmd, entries, p)
	}

	if r == nil {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case r = <- c:
		}
	}
	if r.err != nil {
		return r.err
	}

	sem := make(chan struct{}, MaxArchiveReder)
	var w sync.WaitGroup
	sels := strings.Split(string(r.sels), "\n")
	for _, sel := range sels {
		sel = strings.Trim(sel, "\n")
		if sel == "" {
			continue
		}
		select {
		case sem <- struct{}{}:
			w.Add(1)
			go exportSoundbanks(&w, ctx, filepath.Join(data, sel), dest)
		default:
			exportSoundbanks(nil, ctx, filepath.Join(data, sel), dest)
		}
	}
	w.Wait()

	return nil
}

func poll(ctx context.Context, c chan *Result) *Result {
	select {
	case <- ctx.Done():
		return &Result{[]byte{}, ctx.Err()}
	case r := <- c:
		return r
	default:
		return nil
	}
}

func writeEntryToFzF(
	ctx context.Context,
	c chan *Result,
	cmd *exec.Cmd,
	entries []os.DirEntry,
	p io.WriteCloser,
) *Result {
	for i, entry := range entries {
		select {
		case <- ctx.Done():
			return &Result{[]byte{}, ctx.Err()}
		case r := <- c:
			return r
		default:
			ext := filepath.Ext(entry.Name())
			if strings.Compare(ext, ".stream") == 0 { continue }
			if strings.Compare(ext, ".gpu_resources") == 0 { continue }
			if strings.Compare(ext, ".ini") == 0 { continue }
			if strings.Compare(ext, ".data") == 0 { continue }
			if strings.Contains(ext, "patch") { continue }

			out := entry.Name()
			if i < len(entries) - 1 {
				out += "\n"
			}

			if _, err := p.Write([]byte(out)); err != nil {
				cmd.Cancel()
				return &Result{[]byte{}, err}
			}
		}
	}
	return nil
}

func ExportSoundbanksTUIDB(ctx context.Context, data string, dest string) error {
	stat, err := os.Lstat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(dest, 0777); err != nil {
				return err
			}
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is a file", dest)
	}

	co, err := conn()
	if err != nil {
		return err
	}
	defer co.Close()
	q := database.New(co)
	bnks, err := q.GetAllSoundbank(ctx)
	if err != nil {
		return err
	}
	if len(bnks) <= 0 {
		slog.Warn("No records of sound bank available")
		return nil
	}
	co.Close()

	fzf := "fzf"
	if runtime.GOOS == "Windows" {
		fzf = "fzf.exe"
	}

	cmd := exec.CommandContext(ctx, fzf, "-m")
	p, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer cmd.Cancel()
	defer p.Close()

	c := make(chan *Result, 1) 
	go func() {
		result, err := cmd.Output()
		c <- &Result{result, err}
	}()

	var r *Result
	for i, bnk := range bnks {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case r = <- c:
			break
		default:
			opt := fmt.Sprintf("%s | %s", bnk.Aid, bnk.Path)
			if i < len(bnks) - 1 {
				opt += "\n"
			}
			if _, err := p.Write([]byte(opt)); err != nil {
				return err
			}
		}
	}

	if r == nil {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case r = <- c:
		}
	}
	if r.err != nil {
		return r.err
	}
	cmd.Cancel()
	p.Close()

	sem := make(chan struct{}, MaxArchiveReder)
	var w sync.WaitGroup
	marks := []string{}
	sels := strings.Split(string(r.sels), "\n")
	for _, sel := range sels {
		sel = strings.Trim(sel, "\n")
		if sel == "" {
			continue
		}
		splits := strings.Split(sel, " | ")
		if slices.ContainsFunc(marks, func(mark string) bool {
			return strings.Compare(mark, splits[0]) == 0
		}) {
			continue
		}
		marks = append(marks, splits[0])
		
		select {
		case sem <- struct{}{}:
			w.Add(1)
			go exportSoundbanks(&w, ctx, filepath.Join(data, splits[0]), dest)
		default:
			exportSoundbanks(nil, ctx, filepath.Join(data, splits[0]), dest)
		}
	}
	w.Wait()

	return nil
}
