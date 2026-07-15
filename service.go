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

type serviceMethod struct {
	function  reflect.Value
	argType   reflect.Type
	replyType reflect.Type
}

type service struct {
	recv    reflect.Value
	methods map[string]*serviceMethod
}

func (s *service) call(name string, arg, reply any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			log.Writer().Write(debug.Stack())
			err = errors.New("panic")
		}
	}()
	method, ok := s.methods[name]
	if !ok {
		err = errors.New("can't find method: " + name)
		return
	}
	var (
		argv   reflect.Value
		replyv reflect.Value
	)
	if arg == nil {
		argv = reflect.New(method.argType)
	} else {
		argv = reflect.ValueOf(arg)
	}
	if reply == nil {
		replyv = reflect.New(method.replyType)
	} else {
		replyv = reflect.ValueOf(reply)
	}
	returnValues := method.function.Call([]reflect.Value{s.recv, argv, replyv})
	if errInterface := returnValues[0].Interface(); errInterface != nil {
		err = errInterface.(error)
	}
	return
}

func New(name string, rcvr any) (IService, error) {
	server := NewServer()
	if err := server.Register(rcvr, name); err != nil {
		return nil, err
	}
	return NewClient(server), nil
}
