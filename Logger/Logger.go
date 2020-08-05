package Logger

import (
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-extras/elogrus.v7"
	"os"
)

const (
	// flagslags that init where logs should go
	UseFile 	= 1
	UseStdOut  	= 2
	UseElastic  = 4

	// data key names
	FilePath = "FileOutPath"
	ElasticUrl = "ElasticUrl"

	//Levels
	LDebug = logrus.DebugLevel
	LInfo = logrus.InfoLevel
	LWarning = logrus.WarnLevel
	LError = logrus.ErrorLevel

)

// default settings //
// specify if std output should be used
var useStdOut = false
// path to default file output
var fileOutPath string
// url to default elastic search server
var elasticUrl string
// shows if default settings have been init
var defaultInit = false
// default logger
var defaultLogger Logger
// levels of logging that are allowed
var levels []logrus.Level

type InitData struct {
	Level string
	UseStd bool
	UseElastic string
	UseFile string
}

type Logger struct {
	log *logrus.Logger
	name string
}

func mapToFields(data map[string]string) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range data {
		out[k] = v
	}

	return out
}

func getLevelGroupFromLevel(level uint32) []logrus.Level {

	levels := []logrus.Level{LDebug, LInfo, LWarning, LError}
	index := -1
	for i, v := range levels {
		if v == logrus.Level(level) {
			index = i
			break
		}
	}

	if index == -1 {panic("given invalid log level type. Type should be one of: {LDebug, LInfo, LWarning, LError}")}
	return levels[index:]
}

// add std output as a hook for given logger
func configureStdOutput(l *logrus.Logger) {
	h := WriteHook{writer: os.Stdout, levels: levels}
	l.AddHook(&h)
}

// adds file output as hook for given logger
func configureFileOutput(l *logrus.Logger, path string) {
	file, err := os.OpenFile(path, os.O_CREATE | os.O_APPEND , 0666)
	if err != nil {panic("Failed to open file while 'configureFileOutput'. Error: " + err.Error())}
	h := WriteHook{writer: file, levels:levels }
	l.AddHook(&h)
}

// adds elastic search output as hook for given logger
func configureElasticOutput(l *logrus.Logger, url string) {

	client, err := elasticsearch7.NewClient(elasticsearch7.Config{
		Addresses: []string{url},
	})
	if err != nil { panic("Failed to create ElasticSearch client. Errpr: " + err.Error()) }

	hook, err := elogrus.NewAsyncElasticHook(client, "localhost", logrus.Level(levels[0]), "mylog")
	if err != nil { panic("Failed to create ElasticSearch hook client. Errpr: " + err.Error()) }

	l.AddHook(hook)
}

