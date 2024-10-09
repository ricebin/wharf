package wtest

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/itchio/randsource"
	"github.com/itchio/screw"
	"github.com/ricebin/wharf/pwr/drip"
	"github.com/stretchr/testify/assert"
)

// see pwr/constants - the same BlockSize is used here to
// generate test data that diffs in a certain way
const BlockSize int64 = 64 * 1024 // 64k

var TestSymlinks = (runtime.GOOS != "windows")

type TestDirEntry struct {
	Path      string
	Mode      int
	Size      int64
	Seed      int64
	Dir       bool
	Dest      string
	Chunks    []TestDirChunk
	Bsmods    []Bsmod
	Swaperoos []Swaperoo
	Data      []byte
}

// Swaperoo swaps two blocks of the file
type Swaperoo struct {
	OldStart int64
	NewStart int64
	Size     int64
}

// Bsmode represents a bsdiff-like corruption
type Bsmod struct {
	// corrupt one byte every `interval`
	Interval int64

	// how much to add to the byte being corrupted
	Delta byte

	// only corrupt `max` times at a time, then skip `skip*interval` bytes
	Max  int
	Skip int
}

type TestDirChunk struct {
	Seed int64
	Size int64
}

type TestDirSettings struct {
	Seed    int64
	Entries []TestDirEntry
}

func MakeTestDir(t *testing.T, dir string, s TestDirSettings) {
	prng := randsource.Reader{
		Source: rand.New(rand.NewSource(s.Seed)),
	}

	Must(t, screw.MkdirAll(dir, 0o755))
	data := new(bytes.Buffer)

	for _, entry := range s.Entries {
		path := filepath.Join(dir, filepath.FromSlash(entry.Path))

		if entry.Dir {
			mode := 0o755
			if entry.Mode != 0 {
				mode = entry.Mode
			}
			Must(t, screw.MkdirAll(entry.Path, os.FileMode(mode)))
			continue
		} else if entry.Dest != "" {
			Must(t, screw.Symlink(entry.Dest, path))
			continue
		}

		parent := filepath.Dir(path)
		mkErr := screw.MkdirAll(parent, 0o755)
		if mkErr != nil {
			if !os.IsExist(mkErr) {
				Must(t, mkErr)
			}
		}

		if entry.Seed == 0 {
			prng.Seed(s.Seed)
		} else {
			prng.Seed(entry.Seed)
		}

		func() {
			mode := 0o644
			if entry.Mode != 0 {
				mode = entry.Mode
			}

			size := BlockSize*8 + 64
			if entry.Size > 0 {
				size = entry.Size
			} else if entry.Size < 0 {
				size = 0
			}

			data.Reset()
			data.Grow(int(size))

			f := new(bytes.Buffer)
			var err error

			if entry.Data != nil {
				_, err = f.Write(entry.Data)
				Must(t, err)
			} else if len(entry.Chunks) > 0 {
				for _, chunk := range entry.Chunks {
					prng.Seed(chunk.Seed)
					data.Reset()
					data.Grow(int(chunk.Size))

					_, err = io.CopyN(f, prng, chunk.Size)
					Must(t, err)
				}
			} else if len(entry.Bsmods) > 0 {
				func() {
					var writer io.Writer = NopWriteCloser(f)
					for _, bsmod := range entry.Bsmods {
						modcount := 0
						skipcount := 0

						drip := &drip.Writer{
							Buffer: make([]byte, bsmod.Interval),
							Writer: writer,
							Validate: func(data []byte) error {
								if bsmod.Max > 0 && modcount >= bsmod.Max {
									skipcount = bsmod.Skip
									modcount = 0
								}

								if skipcount > 0 {
									skipcount--
									return nil
								}

								data[0] = data[0] + bsmod.Delta
								modcount++
								return nil
							},
						}
						defer drip.Close()
						writer = drip
					}

					_, err = io.CopyN(writer, prng, size)
					Must(t, err)
				}()
			} else {
				_, err = io.CopyN(f, prng, size)
				Must(t, err)
			}

			finalBuf := f.Bytes()
			for _, s := range entry.Swaperoos {
				stagingBuf := make([]byte, s.Size)
				copy(stagingBuf, finalBuf[s.OldStart:s.OldStart+s.Size])
				copy(finalBuf[s.OldStart:s.OldStart+s.Size], finalBuf[s.NewStart:s.NewStart+s.Size])
				copy(finalBuf[s.NewStart:s.NewStart+s.Size], stagingBuf)
			}

			err = screw.WriteFile(path, finalBuf, os.FileMode(mode))
			Must(t, err)
		}()
	}
}

func WipeAndMkdir(t *testing.T, dst string) {
	Must(t, screw.RemoveAll(dst))
	Must(t, screw.MkdirAll(dst, 0o755))
}

func WipeAndCpDir(t *testing.T, src string, dst string) {
	Must(t, screw.RemoveAll(dst))
	Must(t, screw.MkdirAll(dst, 0o755))
	CpDir(t, src, dst)
}

func CpFile(t *testing.T, src string, dst string) {
	sf, fErr := screw.Open(src)
	Must(t, fErr)
	defer sf.Close()

	info, fErr := sf.Stat()
	Must(t, fErr)

	df, fErr := screw.OpenFile(dst, os.O_CREATE|os.O_WRONLY, info.Mode())
	Must(t, fErr)
	defer df.Close()

	_, fErr = io.Copy(df, sf)
	Must(t, fErr)
}

func CpDir(t *testing.T, src string, dst string) {
	Must(t, screw.MkdirAll(dst, 0o755))

	Must(t, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		Must(t, err)
		name, fErr := filepath.Rel(src, path)
		Must(t, fErr)

		dstPath := filepath.Join(dst, name)

		if info.IsDir() {
			Must(t, screw.MkdirAll(dstPath, info.Mode()))
		} else if info.Mode()&os.ModeSymlink > 0 {
			dest, fErr := screw.Readlink(path)
			Must(t, fErr)

			Must(t, screw.Symlink(dest, dstPath))
		} else if info.Mode().IsRegular() {
			df, fErr := screw.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, info.Mode())
			Must(t, fErr)
			defer df.Close()

			sf, fErr := screw.Open(path)
			Must(t, fErr)
			defer sf.Close()

			_, fErr = io.Copy(df, sf)
			Must(t, fErr)
		} else {
			return fmt.Errorf("not regular, not symlink, not dir, what is it? %s", path)
		}

		return nil
	}))
}

func AssertDirEmpty(t *testing.T, dir string) {
	files, err := screw.ReadDir(dir)
	Must(t, err)
	assert.Equal(t, 0, len(files))
}
