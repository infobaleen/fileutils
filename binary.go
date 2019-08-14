package fileutils

import (
	"encoding/binary"
	"io"
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

func writeBinary(w io.Writer, v interface{}) error {
	var val = reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Slice {
		var l = val.Len()
		for i := 0; i < l; i++ {
			var err = writeBinary(w, val.Index(i).Interface())
			if err != nil  {
				return err
			}
		}
		return nil
	}
	return errors.WithTrace(binary.Write(w, binary.LittleEndian, v))
}

func WriteBinaryFile(filename string, v interface{}) error {
	var file, err = CreateFileTmp(filename)
	if err != nil {
		return err
	}
	defer file.RemoveIfTmp()
	err = writeBinary(file, v)
	if err != nil {
		return err
	}
	return file.Close()
}
