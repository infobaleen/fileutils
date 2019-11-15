package fileutils

import (
	"github.com/matryer/is"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestBufferedIo(t *testing.T) {
	var is = is.New(t)
	var f, err = CreateFileTmp("fileutils")
	is.NoErr(err)
	defer f.Remove()

	_, err = f.Write([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	is.NoErr(err)
	_, err = f.Seek(4, 0)
	is.NoErr(err)
	var oneByte [1]byte
	_, err = f.Read(oneByte[:])
	is.NoErr(err)
	is.Equal(oneByte[0], byte(4))
	_, err = f.Write([]byte{5})
	is.NoErr(err)
	_, err = f.Read(oneByte[:])
	is.NoErr(err)
	is.Equal(oneByte[0], byte(6))
	_, err = f.Write([]byte{7, 8, 9})
	is.NoErr(err)
	var size int64
	size, err = f.Size()
	is.NoErr(err)
	is.Equal(size, int64(10))
}

type S struct {
	A int64 `json-file:"a"`
	B int64 `binary-file:"b"`
}

var original = S{1, 2}

func TestTaggedStructDir(t *testing.T) {
	var is = is.New(t)
	var tmpDir, err = ioutil.TempDir("", "")
	is.NoErr(err)
	defer os.Remove(tmpDir)

	err = WriteTaggedStructFiles(tmpDir, original)
	is.NoErr(err)
	var check S
	is.NoErr(PopulateTaggedStruct(tmpDir, &check))
	is.Equal(original, check)
}

func TestTaggedStructTar(t *testing.T) {
	var is = is.New(t)
	var tmpDir, err = ioutil.TempDir("", "")
	is.NoErr(err)
	//defer os.Remove(tmpDir)

	var r, w = io.Pipe()
	go func() {
		var tar = NewTarWriter(w)
		is.NoErr(tar.AddTaggedStruct("", original))
		is.NoErr(tar.Close())
		is.NoErr(w.Close())
	}()
	is.NoErr(Untar(tmpDir, r))
	var check S
	is.NoErr(PopulateTaggedStruct(tmpDir, &check))
	is.Equal(original, check)
}
