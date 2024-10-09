package pwr

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/itchio/headway/counter"
	"github.com/itchio/headway/state"
	"github.com/itchio/headway/united"

	"github.com/itchio/httpkit/eos"
	"github.com/itchio/httpkit/eos/option"

	"github.com/ricebin/wharf/ctxcopy"
	"github.com/ricebin/wharf/werrors"

	"github.com/itchio/arkive/zip"

	"github.com/itchio/lake"
	"github.com/itchio/lake/pools/fspool"
	"github.com/itchio/lake/pools/zippool"
	"github.com/itchio/lake/tlc"

	"github.com/pkg/errors"
)

// An ArchiveHealer can repair from a .zip file (remote or local)
type ArchiveHealer struct {
	// the directory we should heal
	Target string

	// an eos path for the archive
	ArchivePath string

	archiveFile    eos.File
	archiveFileErr error
	archiveLock    sync.Mutex
	archiveOnce    sync.Once

	// A consumer to report progress to
	Consumer *state.Consumer

	// internal
	progressMutex  sync.Mutex
	totalCorrupted int64
	totalHealing   int64
	totalHealed    int64
	totalHealthy   int64
	hasWounds      bool

	container *tlc.Container

	lockMap LockMap
}

var _ Healer = (*ArchiveHealer)(nil)

type chunkHealedFunc func(chunkHealed int64)

// Do starts receiving from the wounds channel and healing
func (ah *ArchiveHealer) Do(parentCtx context.Context, container *tlc.Container, wounds chan *Wound) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	ah.container = container

	files := make(map[int64]bool)
	fileIndices := make(chan int64, len(container.Files))

	targetPool := fspool.New(container, ah.Target)

	errs := make(chan error, 1)

	onChunkHealed := func(healedChunk int64) {
		ah.progressMutex.Lock()
		ah.totalHealed += healedChunk
		ah.progressMutex.Unlock()
		ah.updateProgress()
	}

	defer func() {
		if ah.archiveFile != nil {
			ah.archiveFile.Close()
		}
	}()

	go func() {
		errs <- ah.heal(ctx, container, targetPool, fileIndices, onChunkHealed)
	}()

	processWound := func(wound *Wound) error {
		if !wound.Healthy() {
			ah.totalCorrupted += wound.Size()
			ah.hasWounds = true
		}

		switch wound.Kind {
		case WoundKind_DIR:
			dirEntry := container.Dirs[wound.Index]
			path := filepath.Join(ah.Target, filepath.FromSlash(dirEntry.Path))

			stats, err := os.Lstat(path)
			if err == nil {
				if stats.IsDir() {
					ah.Consumer.Debugf("For dir wound, found existing dir (%s), all good", path)
					return nil
				} else {
					ah.Consumer.Debugf("For dir wound, found file/symlink (%s), removing", path)
					err = os.Remove(path)
					if err != nil {
						return errors.WithStack(err)
					}
				}
			}

			ah.Consumer.Debugf("For dir wound, doing MkdirAll (%s)", path)
			err = os.MkdirAll(path, 0o755)
			if err != nil {
				return errors.WithStack(err)
			}

		case WoundKind_SYMLINK:
			symlinkEntry := container.Symlinks[wound.Index]
			path := filepath.Join(ah.Target, filepath.FromSlash(symlinkEntry.Path))

			dir := filepath.Dir(path)
			err := os.MkdirAll(dir, 0o755)
			if err != nil {
				return errors.WithStack(err)
			}

			stats, err := os.Lstat(path)
			if err == nil {
				if stats.IsDir() {
					ah.Consumer.Debugf("For symlink wound, found dir (%s), doing RemoveAll", path)
					err = os.RemoveAll(path)
					if err != nil {
						return errors.WithStack(err)
					}
				} else {
					ah.Consumer.Debugf("For symlink wound, found file/symlink (%s), doing Remove", path)
					err = os.Remove(path)
					if err != nil {
						return errors.WithStack(err)
					}
				}
			}

			ah.Consumer.Debugf("For symlink wound, doing Symlink (%s) => (%s)", path, symlinkEntry.Dest)
			err = os.Symlink(symlinkEntry.Dest, path)
			if err != nil {
				return errors.WithStack(err)
			}

		case WoundKind_FILE:
			if files[wound.Index] {
				// already queued
				return nil
			}

			file := container.Files[wound.Index]
			ah.Consumer.ProgressLabel(file.Path)

			ah.progressMutex.Lock()
			ah.totalHealing += file.Size
			ah.progressMutex.Unlock()
			ah.updateProgress()
			files[wound.Index] = true

			select {
			case err := <-errs:
				return errors.WithStack(err)
			case fileIndices <- wound.Index:
				// queued for work!
			}

		case WoundKind_CLOSED_FILE:
			if files[wound.Index] {
				// already healing whole file
			} else {
				fileSize := container.Files[wound.Index].Size

				// whole file was healthy
				if wound.End == fileSize {
					ah.progressMutex.Lock()
					ah.totalHealthy += fileSize
					ah.progressMutex.Unlock()
					ah.updateProgress()
				}
			}

		default:
			return fmt.Errorf("Unknown wound kind: %d", wound.Kind)
		}

		return nil
	}

	for wound := range wounds {
		select {
		case <-ctx.Done():
			return werrors.ErrCancelled
		default:
			// keep going!
		}

		err := processWound(wound)
		if err != nil {
			return err
		}
	}

	// queued everything
	close(fileIndices)

	err := <-errs
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (ah *ArchiveHealer) openArchive() (eos.File, error) {
	ah.archiveLock.Lock()
	defer ah.archiveLock.Unlock()

	ah.archiveOnce.Do(func() {
		file, err := eos.Open(ah.ArchivePath, option.WithConsumer(ah.Consumer))
		ah.archiveFile = file
		ah.archiveFileErr = err
	})
	return ah.archiveFile, ah.archiveFileErr
}

