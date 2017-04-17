package main

import (
	"reflect"
	"io"
	"fmt"
	"bytes"
	"sync"
)

var bufferPool *sync.Pool

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
}

func GetString(val interface{}) string {
	var buffer *bytes.Buffer
	buffer = bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer bufferPool.Put(buffer)

	toString(buffer, reflect.ValueOf(val))
	return buffer.String()
}

func toString(w io.Writer, val reflect.Value) {
	if val == (reflect.Value{}) {
		w.Write([]byte("<nil>"))
		return
	}
	if val.Kind() == reflect.Ptr && val.IsNil() {
		w.Write([]byte("<nil>"))
		return
	}
	v := reflect.Indirect(val)

	switch v.Kind() {
	case reflect.String:
		String(w, v)
	case reflect.Slice:
		Slice(w, v)
	case reflect.Map:
		Map(w, v)
	case reflect.Struct:
		Struct(w, v)
	default:
		if v.CanInterface() {
			fmt.Fprint(w, v.Interface())
		}
	}
}

func String(w io.Writer, v reflect.Value) {
	fmt.Fprintf(w, `"%s"`, v)
}

func Slice(w io.Writer, v reflect.Value) {
	w.Write([]byte{'['})

	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			w.Write([]byte{' '})
		}
		toString(w, v.Index(i))
	}

	w.Write([]byte{']'})
}

func Map(w io.Writer, v reflect.Value) {
	w.Write([]byte("map["))

	for i, key := range v.MapKeys() {
		if i > 0 {
			w.Write([]byte{' '})
		}
		toString(w, key)
		w.Write([]byte{':'})
		toString(w, v.MapIndex(key))
	}

	w.Write([]byte{']'})
}

func Time(w io.Writer, v reflect.Value) {
	fmt.Fprintf(w, "{%s}", v.Interface())
}

func Struct(w io.Writer, v reflect.Value) {
	if v.Type().Name() != "" {
		w.Write([]byte(v.Type().String()))
	}
	if v.Type().String() == "time.Time" {
		Time(w, v)
		return
	}

	w.Write([]byte{'{'})

	var sep bool
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}
		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			continue
		}
		if fv.Kind() == reflect.Slice && fv.IsNil() {
			continue
		}

		if sep {
			w.Write([]byte(", "))
		} else {
			sep = true
		}

		w.Write([]byte(v.Type().Field(i).Name))
		w.Write([]byte{':'})
		toString(w, fv)
	}

	w.Write([]byte{'}'})
}
