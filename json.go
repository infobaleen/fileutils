package fileutils

import (
	"encoding/json"
	"os"

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
	var file, err = os.Create(filename)
	if err != nil {
		return errors.WithAftermath(err, file.Close())
	}
	var enc = json.NewEncoder(file)
	enc.SetIndent("", "\t")
	err = enc.Encode(v)
	if err != nil {
		return errors.WithAftermath(err, file.Close())
	}
	return file.Close()
}
