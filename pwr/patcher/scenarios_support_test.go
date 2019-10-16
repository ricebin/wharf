package patcher_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/itchio/lake/tlc"
	"github.com/itchio/randsource"
	"github.com/itchio/screw"
	"github.com/itchio/wharf/pwr"
	"github.com/itchio/wharf/pwr/drip"
	"github.com/itchio/wharf/wtest"
	"github.com/stretchr/testify/assert"
)

type patchScenario struct {
	name         string
	v1           testDirSettings
	intermediate *testDirSettings
	corruptions  *testCorruption
	v2           testDirSettings
}

type testCorruption struct {
	before func(t *testing.T, dir string)
	files  testDirSettings
	after  func(t *testing.T, dir string)
}

const largeAmount int64 = 16

var testSymlinks = (runtime.GOOS != "windows")

type testDirEntry struct {
	path   string
	mode   int
	size   int64
	seed   int64
	dir    bool
	dest   string
	chunks []testDirChunk
	bsmods []bsmod
	data   []byte
}

type bsmod struct {
	// corrupt one byte every `interval`
	interval int64

	// how much to add to the byte being corrupted
	delta byte

	// only corrupt `max` times at a time, then skip `skip*interval` bytes
	max  int
	skip int
}

type testDirChunk struct {
	seed int64
	size int64
}

type testDirSettings struct {
	seed    int64
	entries []testDirEntry
}

type nopCloserWriter struct {
	writer io.Writer
}

var _ io.Writer = (*nopCloserWriter)(nil)

func (ncw *nopCloserWriter) Write(buf []byte) (int, error) {
	return ncw.writer.Write(buf)
}

func applyCorruptions(t *testing.T, dir string, c testCorruption) {
	dump := func() {
		container, err := tlc.WalkAny(dir, &tlc.WalkOpts{})
		wtest.Must(t, err)
		container.Print(func(line string) {
			t.Logf("%s", line)
		})
	}

	t.Logf("=================================")
	t.Logf("Before corruptions:")
	dump()

	if c.before != nil {
		c.before(t, dir)
	}
	makeTestDir(t, dir, c.files)
	if c.after != nil {
		c.after(t, dir)
	}

	t.Logf("---------------------------------")
	t.Logf("After corruptions:")
	dump()
	t.Logf("=================================")
}

func makeTestDir(t *testing.T, dir string, s testDirSettings) {
	prng := randsource.Reader{
		Source: rand.New(rand.NewSource(s.seed)),
	}

	assert.NoError(t, screw.MkdirAll(dir, 0o755))
	data := new(bytes.Buffer)

	for _, entry := range s.entries {
		path := filepath.Join(dir, filepath.FromSlash(entry.path))

		if entry.dir {
			mode := 0o755
			if entry.mode != 0 {
				mode = entry.mode
			}
			assert.NoError(t, screw.MkdirAll(entry.path, os.FileMode(mode)))
			continue
		} else if entry.dest != "" {
			assert.NoError(t, screw.Symlink(entry.dest, path))
			continue
		}

		parent := filepath.Dir(path)
		mkErr := screw.MkdirAll(parent, 0o755)
		if mkErr != nil {
			if !os.IsExist(mkErr) {
				assert.NoError(t, mkErr)
			}
		}

		if entry.seed == 0 {
			prng.Seed(s.seed)
		} else {
			prng.Seed(entry.seed)
		}

		data.Reset()
		data.Grow(int(entry.size))

		func() {
			mode := 0644
			if entry.mode != 0 {
				mode = entry.mode
			}

			size := pwr.BlockSize*8 + 64
			if entry.size != 0 {
				size = entry.size
			}

			f, fErr := screw.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(mode))
			assert.NoError(t, fErr)
			defer f.Close()

			if entry.data != nil {
				_, fErr = f.Write(entry.data)
				assert.NoError(t, fErr)
			} else if len(entry.chunks) > 0 {
				for _, chunk := range entry.chunks {
					prng.Seed(chunk.seed)
					data.Reset()
					data.Grow(int(chunk.size))

					_, fErr = io.CopyN(f, prng, chunk.size)
					assert.NoError(t, fErr)
				}
			} else if len(entry.bsmods) > 0 {
				func() {
					var writer io.Writer = &nopCloserWriter{f}
					for _, bsmod := range entry.bsmods {
						modcount := 0
						skipcount := 0

						drip := &drip.Writer{
							Buffer: make([]byte, bsmod.interval),
							Writer: writer,
							Validate: func(data []byte) error {
								if bsmod.max > 0 && modcount >= bsmod.max {
									skipcount = bsmod.skip
									modcount = 0
								}

								if skipcount > 0 {
									skipcount--
									return nil
								}

								data[0] = data[0] + bsmod.delta
								modcount++
								return nil
							},
						}
						defer drip.Close()
						writer = drip
					}

					_, fErr = io.CopyN(writer, prng, size)
					assert.NoError(t, fErr)
				}()
			} else {
				_, fErr = io.CopyN(f, prng, size)
				assert.NoError(t, fErr)
			}
		}()
	}
}

func cpFile(t *testing.T, src string, dst string) {
	sf, fErr := screw.Open(src)
	assert.NoError(t, fErr)
	defer sf.Close()

	info, fErr := sf.Stat()
	assert.NoError(t, fErr)

	df, fErr := screw.OpenFile(dst, os.O_CREATE|os.O_WRONLY, info.Mode())
	assert.NoError(t, fErr)
	defer df.Close()

	_, fErr = io.Copy(df, sf)
	assert.NoError(t, fErr)
}

func wipeAndMkDir(t *testing.T, dst string) {
	wtest.Must(t, screw.RemoveAll(dst))
	wtest.Must(t, screw.MkdirAll(dst, 0o755))
}

func wipeAndCpDir(t *testing.T, src string, dst string) {
	wtest.Must(t, screw.RemoveAll(dst))
	wtest.Must(t, screw.MkdirAll(dst, 0o755))
	cpDir(t, src, dst)
}

func cpDir(t *testing.T, src string, dst string) {
	assert.NoError(t, screw.MkdirAll(dst, 0o755))

	assert.NoError(t, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		assert.NoError(t, err)
		name, fErr := filepath.Rel(src, path)
		assert.NoError(t, fErr)

		dstPath := filepath.Join(dst, name)

		if info.IsDir() {
			assert.NoError(t, screw.MkdirAll(dstPath, info.Mode()))
		} else if info.Mode()&os.ModeSymlink > 0 {
			dest, fErr := screw.Readlink(path)
			assert.NoError(t, fErr)

			assert.NoError(t, screw.Symlink(dest, dstPath))
		} else if info.Mode().IsRegular() {
			df, fErr := screw.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, info.Mode())
			assert.NoError(t, fErr)
			defer df.Close()

			sf, fErr := screw.Open(path)
			assert.NoError(t, fErr)
			defer sf.Close()

			_, fErr = io.Copy(df, sf)
			assert.NoError(t, fErr)
		} else {
			return fmt.Errorf("not regular, not symlink, not dir, what is it? %s", path)
		}

		return nil
	}))
}

func assertDirEmpty(t *testing.T, dir string) {
	files, err := ioutil.ReadDir(dir)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(files))
}
