package fileutils

import (
	"github.com/infobaleen/errors"
	"path"
	"reflect"
)

const (
	TagKeyBinaryFile = "binary-file"
)

func PopulateTaggedStruct(dir string, p interface{}) error {
	var val = toValue(p)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.Fmt("expected pointer to struct")
	}
	val = val.Elem()

	for i := 0; i < val.NumField(); i++ {
		var tag = val.Type().Field(i).Tag.Get(TagKeyBinaryFile)
		if tag != "" {
			var err = ReadBinaryFile(path.Join(dir, tag), val.Field(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func IterateTaggedStruct(v interface{}, f func(fileType, fileName string, field reflect.Value) error) error {
	var val = recursiveIndirect(toValue(v))
	if val.Kind() != reflect.Struct {
		return errors.Fmt("expected struct or pointer to struct")
	}
	for i := 0; i < val.NumField(); i++ {
		var tag = val.Type().Field(i).Tag.Get(TagKeyBinaryFile)
		if tag != "" {
			var err = f(TagKeyBinaryFile, tag, val.Field(i).Addr())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func toValue(v interface{}) reflect.Value {
	var value, isValue = v.(reflect.Value)
	if !isValue {
		value = reflect.ValueOf(v)
	}
	return value
}

func recursiveIndirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
