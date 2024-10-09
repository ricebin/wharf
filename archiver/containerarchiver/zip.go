package containerarchiver

import (
	"io"
	"os"
	"time"

	"github.com/itchio/arkive/zip"

	"github.com/itchio/headway/counter"
	"github.com/itchio/headway/state"

	"github.com/ricebin/wharf/archiver"

	"github.com/itchio/lake"
	"github.com/itchio/lake/tlc"

	"github.com/pkg/errors"
)

func CompressZip(archiveWriter io.Writer, container *tlc.Container, pool lake.Pool, consumer *state.Consumer) (*archiver.CompressResult, error) {
	var err error
	var uncompressedSize int64
	var compressedSize int64

	archiveCounter := counter.NewWriter(archiveWriter)

	zipWriter := zip.NewWriter(archiveCounter)
	defer zipWriter.Close()
	defer func() {
		if zipWriter != nil {
			if zErr := zipWriter.Close(); err == nil && zErr != nil {
				err = errors.WithStack(zErr)
			}
		}
	}()

	for _, dir := range container.Dirs {
		fh := zip.FileHeader{
			Name: dir.Path + "/",
		}
		fh.SetMode(os.FileMode(dir.Mode))
		fh.SetModTime(time.Now())

		_, hErr := zipWriter.CreateHeader(&fh)
		if hErr != nil {
			return nil, errors.WithStack(hErr)
		}
	}

	for fileIndex, file := range container.Files {
		fh := zip.FileHeader{
			Name:               file.Path,
			UncompressedSize64: uint64(file.Size),
			Method:             zip.Deflate,
		}
		fh.SetMode(os.FileMode(file.Mode))
		fh.SetModTime(time.Now())

		entryWriter, eErr := zipWriter.CreateHeader(&fh)
		if eErr != nil {
			return nil, errors.WithStack(eErr)
		}

		entryReader, eErr := pool.GetReader(int64(fileIndex))
		if eErr != nil {
			return nil, errors.WithStack(eErr)
		}

		copiedBytes, eErr := io.Copy(entryWriter, entryReader)
		if eErr != nil {
			return nil, errors.WithStack(eErr)
		}

		uncompressedSize += copiedBytes
	}

	for _, symlink := range container.Symlinks {
		fh := zip.FileHeader{
			Name: symlink.Path,
		}
		fh.SetMode(os.FileMode(symlink.Mode))

		entryWriter, eErr := zipWriter.CreateHeader(&fh)
		if eErr != nil {
			return nil, errors.WithStack(eErr)
		}

		entryWriter.Write([]byte(symlink.Dest))
	}

	err = zipWriter.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	zipWriter = nil

	compressedSize = archiveCounter.Count()

	return &archiver.CompressResult{
		UncompressedSize: uncompressedSize,
		CompressedSize:   compressedSize,
	}, nil
}
