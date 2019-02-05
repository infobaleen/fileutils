package fileutils

import (
	"io/ioutil"
	"os"

	"github.com/infobaleen/errors"
)

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
