package service

import (
	"errors"
	"go/token"
	"reflect"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

type Server struct {
	services map[string]*service
}

func (s *Server) Register(rcvr any, name string) error {
	typ := reflect.TypeOf(rcvr)
	val := reflect.ValueOf(rcvr)
	if name == "" {
		name = reflect.Indirect(val).Type().Name()
	}
	if name == "" {
		return errors.New("no service name for type " + typ.String())
	}
	if _, ok := s.services[name]; ok {
		return errors.New("service already defined: " + name)
	}
	funcs := make(map[string]reflect.Value)
	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)
		// Method must be exported.
		if !m.IsExported() {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if m.Type.NumIn() != 3 {
			continue
		}
		// First arg need not be a pointer.
		argType := m.Type.In(1)
		if !isExportedOrBuiltinType(argType) {
			continue
		}
		// Second arg must be a pointer.
		replyType := m.Type.In(2)
		if replyType.Kind() != reflect.Pointer {
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			continue
		}
		// Method needs one out.
		if m.Type.NumOut() != 1 {
			continue
		}
		// The return type of the method must be error.
		if returnType := m.Type.Out(0); returnType != typeOfError {
			continue
		}
		funcs[m.Name] = m.Func
	}
	if len(funcs) == 0 {
		return errors.New("type " + name + " has no exported methods of sutable type")
	}
	s.services[name] = &service{val, funcs}
	return nil
}

func NewServer() *Server {
	return &Server{make(map[string]*service)}
}
