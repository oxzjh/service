package service

import (
	"errors"
	"reflect"
	"strings"
)

type Call struct {
	Method string        // The name of the service and method to call.
	Args   any           // The argument to the function (*struct).
	Reply  any           // The reply from the function (*struct).
	Error  error         // After completion, the error status.
	Done   chan struct{} // Receives *Call when Go is complete.
}

type Client struct {
	server *Server
	callC  chan *Call
}

func (c *Client) Call(method string, args, reply any) error {
	call := &Call{Method: method, Args: args, Reply: reply, Done: make(chan struct{}, 1)}
	c.callC <- call
	<-call.Done
	return call.Error
}

func (c *Client) run() {
	for call := range c.callC {
		call.Error = c.exec(call)
		close(call.Done)
	}
}

func (c *Client) exec(call *Call) error {
	dot := strings.LastIndex(call.Method, ".")
	if dot < 0 {
		return errors.New("illegal method: " + call.Method)
	}
	serviceName := call.Method[:dot]
	methodName := call.Method[dot+1:]
	service, ok := c.server.services[serviceName]
	if !ok {
		return errors.New("can't find service: " + serviceName)
	}
	return service.call(methodName, reflect.ValueOf(call.Args), reflect.ValueOf(call.Reply))
}

func NewClient(server *Server) *Client {
	c := &Client{server, make(chan *Call, 128)}
	go c.run()
	return c
}
