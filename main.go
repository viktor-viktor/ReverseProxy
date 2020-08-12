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
	"proxy/Protocol"
	"strconv"
	"sync"
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

	defaultFileName, envFileName := "settings.json", "settings" + *env + ".json"

	defFile, err1 := os.Open(defaultFileName)
	defer func() { if err1 == nil { defFile.Close()}}()
	envFile, err2 := os.Open(envFileName)
	defer func() { if err2 == nil { envFile.Close()}}()

	if err1 != nil && err2 != nil {
		panic("not settings file found under the following environment: " + *env)
	}

	defSettings := map[string]interface{}{}
	if err1 == nil {
		defByteVal, _ := ioutil.ReadAll(defFile)
		if err1 = json.Unmarshal(defByteVal, &defSettings); err1 != nil {
			panic("Can't unmarshal default settings.json file. Error: " + err1.Error())
		}
	}

	envSettings := map[string]interface{}{}
	if err2 == nil {
		envByteVal, _ := ioutil.ReadAll(envFile)
		if err2 = json.Unmarshal(envByteVal, &envSettings); err2 != nil {
			panic("Can't unmarshal environment settings file under env: " + *env + " . Error: " + err2.Error())
		}
	}

	for key, val := range envSettings {
		defSettings[key] = val
	}

	settingsFile = defSettings
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

	Protocol.InitProtocols(settingsFile)

	cl := gin.New()
	initAuth(cl)
	initEndpoint(cl)

	if v, exist := settingsFile["Addr"]; exist {
		v2, ok := v.(string)
		if ok == false {panic("Can't cast 'Addr' field to string")}
		Addr = v2
	}

	wg := sync.WaitGroup{}
	wg.Add(len(Protocol.Protocols))

	for _, v := range Protocol.Protocols {
		if v.Type == "https" {
			go func(p Protocol.Protocol) {
				defer wg.Done()
				addr := Addr + ":" + strconv.Itoa(p.Port)
				err := cl.RunTLS(addr, "TLS/cert.pem", "TLS/key.pem")
				panic("Unable to run TLS server. Error: " + err.Error())
			}(v)
			continue
		}

		go func(p Protocol.Protocol) {
			defer wg.Done()
			// Create address from ip, protocol and port. above the sames
			addr := Addr + ":" + strconv.Itoa(p.Port)
			cl.Run(addr)
		}(v)
	}

	wg.Wait()
}