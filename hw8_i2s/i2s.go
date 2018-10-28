package main

import (
	"errors"
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {

	v := reflect.ValueOf(data)
	vOut := reflect.ValueOf(out)
	vOut = vOut.Elem()

	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			fieldVal := v.MapIndex(key)
			fmt.Println(key.Interface(), fieldVal.Interface(), reflect.TypeOf(fieldVal.Interface()))

			fOut := vOut.FieldByName(key.String())

			switch reflect.TypeOf(fieldVal.Interface()).String() {
			case "float64":
				fmt.Println("Int value: ", fieldVal.Interface())
				castVal, ok := fieldVal.Interface().(float64)
				if !ok {
					return errors.New("can not cast to int64")
				}
				fOut.SetInt(int64(castVal))

			case "string":
				fOut.SetString(fieldVal.Interface().(string))

			case "bool":
				fOut.SetBool(fieldVal.Interface().(bool))
			}
		}
	}
	return nil
}
