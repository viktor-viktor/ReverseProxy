package main

import (
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	log "proxy/Logger"
)

var Addr string

func init() {

}

func main () {

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

	l := log.New("main", 0, nil)


	// Reading addr of the server
	if v,ok := parsedFile["ProxyAddr"]; ok {
		Addr = v.(string)
	} else {
		//logs
		l.Error(map[string]string{},"ProxyAddr is required at settings")
		return
	}

	err = RegisterAuth(cl, parsedFile)
	if err != nil {
		l.Error(map[string]string{}, err.Error())
		return
	}

	err = RegisterEndpoints(cl, parsedFile)
	if err != nil {
		l.Error(map[string]string{}, err.Error())
		return
	}
	
	cl.Run(Addr)
}