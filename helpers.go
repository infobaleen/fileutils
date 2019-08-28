package fileutils

import (
	"github.com/infobaleen/errors"
	"path"
	"reflect"
)

func PopulateStruct(dir string, model interface{}) error {
	var val = reflect.ValueOf(model)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.Fmt("expected pointer to struct")
	}
	val = val.Elem()

	for i := 0; i < val.NumField(); i++ {
		var tag = val.Type().Field(i).Tag.Get("binary-file")
		if tag != "" {
			var err = ReadBinaryFile(path.Join(dir, tag), val.Field(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
