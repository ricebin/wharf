package pwr

import (
	"io/ioutil"
	"testing"

	"github.com/ricebin/wharf/wtest"
)

func Test_SafeKeeper(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	wtest.Must(t, err)

	wtest.MakeTestDir(t, dir, wtest.TestDirSettings{})
}
