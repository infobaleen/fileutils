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
	val = recursiveIndirect(val)
	if val.Kind() == reflect.Slice {
		var elemSize = sizeBinary(reflect.Zero(val.Type().Elem()))
		var fileInfo, err = file.Stat()
		if err != nil {
			return errors.WithTrace(err)
		}
		var len = int(fileInfo.Size() / int64(elemSize))
		if val.Cap() < len {
			val.Set(reflect.MakeSlice(val.Type(), 0, len))
		}
		val.SetLen(len)
	}
	return binary.Read(file, binary.LittleEndian, p)
}

func sizeBinary(v reflect.Value) int {
	var val = reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() == reflect.Slice {
		var total int
		var l = val.Len()
		for i := 0; i < l; i++ {
			total += sizeBinary(val.Index(i))
		}
		return total
	}
	return binary.Size(v.Interface())
}

func writeBinary(w io.Writer, v reflect.Value) error {
	v = recursiveIndirect(v)
	var val = reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() == reflect.Slice && val.Type().Elem().Kind() == reflect.Slice {
		var l = val.Len()
		for i := 0; i < l; i++ {
			var err = writeBinary(w, val.Index(i))
			if err != nil {
				return err
			}
		}
		return nil
	}
	return errors.WithTrace(binary.Write(w, binary.LittleEndian, v.Addr().Interface()))
}

func SizeBinary(v interface{}) int {
	return sizeBinary(toValue(v))
}

func WriteBinaryFile(filename string, v interface{}) error {
	var file, err = CreateFileTmp(filename)
	if err != nil {
		return err
	}
	defer file.RemoveIfTmp()
	err = writeBinary(file, toValue(v))
	if err != nil {
		return err
	}
	return file.Close()
}
