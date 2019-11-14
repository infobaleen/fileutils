package fileutils

import (
	"encoding/binary"
	"io"
	"os"
	"reflect"

	"github.com/infobaleen/errors"
)

func ReadBinaryFile(filename string, p interface{}) error {
	var val = toValue(p)
	if val.Kind() != reflect.Ptr {
		return errors.Fmt("expected pointer")
	}
	val = recursiveIndirect(val)
	var file, err = os.Open(filename)
	if err != nil {
		return errors.WithTrace(err)
	}
	defer file.Close()
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
	var val = recursiveIndirect(v)
	if val.Kind() == reflect.Slice {
		var total int
		var l = val.Len()
		for i := 0; i < l; i++ {
			total += sizeBinary(val.Index(i))
		}
		return total
	}
	return binary.Size(referenceInterface(v))
}

func writeBinary(w io.Writer, v reflect.Value) error {
	v = recursiveIndirect(v)
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Slice {
		var l = v.Len()
		for i := 0; i < l; i++ {
			var err = writeBinary(w, v.Index(i))
			if err != nil {
				return err
			}
		}
		return nil
	}
	return errors.WithTrace(binary.Write(w, binary.LittleEndian, referenceInterface(v)))
}

func SizeBinary(v interface{}) int {
	return sizeBinary(recursiveIndirect(toValue(v)))
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
