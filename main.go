package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
)

var Addr string

func main () {
	
	cl := gin.New()

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

	err = RegisterAuth(cl, parsedFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = RegisterEndpoints(cl, parsedFile)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	
	cl.Run(Addr)
}