func (ah *ArchiveHealer) heal(ctx context.Context, container *tlc.Container, targetPool lake.WritablePool,
	fileIndices chan int64, chunkHealed chunkHealedFunc) error {

	var sourcePool lake.Pool
	var err error

	for {
		select {
		case <-ctx.Done():
			// something else stopped the healing
			return nil
		case fileIndex, ok := <-fileIndices:
			if !ok {
				// no more files to heal
				return nil
			}

			// lazily open file
			if sourcePool == nil {
				file, err := ah.openArchive()
				if err != nil {
					return errors.WithStack(err)
				}

				stat, err := file.Stat()
				if err != nil {
					return err
				}

				zipReader, err := zip.NewReader(file, stat.Size())
				if err != nil {
					return errors.WithStack(err)
				}

				sourcePool = zippool.New(container, zipReader)
				// sic: we're inside a for, not a function, so this correctly happens
				// when we actually return
				defer sourcePool.Close()
			}

			err = ah.healOne(ctx, sourcePool, targetPool, fileIndex, chunkHealed)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}
}

func (ah *ArchiveHealer) healOne(ctx context.Context, sourcePool lake.Pool, targetPool lake.WritablePool, fileIndex int64, chunkHealed chunkHealedFunc) error {
	if ah.lockMap != nil {
		lock := ah.lockMap[fileIndex]
		select {
		case <-lock:
			// keep going
		case <-ctx.Done():
			return werrors.ErrCancelled
		}
	}

	var err error
	var reader io.Reader
	var writer io.WriteCloser

	f := ah.container.Files[fileIndex]

	ah.Consumer.Debugf("Healing (%s) %s", f.Path, united.FormatBytes(f.Size))

	reader, err = sourcePool.GetReader(fileIndex)
	if err != nil {
		return err
	}

	writer, err = targetPool.GetWriter(fileIndex)
	if err != nil {
		return err
	}
	defer writer.Close()

	if f, ok := writer.(*os.File); ok {
		stats, err := f.Stat()
		if err != nil {
			ah.Consumer.Debugf("Was a file but can't stat: %+v", err)
		} else {
			ah.Consumer.Debugf("Was a file: %s", united.FormatBytes(stats.Size()))
		}
	}

	lastCount := int64(0)
	cw := counter.NewWriterCallback(func(count int64) {
		chunk := count - lastCount
		chunkHealed(chunk)
		lastCount = count
	}, writer)

	_, err = ctxcopy.Do(ctx, cw, reader)
	if err != nil {
		return err
	}

	return err
}

// HasWounds returns true if the healer ever received wounds
func (ah *ArchiveHealer) HasWounds() bool {
	return ah.hasWounds
}

// TotalCorrupted returns the total amount of corrupted data
// contained in the wounds this healer has received. Dirs
// and symlink wounds have 0-size, use HasWounds to know
// if there were any wounds at all.
func (ah *ArchiveHealer) TotalCorrupted() int64 {
	return ah.totalCorrupted
}

// TotalHealed returns the total amount of data written to disk
// to repair the wounds. This might be more than TotalCorrupted,
// since ArchiveHealer always redownloads whole files, even if
// they're just partly corrupted
func (ah *ArchiveHealer) TotalHealed() int64 {
	return ah.totalHealed
}

// SetConsumer gives this healer a consumer to report progress to
func (ah *ArchiveHealer) SetConsumer(consumer *state.Consumer) {
	ah.Consumer = consumer
}

func (ah *ArchiveHealer) updateProgress() {
	if ah.Consumer == nil {
		return
	}

	ah.progressMutex.Lock()
	progress := float64(ah.totalHealthy+ah.totalHealed) / float64(ah.container.Size)
	ah.Consumer.Progress(progress)
	ah.progressMutex.Unlock()
}

func (ah *ArchiveHealer) SetLockMap(lockMap LockMap) {
	ah.lockMap = lockMap
}
