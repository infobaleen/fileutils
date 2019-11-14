package fileutils

import (
	"encoding/json"
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

func WriteJsonFile(filename string, v interface{}) error {
	var file, err = CreateFileTmp(filename)
	defer file.RemoveIfTmp()
	if err != nil {
		return err
	}
	err = WriteJson(file, v)
	if err != nil {
		return err
	}
	return file.Close()
}

func WriteJson(w io.Writer, v interface{}) error {
	return writeJson(w, toValue(v))
}

func writeJson(w io.Writer, value reflect.Value) error {
	var enc = json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(referenceInterface(value))
}
