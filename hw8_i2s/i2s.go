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
	vOut = vOut.Elem()
	//fmt.Println("KIND: ", v.Kind())
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			fieldVal := v.MapIndex(key)
			fmt.Println("===", key.Interface(), fieldVal.Interface(), reflect.TypeOf(fieldVal.Interface()))

			fOut := vOut.FieldByName(key.String())

			switch reflect.TypeOf(fieldVal.Interface()).Kind() {
			case reflect.Map:
				outVal := reflect.New(fOut.Type()).Elem()
				i2s(fieldVal.Interface(), outVal.Addr().Interface())
				fOut.Set(outVal)

			case reflect.Float64:
				castVal, ok := fieldVal.Interface().(float64)
				if !ok {
					return errors.New("can not cast to float64")
				}
				fOut.SetInt(int64(castVal))

			case reflect.String:
				fOut.SetString(fieldVal.Interface().(string))

			case reflect.Bool:
				fOut.SetBool(fieldVal.Interface().(bool))

			case reflect.Slice:
				sliceItself := reflect.ValueOf(fieldVal.Interface())
				outSlice := reflect.MakeSlice(fOut.Type(), sliceItself.Len(), sliceItself.Len())
				fmt.Println(reflect.TypeOf(outSlice), fOut.Type())
				for i := 0; i < sliceItself.Len(); i++ {
					i2s(sliceItself.Index(i).Interface(), outSlice.Index(i).Addr().Interface())
				}
				fmt.Println("RES: ", outSlice)
				fmt.Println(reflect.TypeOf(fOut))
				fOut.Set(outSlice)
			}
		}
	}

	return nil
}
