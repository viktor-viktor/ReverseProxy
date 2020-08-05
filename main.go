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

var (
	 Addr string
	 l *log.Logger
)


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
	if v, exist := settingsFile["Logging"]; exist {

		logData := log.ReadLoggerDataFromFile(v)
		flags, data := log.PrepareInitData(logData)
		log.Init(log.StringLevelToLevel(logData.Level), flags, data)
	} else {

		log.Init(uint32(log.LInfo), log.UseStdOut, nil)
	}
}

func readProxyAddr() {
	if v,ok := settingsFile["ProxyAddr"]; ok {
		Addr = v.(string)
	} else {
		l.Error(map[string]string{},"ProxyAddr is required at settings")
		panic("ProxyAddr is required at settings")
	}
}

func initAuth(cl *gin.Engine) {
	defer func() {
		if r:=recover(); r != nil {
			if message, ok := r.(string); ok {
				l.Error(map[string]string{}, message,)
			}
			panic(r)
		}
	}()
	Authentication.RegisterAuth(cl, settingsFile)
}

func initEndpoint(cl *gin.Engine) {
	defer func() {
		if r:=recover(); r != nil {
			if message, ok := r.(string); ok {
				l.Error(map[string]string{}, message,)
			}
			panic(r)
		}
	}()

	Endpoint.RegisterEndpoints(cl, settingsFile)
}

func main () {

	readSettingFile()
	initLogging()
	l = log.New("main", 0, map[string]string{})

	readProxyAddr()

	cl := gin.New()
	initAuth(cl)
	initEndpoint(cl)
	
	cl.Run(Addr)
}