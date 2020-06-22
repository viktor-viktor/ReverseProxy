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
	if err != nil {panic(err.Error())}

	env := flag.String("env", "", "a string")
	flag.Parse()
	if *env != "" {
		*env = "." + *env
	}
	
	cl := gin.New()

	// parsing file settings with native go
	jsonFile, err := os.Open("settings" + *env + ".json")
	if err != nil {
		panic("Unable to open settings" + *env + ".json")
	}
	defer jsonFile.Close()
	var parsedFile map[string]interface{}

	byteVal, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteVal, &parsedFile)
	if err != nil {panic(err.Error())}

	if v, exist := parsedFile["Logging"]; exist {

		logData, err := log.ReadLoggerDataFromFile(v)
		if err != nil {	panic(err.Error()) }

		f, d := log.PrepareInitData(logData)
		err = log.Init(log.StringLevelToLevel(logData.Level), f, d)
		if err != nil {panic(err.Error())}
	} else {

		err = log.Init(uint32(log.LInfo), log.UseStdOut, nil)
		if err != nil {panic(err.Error())}
	}

	//l, err := log.New("main", 0, nil)



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