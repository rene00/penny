package logwrap

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	DEBUG   = 0
	INFO    = 10
	WARNING = 20
	ERROR   = 30
	NONE    = 100
)

var (
	pkgLock sync.Mutex
	allLogs = make(map[string]*LogWrap)
)

type LogWrap struct {
	name         string
	level        int
	errorLogWrap *log.Logger
	debugLogWrap *log.Logger
	infoLogWrap  *log.Logger
	logfileinfo  bool
	lock         sync.Mutex
}

func New(name string, dest io.Writer, logfileinfo bool, flags ...int) *LogWrap {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	if _, exists := allLogs[name]; exists {
		panic(fmt.Sprintf("Unable to create logger with %s: name already in use", name))
	}

	var logFlags int
	if flags == nil {
		logFlags = log.Ldate | log.Ltime | log.Lmsgprefix
	} else {
		logFlags = flags[0]
	}

	logger := &LogWrap{
		name:         name,
		logfileinfo:  logfileinfo,
		errorLogWrap: log.New(dest, "ERROR: ", logFlags),
		debugLogWrap: log.New(dest, "DEBUG: ", logFlags),
		infoLogWrap:  log.New(dest, "INFO: ", logFlags),
	}

	allLogs[name] = logger
	return logger
}

func getfileinfo() string {
	_, filename, line, ok := runtime.Caller(2)
	if !ok {
		filename = "Unknown"
		line = 0
	}
	return fmt.Sprintf("%s:%d: ", filepath.Base(filename), line)
}

func (l *LogWrap) SetLevel(level int) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level
}

func (l *LogWrap) Info(msg string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.level > INFO {
		return
	}
	if l.logfileinfo {
		msg = getfileinfo() + msg
	}
	l.infoLogWrap.Println(msg)
}

func (l *LogWrap) Error(msg string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.level > ERROR {
		return
	}
	if l.logfileinfo {
		msg = getfileinfo() + msg
	}
	l.errorLogWrap.Println(msg)
}

func (l *LogWrap) Debug(msg string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.level > DEBUG {
		return
	}
	if l.logfileinfo {
		msg = getfileinfo() + msg
	}
	l.debugLogWrap.Println(msg)
}

func Get(name string) *LogWrap {
	pkgLock.Lock()
	defer pkgLock.Unlock()
	if foundLog, ok := allLogs[name]; ok {
		return foundLog
	}
	return nil
}
