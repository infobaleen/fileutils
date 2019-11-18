package fileutils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/infobaleen/errors"
)

func ReadJsonFile(filename string, p interface{}) error {
	var file, err = os.Open(filename)
	if err != nil {
		return err
	}
	err = json.NewDecoder(file).Decode(p)
	if err != nil {
		return errors.WithAftermath(err, file.Close())
	}
	return file.Close()
}

func ReadJsonFileAppend(filename string, p interface{}) error {
	var value = recursiveIndirect(toValue(p))
	if !value.CanAddr() || value.Kind() != reflect.Slice {
		return fmt.Errorf("value is not an addressable slice")
	}
	value.Addr()
	var file, err = os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	var dec = json.NewDecoder(file)
	var single = reflect.New(value.Type().Elem())
	for {
		single.Set(reflect.Zero(value.Type().Elem()))
		err = dec.Decode(single.Interface())
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		value.Set(reflect.Append(value, single.Elem()))
	}
}

func WriteJsonFile(filename string, v ...interface{}) error {
	var file, err = CreateFileTmp(filename)
	defer file.RemoveIfTmp()
	if err != nil {
		return err
	}
	err = WriteJson(file, v...)
	if err != nil {
		return err
	}
	return file.Close()
}

func WriteJson(w io.Writer, v ...interface{}) error {
	for i := range v {
		var err = writeJson(w, toValue(v[i]))
		if err != nil {
			return err
		}
	}
	return nil
}

func writeJson(w io.Writer, value reflect.Value) error {
	var enc = json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(getInterface(value))
}
