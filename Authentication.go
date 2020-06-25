package main

//TODO: better way for import
import (
	"errors"
	"io/ioutil"
	"proxy/Logger"
	"strings"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"net/http"
)

type Authentication struct {
	Name string
	Auth_scheme string
	Auth_addr string
	Auth_type string
	Url_path string
	Req_headers []string
}

var lauth *Logger.Logger
var AuthMiddlewares = map[string]*gin.RouterGroup{}

func (auth *Authentication) Validate() error {
	var rv string = ""
	if auth.Name == "" {
		rv += "name, "
	}
	if auth.Auth_addr == "" {
		rv += "auth_name, "
	}
	if auth.Auth_scheme == "" {
		rv += "auth_scheme, "
	} else if auth.Auth_scheme != "http" {
		return errors.New("Invalid auth_scheme provided: " + auth.Auth_scheme + " . Currently supported only http")
	}
	if auth.Auth_type == "" {
		rv += "auth_type, "
	} else if auth.Auth_type != "epp" {
		return errors.New("Supported auth type is only 'epp'")
	}
	if auth.Url_path == "" {
		rv += "url_path, "
	}

	if rv != "" {
		rv =  "Missing required fields: " + rv
		return errors.New(rv)
	}

	return nil
}

//TODO: investigate method, it's comparably slow
func ReadAuthFromFile(auth interface{}) ([]Authentication, error) {
	lauth = Logger.New("Authentication", 0, nil)

	var rv []Authentication

	auth2, ok := auth.([]interface{})
	if ok == false {
		return nil, errors.New("Can't cast Authentication interface{} to map[string]interface{}")
	}
	for _,v := range auth2 {
		var tmp Authentication
		err := mapstructure.Decode(v, &tmp)
		if err != nil {
			return nil, err
		}
		err = tmp.Validate()
		if err != nil {return nil, err}
		rv = append(rv, tmp)
	}

	return rv, nil
}

func RegisterAuth(cl *gin.Engine, file map[string]interface{}) error {

	auths, err := ReadAuthFromFile(file["Auth"])
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	for _,a := range auths {
		middle, err := RegisterMiddleware(a)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		tmp := cl.Group("/")
		tmp.Use(middle)
		AuthMiddlewares[a.Name] = tmp
	}

	return nil
}

//TODO: change to return error
func DefaultAuthMiddleware(auth Authentication) gin.HandlerFunc {

	return func(c *gin.Context) {

		//init client and request
		cl := http.Client{}
		req, err := http.NewRequest("GET", auth.Auth_scheme + "://" + auth.Auth_addr + auth.Url_path, nil)
		if err != nil {
			lauth.Error(map[string]string{"Error": err.Error()}, "Error while creating new 'Request'")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		//fill request with required auth headers
		for _,h := range auth.Req_headers {
			if v := c.Request.Header.Get(h); len(v) != 0 {
				req.Header.Add(h, v)
			} else {
				lauth.Error(
					map[string]string{"Missing header": h, "Required headers": strings.Join(auth.Req_headers[:], ",")},
				"Missing required header")
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Missing required header: " + h})
				return
			}
		}

		// send request to auth server
		resp, err := cl.Do(req)
		if err != nil {
			lauth.Error(map[string]string{"Error": err.Error()}, "Error when trying to send auth request")
			code := http.StatusInternalServerError
			if resp != nil { code = resp.StatusCode }
			c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
			return
		} else if resp.StatusCode >= 300 { //need to check status not only for 200

			bodyB, _ := ioutil.ReadAll(resp.Body)
			lauth.Error(map[string]string{"Status": resp.Status, "Response": string(bodyB)},
			"Unsuccessful code return form auth service")
			c.AbortWithStatusJSON(resp.StatusCode, gin.H{"Status": resp.Status, "body": resp.Body})
			return
		}

		//process if authorized
		c.Next()
	}
}

func RegisterMiddleware(auth Authentication) (gin.HandlerFunc, error) {
	// currently supported only type 'endpoint per permission
	if auth.Auth_type == "epp" {
		return DefaultAuthMiddleware(auth), nil
	}
	return nil, errors.New("Unsupported auth_type is used")
}