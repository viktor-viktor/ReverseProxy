package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	log "proxy/Logger"
)

var Addr string

func main () {

	err := log.Init(uint32(log.LInfo), log.UseElastic | log.UseStdOut, map[string]string{log.ElasticUrl: "http://192.168.0.101:9200/"})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	l, err := log.New("test", 0, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	l.Info(map[string]string{"test": "accepted"})

	env := flag.String("env", "", "a string")
	flag.Parse()
	if *env != "" {
		*env = "." + *env
	}

	
	cl := gin.New()

	// parsing file settings with native go
	jsonFile, err := os.Open("settings" + *env + ".json")
	if err != nil {
		fmt.Print("Unable to open settings" + *env + ".json")
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