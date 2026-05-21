package service

import (
	"errors"
	"log"
	"reflect"
	"runtime/debug"
)

type IService interface {
	Call(string, any, any) error
}

type service struct {
	recv  reflect.Value
	funcs map[string]reflect.Value
}

func (s *service) call(method string, argv, replyv reflect.Value) (err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			log.Writer().Write(debug.Stack())
			err = errors.New("panic")
		}
	}()
	function, ok := s.funcs[method]
	if !ok {
		err = errors.New("can't find method: " + method)
		return
	}
	returnValues := function.Call([]reflect.Value{s.recv, argv, replyv})
	if errInterface := returnValues[0].Interface(); errInterface != nil {
		return errInterface.(error)
	}
	return nil
}
