package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"net/http/httputil"
)

//json description of the struct isn't obligatory
type EndpointSettings struct {
	Entry string
	Redir string
	Addr string
	Use_auth bool
	Auth_name string
	Methods []string
}

func (endSet *EndpointSettings) Validate() error {
	if endSet.Entry == "" || endSet.Redir == "" || endSet.Addr == "" {
		return errors.New("missing one of required fields: Entry, Redir, Addr")
	}
	//check if all methods are actual methods
	for _, m := range  endSet.Methods {
		if m != http.MethodGet && m != http.MethodPost && m != http.MethodPut && m != http.MethodDelete {
			return errors.New("Unsupported methods: " + m + "  under entry: " + endSet.Entry)
		}
	}

	if len(endSet.Methods) == 0 {
		endSet.Methods = []string {http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
	}

	return nil
}

func RegisterEndpoint(engine *gin.Engine, settings *EndpointSettings) error {

	var groupRoute *gin.RouterGroup = nil
	if settings.Use_auth {
		if v, ok := AuthMiddlewares[settings.Auth_name]; ok {
			groupRoute = v
		} else {
			return errors.New("Trying to register endpoint with unexisted auth name")
		}
	}

	redirectionMethod := func(c *gin.Context) {
		// still unclear what is the difference between req.Url.Host and req.Host
		director := func(req *http.Request) {
			req.URL.Host = settings.Addr
			req.URL.Path = settings.Redir
			req.URL.Scheme = "http"

			req.Host = settings.Addr
		}

		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	for _, method := range settings.Methods {

		if groupRoute == nil {
			engine.Handle(method, settings.Entry, redirectionMethod)
		} else {
			groupRoute.Handle(method, settings.Entry, redirectionMethod)
		}
	}

	return nil
}

func ReadEndpointsFromFile(cl *gin.Engine, file interface{}) error {
	val2,ok := file.([]interface{})
	if ok == false {
		return errors.New("Can't cast interface{} to []interface{}")
	}

	var endpoints []EndpointSettings
	for _,v := range val2 {
		var tmp EndpointSettings
		mapstructure.Decode(v, &tmp)
		endpoints = append(endpoints, tmp)
	}

	for _, endp := range endpoints {
		err := endp.Validate()
		if err != nil {
			return err
		}
		RegisterEndpoint(cl, &endp)
	}

	return nil
}
