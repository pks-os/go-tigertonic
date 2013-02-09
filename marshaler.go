package tigertonic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type Request interface{}

type Response interface{}

type Marshaler struct {
	f reflect.Value
}

func Marshaled(f interface{}) *Marshaler {
	t := reflect.TypeOf(f)
	if reflect.Func != t.Kind() {
		panic(newMarshalerError(fmt.Sprintf("kind was %v, not Func", t.Kind())))
	}
	if 3 != t.NumIn() {
		panic(newMarshalerError(fmt.Sprintf("input arity was %v, not 3", t.NumIn())))
	}
	if "*url.URL" != t.In(0).String() {
		panic(newMarshalerError(fmt.Sprintf("type of first argument was %v, not *url.URL", t.In(0))))
	}
	if "http.Header" != t.In(1).String() {
		panic(newMarshalerError(fmt.Sprintf("type of second argument was %v, not http.Header", t.In(1))))
	}
	if !t.In(2).Implements(reflect.TypeOf((*Request)(nil)).Elem()) {
		panic(newMarshalerError(fmt.Sprintf("type of third argument was %v, not Request", t.Out(2))))
	}
	if 4 != t.NumOut() {
		panic(newMarshalerError(fmt.Sprintf("output arity was %v, not 4", t.NumOut())))
	}
	if reflect.Int != t.Out(0).Kind() {
		panic(newMarshalerError(fmt.Sprintf("type of first return value was %v, not int", t.Out(0))))
	}
	if "http.Header" != t.Out(1).String() {
		panic(newMarshalerError(fmt.Sprintf("type of second return value was %v, not http.Header", t.Out(1))))
	}
	if !t.Out(2).Implements(reflect.TypeOf((*Response)(nil)).Elem()) {
		panic(newMarshalerError(fmt.Sprintf("type of third return value was %v, not Response", t.Out(2))))
	}
	if "error" != t.Out(3).String() {
		panic(newMarshalerError(fmt.Sprintf("type of fourth return value was %v, not error", t.Out(3))))
	}
	return &Marshaler{reflect.ValueOf(f)}
}

func (m *Marshaler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	rq := reflect.New(m.f.Type().In(2).Elem())
	if "POST" == r.Method || "PUT" == r.Method {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		decoder := reflect.ValueOf(json.NewDecoder(r.Body))
		out := decoder.MethodByName("Decode").Call([]reflect.Value{rq})
		if !out[0].IsNil() {
			err := out[0].Interface().(error)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err) // TODO JSON
			log.Println(w, err)
			return
		}
	}
	out := m.f.Call([]reflect.Value{
		reflect.ValueOf(r.URL),
		reflect.ValueOf(r.Header),
		rq,
	})
	status := int(out[0].Int())
	header := out[1].Interface().(http.Header)
	rs := out[2].Interface().(Response)
	if !out[3].IsNil() {
		if 100 > status {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(status)
		}
		err := out[3].Interface().(error)
		fmt.Fprintln(w, err) // TODO JSON
		log.Println(w, err)
		return
	}
	if nil != header {
		wHeader := w.Header()
		for key, values := range header {
			wHeader.Del(key)
			for _, value := range values {
				wHeader.Add(key, value)
			}
		}
	}
	w.WriteHeader(status)
	if nil != rs {
		err := json.NewEncoder(w).Encode(rs)
		if nil != err {
			log.Println(err)
		}
	}
}
