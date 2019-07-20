package extremofile

import (
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDirPrefix = "test-extremofile-delete-this-"

type tenv struct {
	t      testing.TB
	tmpdir string
	e      *efile
	w      io.Writer
}

func newTest(t testing.TB) *tenv {
	tmpdir, err := ioutil.TempDir("", testDirPrefix)
	require.NoError(t, err)
	data, e, err := Open(tmpdir)
	require.NoError(t, err)
	require.NotNil(t, e)
	require.Nil(t, data)
	return &tenv{t: t, tmpdir: tmpdir, e: e.(*efile), w: e}
}

func TestCreateDir(t *testing.T) {
	t.Parallel()

	tmpdir, err := ioutil.TempDir("", testDirPrefix)
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	dir := filepath.Join(tmpdir, "sub", "dir")
	data, e, err := Open(dir)
	require.NoError(t, err)
	require.NotNil(t, e)
	assert.Nil(t, data)

	data = []byte("test data")
	_, werr := e.Write(data)
	require.NoError(t, werr)

	rdata, _, rerr := Open(dir)
	require.NoError(t, rerr)
	assert.Equal(t, data, rdata)
}

func TestOpenEmpty(t *testing.T) {
	t.Parallel()

	env := newTest(t)
	defer os.RemoveAll(env.tmpdir)
}

func TestWriteReadSame(t *testing.T) {
	t.Parallel()

	env := newTest(t)
	defer os.RemoveAll(env.tmpdir)

	data := []byte("test data")
	_, werr := env.e.Write(data)
	require.NoError(t, werr)

	rdata, _, rerr := Open(env.tmpdir)
	require.NoError(t, rerr)
	assert.Equal(t, data, rdata)
}

func TestMainCorruptBackupOK(t *testing.T) {
	t.Parallel()

	env := newTest(t)
	defer os.RemoveAll(env.tmpdir)

	data := []byte("test data")
	_, werr := env.e.Write(data)
	require.NoError(t, werr)

	corrupt(t, env.e.pathMain, 1)

	rdata, _, rerr := Open(env.tmpdir)
	assertErrorCorrupt(t, rerr, true)
	assertErrorCritical(t, rerr, false)
	assert.Equal(t, data, rdata)
}

func TestMainCorruptMissingBackup(t *testing.T) {
	t.Parallel()

	env := newTest(t)
	defer os.RemoveAll(env.tmpdir)

	data := []byte("test data")
	_, werr := env.e.Write(data)
	require.NoError(t, werr)

	corrupt(t, env.e.pathMain, 1)
	os.Remove(env.e.pathBackup)

	rdata, _, rerr := Open(env.tmpdir)
	assert.Nil(t, rdata)
	assert.Error(t, rerr)
	assertErrorCorrupt(t, rerr, true)
	assertErrorCritical(t, rerr, true)
}

func TestMainMissingBackupCorrupt(t *testing.T) {
	t.Parallel()

	env := newTest(t)
	defer os.RemoveAll(env.tmpdir)

	data := []byte("test data")
	_, werr := env.e.Write(data)
	require.NoError(t, werr)

	os.Remove(env.e.pathMain)
	corrupt(t, env.e.pathBackup, 1)

	rdata, _, rerr := Open(env.tmpdir)
	assert.Nil(t, rdata)
	assert.Error(t, rerr)
	assertErrorCorrupt(t, rerr, true)
	assertErrorCritical(t, rerr, true)
}

func TestBothCorrupt(t *testing.T) {
	t.Parallel()

	env := newTest(t)
	defer os.RemoveAll(env.tmpdir)

	data := []byte("test data")
	_, werr := env.e.Write(data)
	require.NoError(t, werr)

	corrupt(t, env.e.pathMain, 1)
	corrupt(t, env.e.pathBackup, 1)

	rdata, _, rerr := Open(env.tmpdir)
	assert.Nil(t, rdata)
	assertErrorCorrupt(t, rerr, true)
	assertErrorCritical(t, rerr, true)
}

func assertErrorCorrupt(t testing.TB, e error, expect bool) {
	assert.Error(t, e)
	assert.Equal(t, expect, IsCorrupt(e), "IsCorrupt type=%#v err=%v", e, e)
}
func assertErrorCritical(t testing.TB, e error, expect bool) {
	assert.Error(t, e)
	assert.Equal(t, expect, IsCritical(e), "IsCritical type=%#v err=%v", e, e)
}

func corrupt(t testing.TB, path string, num uint32) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	randbs := make([]byte, num)
	rnd.Read(randbs)

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = f.WriteAt(randbs, 0)
	require.NoError(t, err)
	f.Close()
}
