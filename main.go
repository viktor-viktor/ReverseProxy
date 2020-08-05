package main

import (
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"proxy/Authentication"
	"proxy/Endpoint"
	log "proxy/Logger"
)

var settingsFile map[string]interface{}

var Addr string

func readSettingFile() {
	env := flag.String("env", "", "a string")
	flag.Parse()
	if *env != "" {
		*env = "." + *env
	}

	// parsing file settings with native go
	jsonFile, err := os.Open("settings" + *env + ".json")
	if err != nil {
		panic("Unable to open settings" + *env + ".json")
	}
	defer jsonFile.Close()

	byteVal, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteVal, &settingsFile)
	if err != nil {panic(err.Error())}
}

func initLogging() {
	var err error

	if v, exist := settingsFile["Logging"]; exist {

		logData, err := log.ReadLoggerDataFromFile(v)
		if err != nil {	panic(err.Error()) } //TODO: panic can be inside (minus line of code)

		f, d := log.PrepareInitData(logData)
		err = log.Init(log.StringLevelToLevel(logData.Level), f, d)
		if err != nil {panic(err.Error())} //TODO: panic can be inside (minus line of code)
	} else {

		err = log.Init(uint32(log.LInfo), log.UseStdOut, nil)
		if err != nil {panic(err.Error())} //TODO: panic can be inside (minus line of code)
	}

}

func readProxyAddr() {
	if v,ok := settingsFile["ProxyAddr"]; ok {
		Addr = v.(string)
	} else {
		//logs
//		Endpoint.l.Error(map[string]string{},"ProxyAddr is required at settings")
		panic("ProxyAddr is required at settings")
	}
}

func initAuth(cl *gin.Engine) {
	err := Authentication.RegisterAuth(cl, settingsFile) //TODO: call panic inside
	if err != nil {
//		Endpoint.l.Error(map[string]string{}, err.Error())
		panic(err.Error())
	}
}

func initEndpoint(cl *gin.Engine) {
	Endpoint.RegisterEndpoints(cl, settingsFile)
}

func main () {

	readSettingFile()
	initLogging()
	readProxyAddr()

	cl := gin.New()
	initAuth(cl)
	initEndpoint(cl)
	
	cl.Run(Addr)
}