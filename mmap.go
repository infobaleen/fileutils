// +build !windows

package fileutils

import (
	"github.com/infobaleen/errors"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"reflect"
	"runtime"
	"unsafe"
)

// Mmap sets the passed slice pointer to the contents of the file.
// It is the users responsibility to stop using the slice after the file is closed.
func (f *File) Mmap(slicePointers ...interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if err := f.emptyBuffers(); err != nil {
		return err
	}
	var unmap, err = MmapFd(f.file, slicePointers...)
	f.onClose = append(f.onClose, unmap.Close)
	return err
}

type MmapHandle struct {
	bytes []byte
	size  int
}

func (h *MmapHandle) Close() error {
	var bytes []byte
	bytes, h.bytes = h.bytes, nil
	return unix.Munmap(bytes)
}

func (h *MmapHandle) setFinalizer() {
	if h.bytes != nil {
		runtime.SetFinalizer(h, func(h *MmapHandle) {
			log.Println("GC found unclosed MmapHandle, closing...")
			var err = h.Close()
			if err != nil {
				log.Println("Error closing MmapHandle:", err)
			}
		})
	}
}

func MmapCreate(path string, size int64, slicePointers ...interface{}) (*MmapHandle, error) {
	var f, err = CreateFileTmp(path)
	if err != nil {
		return nil, err
	}
	if err = f.SetSize(size); err != nil {
		return nil, err
	}
	var h *MmapHandle
	if h, err = MmapFd(f.file, slicePointers...); err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		_ = h.Close()
		return nil, err
	}
	h.SetSlicePointer(slicePointers...)
	return h, nil
}

func Mmap(path string, slicePointers ...interface{}) (*MmapHandle, error) {
	var f, err = os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	var h *MmapHandle
	h, err = MmapFd(f, slicePointers...)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		_ = h.Close()
		return nil, err
	}
	h.SetSlicePointer(slicePointers...)
	return h, nil
}

func (h *MmapHandle) SetSlicePointer(slicePointers ...interface{}) {
	for _, slicePointer := range slicePointers {
		var v = reflect.ValueOf(slicePointer)
		var t = v.Type()
		if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Slice {
			panic("not a pointer to a slice")
		}

		var sliceHeader = (*reflect.SliceHeader)(unsafe.Pointer(v.Pointer()))
		var elementSize = t.Elem().Elem().Size()
		sliceHeader.Len = h.size / int(elementSize)
		sliceHeader.Cap = sliceHeader.Len
		sliceHeader.Data = 0
		if sliceHeader.Len > 0 {
			_ = h.bytes[0]
			sliceHeader.Data = uintptr(unsafe.Pointer(&h.bytes[0]))
		}
	}
}

func MmapFd(f *os.File, slicePointers ...interface{}) (*MmapHandle, error) {
	if err := f.Sync(); err != nil {
		return nil, err
	}
	var info, err = f.Stat()
	if err != nil {
		return nil, err
	}
	var h = new(MmapHandle)
	h.size = int(info.Size())
	if h.size > 0 {
		h.bytes, err = unix.Mmap(int(f.Fd()), 0, h.size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			return nil, errors.Wrap(err, "mmap failed")
		}
		h.setFinalizer()
	}
	h.SetSlicePointer(slicePointers...)
	return h, nil
}
