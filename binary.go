package fileutils

import (
	"encoding/binary"
	"os"
	"reflect"

	"github.com/infobaleen/errors"
)

func ReadBinaryFile(filename string, p interface{}) error {
	var file, err = os.Open(filename)
	if err != nil {
		return errors.WithTrace(err)
	}
	defer file.Close()
	var val = reflect.ValueOf(p)
	if val.Kind() != reflect.Ptr {
		return errors.Fmt("expected pointer")
	}
	val = val.Elem()
	if val.Kind() == reflect.Slice {
		var elemSize = int64(val.Type().Elem().Size())
		var fileInfo, err = file.Stat()
		if err != nil {
			return errors.WithTrace(err)
		}
		var len = int(fileInfo.Size() / elemSize)
		if val.Cap() < len {
			val.Set(reflect.MakeSlice(val.Type(), 0, len))
		}
		val.SetLen(len)
	}
	return binary.Read(file, binary.LittleEndian, p)
}

func WriteBinaryFile(filename string, v interface{}) error {
	var file, err = os.Create(filename)
	if err != nil {
		return errors.WithTrace(err)
	}
	err = binary.Write(file, binary.LittleEndian, v)
	if err != nil {
		return errors.WithTrace(err)
	}
	return errors.WithTrace(file.Close())
}
