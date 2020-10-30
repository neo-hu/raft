package logger

import (
	"log"
	"os"
)

type Logger interface {
	Debug(args ...interface{})
	Debugln(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Warning(args ...interface{})
	Warningln(args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})
	V(l int) bool
}

var logger = newLogger()

func SetLogger(l Logger) {
	logger = l
}

const (
	debugLog int = iota
	infoLog
	warningLog
	errorLog
	fatalLog
)

var severityName = []string{
	debugLog:   "DEBUG",
	infoLog:    "INFO",
	warningLog: "WARNING",
	errorLog:   "ERROR",
	fatalLog:   "FATAL",
}

type loggerT struct {
	m []*log.Logger
	v int
}

func newLogger() Logger {
	var m []*log.Logger
	m = append(m, log.New(os.Stderr, severityName[debugLog]+": ", log.LstdFlags))
	m = append(m, log.New(os.Stderr, severityName[infoLog]+": ", log.LstdFlags))
	m = append(m, log.New(os.Stderr, severityName[warningLog]+": ", log.LstdFlags))
	m = append(m, log.New(os.Stderr, severityName[errorLog]+": ", log.LstdFlags))
	m = append(m, log.New(os.Stderr, severityName[fatalLog]+": ", log.LstdFlags))
	return &loggerT{m: m, v: debugLog}
}

func (g *loggerT) Debug(args ...interface{}) {
	g.m[debugLog].Print(args...)

}
func (g *loggerT) Debugln(args ...interface{}) {
	g.m[debugLog].Println(args...)

}
func (g *loggerT) Debugf(format string, args ...interface{}) {
	g.m[debugLog].Printf(format, args...)
}

func (g *loggerT) Info(args ...interface{}) {
	g.m[infoLog].Print(args...)
}

func (g *loggerT) Infoln(args ...interface{}) {
	g.m[infoLog].Println(args...)
}

func (g *loggerT) Infof(format string, args ...interface{}) {
	g.m[infoLog].Printf(format, args...)
}

func (g *loggerT) Warning(args ...interface{}) {
	g.m[warningLog].Print(args...)
}

func (g *loggerT) Warningln(args ...interface{}) {
	g.m[warningLog].Println(args...)
}

func (g *loggerT) Warningf(format string, args ...interface{}) {
	g.m[warningLog].Printf(format, args...)
}

func (g *loggerT) Error(args ...interface{}) {
	g.m[errorLog].Print(args...)
}

func (g *loggerT) Errorln(args ...interface{}) {
	g.m[errorLog].Println(args...)
}

func (g *loggerT) Errorf(format string, args ...interface{}) {
	g.m[errorLog].Printf(format, args...)
}

func (g *loggerT) Fatal(args ...interface{}) {
	g.m[fatalLog].Fatal(args...)
}

func (g *loggerT) Fatalln(args ...interface{}) {
	g.m[fatalLog].Fatalln(args...)
}

func (g *loggerT) Fatalf(format string, args ...interface{}) {
	g.m[fatalLog].Fatalf(format, args...)
}

func (g *loggerT) V(l int) bool {
	return l <= g.v
}

func V(l int) bool {
	return logger.V(l)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}
func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

func Warning(args ...interface{}) {
	logger.Warning(args...)
}

func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

func Warningln(args ...interface{}) {
	logger.Warningln(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
	os.Exit(1)
}

func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
	os.Exit(1)
}

func Print(args ...interface{}) {
	logger.Info(args...)
}

func Printf(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Println(args ...interface{}) {
	logger.Infoln(args...)
}
