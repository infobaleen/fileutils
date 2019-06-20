package fileutils

import (
	"encoding/binary"
	"os"

	"github.com/infobaleen/errors"
)

func ReadBinaryFile(filename string, p interface{}) error {
	var file, err = os.Open(filename)
	if err != nil {
		return err
	}
	err = binary.Read(file, binary.LittleEndian, p)
	if err != nil {
		return errors.WithAftermath(err, file.Close())
	}
	return file.Close()
}

func WriteBinaryFile(filename string, v interface{}) error {
	var file, err = os.Create(filename)
	if err != nil {
		return errors.WithAftermath(err, file.Close())
	}
	err = binary.Write(file, binary.LittleEndian, v)
	if err != nil {
		return errors.WithAftermath(err, file.Close())
	}
	return file.Close()
}
