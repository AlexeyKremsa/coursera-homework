package main

import (
	"errors"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	fmt.Println(reflect.TypeOf(out))
	v := reflect.ValueOf(data)

	vOut := reflect.ValueOf(out)

	if vOut.Kind() != reflect.Ptr {
		return errors.New("out parameter is not a poitner")
	}

	vOut = vOut.Elem()

	if v.Kind() == reflect.Slice {
		if vOut.Kind() != reflect.Slice {
			return errors.New("expected slice as out parameter")
		}
		sliceItself := reflect.ValueOf(v.Interface())
		outSlice := reflect.MakeSlice(vOut.Type(), sliceItself.Len(), sliceItself.Len())
		for i := 0; i < v.Len(); i++ {
			err := i2s(sliceItself.Index(i).Interface(), outSlice.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
		vOut.Set(outSlice)
	}

	if v.Kind() == reflect.Map {
		if len(v.MapKeys()) == 0 {
			return fmt.Errorf("empty map provided")
		}

		for _, key := range v.MapKeys() {
			fieldVal := v.MapIndex(key)
			fmt.Println("===", key.Interface(), fieldVal.Interface(), reflect.TypeOf(fieldVal.Interface()))
			fOut := vOut.FieldByName(key.String())

			switch reflect.TypeOf(fieldVal.Interface()).Kind() {
			case reflect.Map:
				outVal := reflect.New(fOut.Type()).Elem()
				err := i2s(fieldVal.Interface(), outVal.Addr().Interface())
				if err != nil {
					return err
				}
				fOut.Set(outVal)

			case reflect.Float64:
				castVal, ok := fieldVal.Interface().(float64)
				if !ok {
					return errors.New("can not cast to float64")
				}
				fmt.Println(fOut.Type().String())
				if fOut.Type().String() != "int" {
					return errors.New("can not set int value")
				}

				fOut.SetInt(int64(castVal))

			case reflect.String:
				val, ok := fieldVal.Interface().(string)
				if !ok {
					return errors.New("can not set string value")
				}
				if fOut.Type().String() != "string" {
					return errors.New("can not set string value")
				}
				fOut.SetString(val)

			case reflect.Bool:
				val, ok := fieldVal.Interface().(bool)
				if !ok {
					return errors.New("can not set bool value")
				}
				if fOut.Type().String() != "bool" {
					return errors.New("can not set string value")
				}
				fOut.SetBool(val)

			case reflect.Slice:
				sliceItself := reflect.ValueOf(fieldVal.Interface())
				outSlice := reflect.MakeSlice(fOut.Type(), sliceItself.Len(), sliceItself.Len())
				for i := 0; i < sliceItself.Len(); i++ {
					err := i2s(sliceItself.Index(i).Interface(), outSlice.Index(i).Addr().Interface())
					if err != nil {
						return err
					}
				}
				fOut.Set(outSlice)
			}
		}
	}

	return nil
}
