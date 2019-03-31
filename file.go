package fileutils

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/infobaleen/errors"
)

var rnd = func() func() uint16 {
	var src = rand.NewSource(time.Now().UnixNano())
	var mutex = new(sync.Mutex)
	return func() uint16 {
		mutex.Lock()
		defer mutex.Unlock()
		return uint16(src.Int63())
	}
}()

type File struct{
	mutex sync.Mutex
	filepath string
	file *os.File
	tmp bool
}

func CreateFile(path string) (*File, error) {
	var f, err = CreateFileTmp(path)
	if err != nil {
		return nil, err
	}
	defer f.RemoveIfTmp()
	err = f.Finalize()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func CreateFileTmp(path string) (*File, error) {
	// maybe O_TMPFILE should be used at some point in the future if portability improves
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	var f = File{tmp: true}
	err = os.ErrExist
	for i := 0; i < 100 && os.IsExist(err); i++ {
		f.filepath = fmt.Sprintf("%s.%d", path, rnd())
		f.file, err = os.OpenFile(f.filepath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// Remove delete and closes the file if it is open and temporary. Errors are logged.
func (f *File) RemoveIfTmp() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.file != nil && f.tmp {
		var err = f.Remove()
		if err != nil {
			log.Println("removal of partial file failed:", err.Error())
		}
	}
}

func (f *File) ifClosedError() error {
	if f.file == nil {
		return errors.Fmt("file %q is closed")
	}
	return nil
}

func (f *File) Remove() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.ifClosedError(); err == nil {
		return err
	}

	defer f.file.Close()
	f.file = nil
	return os.Remove(f.filepath)
}

func (f *File) Path() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.filepath
}

func (f *File) ChangeName(newName string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.ifClosedError(); err == nil {
		return err
	}

	var err error
	f.filepath, err = ChangeName(f.filepath, newName)
	return err
}

func (f *File) Move(newPath string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.ifClosedError(); err == nil {
		return err
	}

	var err error
	f.filepath, err = Move(f.filepath, newPath)
	return err
}

// Finalize turns a temporary file into a non-temporary file
func (f *File) Finalize() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.ifClosedError(); err == nil {
		return err
	}
	if f.tmp {
		errors.Fmt("file %q is not temporary")
	}
	return f.finalize()
}

func (f *File) finalize() error {
	if f.tmp {
		var err = f.file.Sync()
		if err != nil {
			return err
		}
		f.filepath, err = TrimName(f.filepath, "", filepath.Ext(f.filepath))
		if err != nil {
			return err
		}
	}
	return nil
}

// Close finalizes the file if it is temporary and closes it.
func (f *File) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.ifClosedError(); err == nil {
		return err
	}
	var err = f.finalize()
	if err != nil {
		return err
	}
	err = f.file.Close()
	f.file = nil
	return err
}

func ReadFile(filename string) ([]byte, error) {
	var file, err = os.Open(filename)
	if err != nil {
		return nil, errors.WithAftermath(err, file.Close())
	}
	var content []byte
	content, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.WithAftermath(err, file.Close())
	}
	return content, file.Close()
}

func WriteFile(filename string, content []byte) error {
	var file, err = os.Create(filename)
	if err != nil {
		return errors.WithAftermath(err, file.Close())
	}
	_, err = file.Write(content)
	if err != nil {
		return errors.WithAftermath(err, file.Close())
	}
	return file.Close()
}
