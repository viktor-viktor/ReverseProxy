package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
)

var Addr string

func TestMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("Before next")
		c.Next()
		fmt.Println("After next")
	}
}

func TestMiddleWare2() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("Before next 2")
		c.AbortWithStatusJSON(401, gin.H{"cool":"error"})
		c.Next()
		fmt.Println("After next 2")
	}
}

func main () {
	
	cl := gin.New()

	///////////////////////////////////////   Testing middlewares /////////////////////////////////
	/*
	cl.Use(TestMiddleWare2())


	test := cl.Group("/")
	test.Handle(http.MethodGet, "test2", func(c *gin.Context) {
		fmt.Println("Inside shitty handle")
		c.JSON(200, gin.H{"asd":123})
	})
	///////////////////////////////////////   Testing middlewares /////////////////////////////////
*/

	// parsing file settings with native go
	jsonFile, err := os.Open("settings.json")
	if err != nil {
		fmt.Print("Unable to open settings.json")
		return
	}
	defer jsonFile.Close()
	var parsedFile map[string]interface{}

	byteVal, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteVal, &parsedFile)

	// Reading addr of the server
	if v,ok := parsedFile["ProxyAddr"]; ok {
		Addr = v.(string)
	} else {
		//logs
		fmt.Println("ProxyAddr is required at settings")
		return
	}

	// Register authentications
	auths, err := ReadAuthFromFile(parsedFile["Auth"])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _,a := range auths {
		middle, err := RegisterMiddleware(a)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		tmp := cl.Group("/")
		tmp.Use(middle)
		AuthMiddlewares[a.Name] = tmp
	}

	// it's better to check if authentication is enabled and register specific middleware in that case

	if val, ok := parsedFile["endpoints"]; ok {
		ReadEndpointsFromFile(cl, val)
	}
	
	cl.Run(Addr)
}