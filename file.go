package fileutils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"github.com/infobaleen/errors"
	"golang.org/x/sys/unix"
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

type File struct {
	mutex    sync.Mutex
	filepath string
	file     *os.File
	tmp      bool
	readBuffer bufio.Reader
	writeBuffer bufio.Writer
}

func OpenFile(path string) (*File, error) {
	var f File
	var err error
	f.filepath, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	f.file, err = os.OpenFile(f.filepath, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	f.initBuffers()
	return &f, nil
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
	f.initBuffers()
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
		f.filepath = fmt.Sprintf("%s.tmp%d", path, rnd())
		f.file, err = os.OpenFile(f.filepath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	}
	if err != nil {
		return nil, err
	}
	f.initBuffers()
	return &f, nil
}

func (f *File) initBuffers() {
	f.readBuffer = *bufio.NewReader(f.file)
	f.writeBuffer = *bufio.NewWriter(f.file)
}

func (f *File) unreadReadBuffer() error {
	var len = f.readBuffer.Buffered()
	if len > 0 {
		var _, err = f.file.Seek(-int64(len), 1)
		if err != nil {
			return err
		}
		_,_ = f.readBuffer.Discard(len)
	}
	return nil
}

func (f *File) flushWriteBuffer() error {
	return f.writeBuffer.Flush()
}

func (f *File) emptyBuffers() error {
	if err := f.unreadReadBuffer(); err != nil {
		return err
	}
	if err := f.flushWriteBuffer(); err != nil {
		return err
	}
	return nil
}

// Mmap sets the passed slice pointer to the contents of the file.
// It is the users responsibility to stop using the slice after the file is closed.
func (f *File) Mmap(slicePointer interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return Mmap(f.file, slicePointer)
}

// Mmap is a helper function to mmap the content of a os.File
func Mmap(f *os.File, slicePointer interface{}) error {
	var v = reflect.ValueOf(slicePointer)
	var t = v.Type()
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		panic("not a pointer to a slice")
	}

	var info, err = f.Stat()
	if err != nil {
		return err
	}
	var size = int(info.Size())

	var bytes []byte
	if size > 0 {
		bytes, err = unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			return errors.Wrap(err, "mmap failed")
		}
	}

	var sliceHeader = (*reflect.SliceHeader)(unsafe.Pointer(v.Pointer()))
	var elementSize = t.Elem().Elem().Size()
	sliceHeader.Len = size / int(elementSize)
	sliceHeader.Cap = sliceHeader.Len
	sliceHeader.Data = 0
	if sliceHeader.Len > 0 {
		sliceHeader.Data = uintptr(unsafe.Pointer(&bytes[0]))
	}
	return nil
}

// Remove deletes and closes the file if it is open and temporary.
func (f *File) RemoveIfTmp() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.file == nil || !f.tmp {
		return nil
	}

	var err = os.Remove(f.filepath)
	if err != nil {
		err = errors.Fmt("removal of partial file failed: %v", err.Error())
	}
	err = errors.WithAnother(err, f.file.Close())
	return err
}

func (f *File) ifClosedError() error {
	if f.file == nil {
		return errors.Fmt("file %q is closed", f.filepath)
	}
	return nil
}

func (f *File) Read(b []byte) (int, error) {
	if err := f.ifClosedError(); err != nil {
		return 0, err
	}
	if err := f.flushWriteBuffer(); err != nil {
		return 0, err
	}
	return f.readBuffer.Read(b)
}

func (f *File) Write(b []byte) (int, error) {
	if err := f.ifClosedError(); err != nil {
		return 0, err
	}
	if err := f.unreadReadBuffer(); err != nil {
		return 0, err
	}
	return f.writeBuffer.Write(b)
}

func (f *File) Size() (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.emptyBuffers(); err != nil {
		return 0, err
	}
	var info, err = f.file.Stat()
	return info.Size(), err
}

// Seek sets the byte offset of the next read or write. The whence argument controls how the offset is interpreted:
// 0 means relative to the beginning, 1 means relative to the current offset, 2 means relative to the end.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.emptyBuffers(); err != nil {
		return 0, err
	}
	if err := f.ifClosedError(); err != nil {
		return 0, err
	}
	return f.file.Seek(offset, whence)
}

func (f *File) Remove() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.ifClosedError(); err != nil {
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
	if err := f.ifClosedError(); err != nil {
		return err
	}

	var err error
	f.filepath, err = ChangeName(f.filepath, newName)
	return err
}

func (f *File) Move(newPath string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()


	if err := f.ifClosedError(); err != nil {
		return err
	} else if f.tmp {
		errors.Fmt("can't move temporary file")
	}

	var err error
	f.filepath, err = Move(f.filepath, newPath)
	return err
}

// Finalize turns a temporary file into a non-temporary file
func (f *File) Finalize() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.ifClosedError(); err != nil {
		return err
	}
	if !f.tmp {
		errors.Fmt("file %q is not temporary", f.filepath)
	}
	return f.finalize()
}

func (f *File) Sync() error {
	if err := f.emptyBuffers(); err != nil {
		return err
	}
	if err := f.file.Sync(); err != nil {
		return err
	}
	return nil
}

func (f *File) finalize() error {
	if f.tmp {
		var err = f.Sync()
		if err != nil {
			return err
		}
		f.filepath, err = TrimName(f.filepath, "", filepath.Ext(f.filepath))
		if err != nil {
			return err
		}
		f.tmp = false
	}
	return nil
}

// Close finalizes the file if it is temporary and closes it.
func (f *File) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.ifClosedError(); err != nil {
		return err
	}
	if err := f.finalize(); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	var err = f.file.Close()
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
