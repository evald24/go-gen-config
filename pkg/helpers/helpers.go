package helpers

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

func ReadEnvAndSet(cfg interface{}) error {
	return readEnvAndSet(reflect.ValueOf(cfg))
}

func readEnvAndSet(v reflect.Value) error {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			if err := readEnvAndSet(v.Field(i)); err != nil {
				return err
			}
		} else if tag := field.Tag.Get("env"); tag != "" {
			if value := os.Getenv(tag); value != "" {
				if err := setValue(v.Field(i), value); err != nil {
					return fmt.Errorf("Failed to set environment value for \"%s\"", field.Name)
				}
			}
		}
	}

	return nil
}

func setValue(field reflect.Value, value string) error {
	valueType := field.Type()
	switch valueType.Kind() {
	// set string value
	case reflect.String:
		field.SetString(value)

	// set boolean value
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	// set integer (or time) value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Kind() == reflect.Int64 && valueType.PkgPath() == "time" && valueType.Name() == "Duration" {
			// try to parse time
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
		} else {
			// parse regular integer
			number, err := strconv.ParseInt(value, 0, valueType.Bits())
			if err != nil {
				return err
			}
			field.SetInt(number)
		}

	// set unsigned integer value
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, err := strconv.ParseUint(value, 0, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetUint(number)

	// set floating point value
	case reflect.Float32, reflect.Float64:
		number, err := strconv.ParseFloat(value, valueType.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(number)

	// unsupported types
	case reflect.Map, reflect.Ptr,
		reflect.Complex64, reflect.Interface,
		reflect.Invalid, reflect.Slice, reflect.Func,
		reflect.Array, reflect.Chan, reflect.Complex128,
		reflect.Struct, reflect.Uintptr, reflect.UnsafePointer:
	default:
		return fmt.Errorf("unsupported type: %v", valueType.Kind())
	}

	return nil
}
