package Logger

import (
	"errors"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-extras/elogrus.v7"
	"os"
)

const (
	// flags that init where logs should go
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

// map of all loggers
var loggers map[string]Logger = map[string]Logger{}
// default settings //
// specify if std output should be used
var useStdOut = false
// path to default file output
var fileOutPath string
// url to default elastic search server
var elasticUrl string
// shows if default settings have been init
var defaultInit = false
// levels of logging that are allowed
var levels []logrus.Level

type Logger struct {
	log *logrus.Logger
	name string
}

func getLevelGroupFromLevel(level uint32) ([]logrus.Level,error) {

	levels := []logrus.Level{LDebug, LInfo, LWarning, LError}
	index := -1
	for i, v := range levels {
		if v == logrus.Level(level) {
			index = i
			break
		}
	}

	if index == -1 {return nil, errors.New("given invalid log level type. Type should be one of: {LDebug, LInfo, LWarning, LError}")}
	logrus.Info()
	return levels[index:], nil
}

// add std output as a hook for given logger
func configureStdOutput(l *logrus.Logger) error {
	h := WriteHook{writer: os.Stdout, levels: levels}
	l.AddHook(&h)
	return nil
}

// adds file output as hook for given logger
func configureFileOutput(l *logrus.Logger, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE | os.O_APPEND , 0666)
	if err != nil {return err}
	h := WriteHook{writer: file, levels:levels }
	l.AddHook(&h)
	return nil
}

// adds elastic search output as hook for given logger
func configureElasticOutput(l *logrus.Logger, url string) error {

	client, err := elasticsearch7.NewClient(elasticsearch7.Config{
		Addresses: []string{url},
	})
	if err != nil { return err }

	hook, err := elogrus.NewAsyncElasticHook(client, "localhost", logrus.Level(levels[0]), "mylog")
	if err != nil {return nil}

	logrus.AddHook(hook)

	return nil
}

// Init default settings of logger. Possible settings:
// useStdOut - bool
// fileOutPath - string
// elasticUrl - string
// level - on of {LDebug, LInfo, LWarning, LError}
func Init(level uint32, flags int, data map[string]string) error {
	if flags == 0 { return errors.New("Logger.Init(); At least one flags for output should be specified")}
	var err error
	
	levels, err = getLevelGroupFromLevel(level)
	if err != nil {return err}

	if flags&UseFile != 0 {
		if v, exist := data[FilePath]; exist {
			fileOutPath = v
		} else {return errors.New("flag 'UseFile' specified but not file path at data ! Logger.Init()")}
	}

	if flags&UseStdOut != 0 {
		useStdOut = true
	}

	if flags&UseElastic != 0 {
		if v, exist := data[ElasticUrl]; exist {
			elasticUrl = v
		} else {errors.New("flag 'UseElastic' specified but not url path at data ! Logger.Init()")}
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	if useStdOut {
		configureStdOutput(logrus.StandardLogger())
	}

	if fileOutPath != "" {
		err := configureFileOutput(logrus.StandardLogger(), fileOutPath)
		if err != nil {return err}
	}

	if elasticUrl != "" {
		err := configureElasticOutput(logrus.StandardLogger(), elasticUrl)
		if err != nil {return err}
	}

	defaultInit = true
	return nil
}

// Inits new logger and return pointer to it
// flags - one of the { UseFile UseStdOut UseElastic }
// data - for each of the keys data should specified (except of UseStdOut)
// if no data for key specified, default value is used, if no default value exist - error is returned
func New(name string, flags int, data map[string]string) (*Logger, error) {

	l := Logger{}
	l.log = logrus.New()
	l.name = name
	
	if flags == 0 {
		if defaultInit == false {return nil, errors.New("can't create logger while default settings aren't init")}
		// use default settings
		return nil, errors.New("can't create logger without output specified")
	}

	if flags&UseFile != 0 {
		if v, exist := data[FilePath]; exist {
			configureFileOutput(l.log, v)
		} else if fileOutPath != "" {
			configureFileOutput(l.log, fileOutPath)
		} else {return nil, errors.New("can't create logger with file output as neither data or default data specified")}
	}

	if useStdOut == true || flags&UseStdOut != 0 {
		configureStdOutput(l.log)
	}

	if flags&UseElastic != 0 {
		if v, exist := data[ElasticUrl]; exist {
			configureElasticOutput(l.log, v)
		} else if elasticUrl != "" {
			configureElasticOutput(l.log, elasticUrl)
		} else {return nil, errors.New("can't create logger with elastic output as neither data or default data specified")}
	}

	loggers[name] = l

	return &l, nil
}

func (l *Logger) Debug(data map[string]string) {
	data["logger"] = l.name
	l.log.Debug(data)
}

func (l *Logger) Info(data map[string]string) {
	data["logger"] = l.name
	l.log.Info(data)
}

func (l *Logger) Warning(data map[string]string) {
	data["logger"] = l.name
	l.log.Warning(data)
}

func (l *Logger) Error(data map[string]string) {
	data["logger"] = l.name
	l.log.Error(data)
}

func handleInvalidName(name string) error {
	logrus.WithFields(logrus.Fields{"Message": "can't find logger with nane: " + name}).Warning()
	return errors.New("logger with name: " + name + " doesn't exist")
}

func Debug(name string, data map[string]string) error {
	if v, exist := loggers[name]; exist {
		v.Debug(data)
		return nil
	}
	return handleInvalidName(name)
}

func Info(name string, data map[string]string) error {
	if v, exist := loggers[name]; exist {
		v.Info(data)
		return nil
	}
	return handleInvalidName(name)
}

func Warning(name string, data map[string]string) error {
	if v, exist := loggers[name]; exist {
		v.Warning(data)
		return nil
	}
	return handleInvalidName(name)
}

func Error(name string, data map[string]string) error {
	if v, exist := loggers[name]; exist {
		v.Error(data)
		return nil
	}
	return handleInvalidName(name)
}