// Init default settings of logger. Possible settings:
// useStdOut - bool
// fileOutPath - string
// elasticUrl - string
// level - on of {LDebug, LInfo, LWarning, LError}
func Init(level uint32, flags int, data map[string]string) {
	if flags == 0 { panic("Logger.Init(); At least one flags for output should be specified")}
	
	levels = getLevelGroupFromLevel(level)

	if flags&UseFile != 0 {
		if v, exist := data[FilePath]; exist {
			fileOutPath = v
		} else {panic("flag 'UseFile' specified but not file path at data ! Logger.Init()")}
	}

	if flags&UseStdOut != 0 {
		useStdOut = true
	}

	if flags&UseElastic != 0 {
		if v, exist := data[ElasticUrl]; exist {
			elasticUrl = v
		} else {panic("flag 'UseElastic' specified but not url path at data ! Logger.Init()")}
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	if useStdOut {
		configureStdOutput(logrus.StandardLogger())
	}

	if fileOutPath != "" {
		configureFileOutput(logrus.StandardLogger(), fileOutPath)
	}

	if elasticUrl != "" {
		configureElasticOutput(logrus.StandardLogger(), elasticUrl)
	}

	defaultLogger = Logger{log: logrus.StandardLogger(), name: "default"}
	defaultInit = true
}

// Read data from parsed json file to InitData structure that can be used to init logger
func ReadLoggerDataFromFile(d interface{}) *InitData {

	var data InitData
	err := mapstructure.Decode(d, &data)
	if err != nil {panic("Can't decode logger data from json to struct. Error: " + err.Error())}

	return &data
}

// Converts string logger level to logger level
func StringLevelToLevel(l string) uint32 {
	switch l {
	case "Debug":
		return uint32(LDebug)
	case "Info":
		return uint32(LInfo)
	case "Warning":
		return uint32(LWarning)
	case "Error":
		return uint32(LError)
	default:
		panic("Invalid string level is within config files !")
	}
}

// converts InitData to falgs and data map[string]string
func PrepareInitData(d *InitData) (int, map[string]string) {

	flags := 0
	data := map[string]string{}

	if d.UseStd == true { flags = flags | UseStdOut	}
	if d.UseElastic != "" {
		flags = flags | UseElastic
		data[ElasticUrl] = d.UseElastic
	}
	if d.UseFile != "" {
		flags = flags | UseFile
		data[FilePath] = d.UseFile
	}

	return flags, data
}

// Inits new logger and return pointer to it
// flags - one of the { UseFile UseStdOut UseElastic }
// data - for each of the keys data should specified (except of UseStdOut)
// if no data for key specified, default value is used, if no default value exist - error is returned
func New(name string, flags int, data map[string]string) *Logger {

	l := Logger{}
	l.log = logrus.New()
	l.name = name
	
	if flags == 0 {
		if defaultInit == false {
			Error(map[string]string{}, "can't create logger while default settings aren't init")
			return &defaultLogger
		}
		if useStdOut {configureStdOutput(l.log)}
		if fileOutPath != "" {configureFileOutput(l.log, fileOutPath)}
		if elasticUrl != "" {configureElasticOutput(l.log, elasticUrl)}

		return &l
	}

	if flags&UseFile != 0 {
		if v, exist := data[FilePath]; exist {
			configureFileOutput(l.log, v)
		} else if fileOutPath != "" {
			configureFileOutput(l.log, fileOutPath)
		} else {
			Error(map[string]string{}, "can't create logger with file output as neither data or default data specified")
			return &defaultLogger
		}
	}

	if useStdOut == true || flags&UseStdOut != 0 {
		configureStdOutput(l.log)
	}

	if flags&UseElastic != 0 {
		if v, exist := data[ElasticUrl]; exist {
			configureElasticOutput(l.log, v)
		} else if elasticUrl != "" {
			configureElasticOutput(l.log, elasticUrl)
		} else {
			Error(map[string]string{}, "can't create logger with elastic output as neither data or default data specified")
			return &defaultLogger
		}
	}

	return &l
}

func (l *Logger) Debug(data map[string]string, message string) {
	data["logger"] = l.name
	l.log.WithFields(mapToFields(data)).Debug(message)
}

func (l *Logger) Info(data map[string]string, message string) {
	data["logger"] = l.name
	l.log.WithFields(mapToFields(data)).Info(message)
}

func (l *Logger) Warning(data map[string]string, message string) {
	data["logger"] = l.name
	l.log.WithFields(mapToFields(data)).Warning(message)
}

func (l *Logger) Error(data map[string]string, message string) {
	data["logger"] = l.name
	l.log.WithFields(mapToFields(data)).Error(message)
}

func Debug(name string, data map[string]string, message string) {
	data["logger"] = "default"
	logrus.StandardLogger().WithFields(mapToFields(data)).Debug(message)
}

func Info(name string, data map[string]string, message string) {
	data["logger"] = "default"
	logrus.StandardLogger().WithFields(mapToFields(data)).Info(message)
}

func Warning(data map[string]string, message string) {
	data["logger"] = "default"
	logrus.StandardLogger().WithFields(mapToFields(data)).Warning(message)
}

func Error(data map[string]string, message string) {
	data["logger"] = "default"
	logrus.StandardLogger().WithFields(mapToFields(data)).Error(message)
}

