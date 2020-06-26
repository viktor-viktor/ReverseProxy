package Endpoint

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"net/http/httputil"
	auth "proxy/Authentication"
	"proxy/Logger"
	"proxy/Protocol"
)

var l *Logger.Logger

//json description of the struct isn't obligatory
type EndpointSettings struct {
	Entry_url string
	Redir_url string
	Redir_addr string
	Use_auth bool
	Auth_name string
	Protocol string
	Methods []string
}

func (endSet *EndpointSettings) Validate() error {
	if endSet.Entry_url == "" || endSet.Redir_addr == "" {
		return errors.New("missing one of required fields: Entry, Redir, Addr")
	}
	//check if all methods are actual methods
	for _, m := range  endSet.Methods {
		if m != http.MethodGet && m != http.MethodPost && m != http.MethodPut && m != http.MethodDelete {
			return errors.New("Unsupported methods: " + m + "  under entry: " + endSet.Entry_url)
		}
	}

	if len(endSet.Methods) == 0 {
		endSet.Methods = []string {http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete}
	}

	if endSet.Protocol == "" {
		endSet.Protocol = Protocol.Protocols[0].Type
	} else {
		found := false
		for _,v := range Protocol.Protocols {
			if v.Type == endSet.Protocol {
				found = true
				break
			}
		}

		if found == false {panic("Endpoint " + endSet.Entry_url + "  use invalid protocol: " + endSet.Protocol)}
	}

	if endSet.Use_auth && endSet.Auth_name == "" {
		return errors.New("Auth name should be specified if 'use_auth' is true")
	}

	return nil
}

func registerEndpoint(engine *gin.Engine, settings *EndpointSettings) {

	var groupRoute *gin.RouterGroup = nil
	if settings.Use_auth {
		if v, ok := auth.AuthMiddlewares[settings.Auth_name]; ok {
			groupRoute = v
		} else {
			panic("Trying to register endpoint with unexisted auth name: " + settings.Auth_name)
		}
	}

	redirectionMethod := func(c *gin.Context) {
		l.Info(map[string]string{}, "Request made for Url: " + settings.Entry_url)

		// still unclear what is the difference between req.Url.Host and req.Host
		director := func(req *http.Request) {
			req.URL.Host = settings.Redir_addr
			req.URL.Path = settings.Redir_url
			req.URL.Scheme = settings.Protocol

			req.Host = settings.Redir_addr
		}

		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	for _, method := range settings.Methods {

		if groupRoute == nil {
			engine.Handle(method, settings.Entry_url, redirectionMethod)
		} else {
			groupRoute.Handle(method, settings.Entry_url, redirectionMethod)
		}
	}
}

func RegisterEndpoints(cl *gin.Engine, file map[string]interface{}) {
	if l == nil { l = Logger.New("Endpoint", 0, nil) }

	if val, ok := file["endpoints"]; ok {
		readEndpointsFromFile(cl, val)
	} else {
		panic("There is no section 'endpoints' in settings.json file")
	}
}

func readEndpointsFromFile(cl *gin.Engine, file interface{}) {
	val2, ok := file.([]interface{})
	if ok == false {
		panic("Can't cast interface{} to []interface{} when parsing 'endpoints' json value")
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
			panic(err.Error())
		}
		registerEndpoint(cl, &endp)
	}
}
