// +build !windows

package fileutils

import (
	"github.com/infobaleen/errors"
	"golang.org/x/sys/unix"
	"os"
	"reflect"
	"unsafe"
)

// Mmap sets the passed slice pointer to the contents of the file.
// It is the users responsibility to stop using the slice after the file is closed.
func (f *File) Mmap(slicePointer interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.emptyBuffers(); err != nil {
		return err
	}
	var unmap, err = Mmap(f.file, slicePointer)
	f.onClose = append(f.onClose, unmap)
	return err
}

// Mmap is a helper function to mmap the content of a os.File
func Mmap(f *os.File, slicePointer interface{}) (func() error, error) {
	var v = reflect.ValueOf(slicePointer)
	var t = v.Type()
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
		panic("not a pointer to a slice")
	}

	if err := f.Sync(); err != nil {
		return nil, err
	}
	var info, err = f.Stat()
	if err != nil {
		return nil, err
	}
	var size = int(info.Size())

	var bytes []byte
	if size > 0 {
		bytes, err = unix.Mmap(int(f.Fd()), 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			return nil, errors.Wrap(err, "mmap failed")
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
	return func() error {
		return unix.Munmap(bytes)
	}, nil
}
