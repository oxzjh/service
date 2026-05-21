package service

import (
	"errors"
	"reflect"
	"strings"
)

type packet struct {
	method string
	args   any
	reply  any
	err    error
	done   chan struct{}
}

type client struct {
	server  *Server
	packetC chan *packet
}

func (c *client) Call(method string, args, reply any) error {
	p := &packet{method: method, args: args, reply: reply, done: make(chan struct{}, 1)}
	c.packetC <- p
	<-p.done
	return p.err
}

func (c *client) run() {
	for p := range c.packetC {
		p.err = c.exec(p)
		close(p.done)
	}
}

func (c *client) exec(p *packet) error {
	dot := strings.LastIndex(p.method, ".")
	if dot < 0 {
		return errors.New("illegal method: " + p.method)
	}
	serviceName := p.method[:dot]
	methodName := p.method[dot+1:]
	service, ok := c.server.services[serviceName]
	if !ok {
		return errors.New("can't find service: " + serviceName)
	}
	return service.call(methodName, reflect.ValueOf(p.args), reflect.ValueOf(p.reply))
}

func NewClient(server *Server) IService {
	c := &client{server, make(chan *packet, 128)}
	go c.run()
	return c
}
