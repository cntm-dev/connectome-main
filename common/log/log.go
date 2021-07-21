package log

import (
	"GoOnchain/common"
	"GoOnchain/config"
	"bytes"
	"fmt"
	"path/filepath"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	PRINTLEVEL = 0
)

const (
	debugLog = iota
	infoLog
	warnLog
	errorLog
	fatalLog
	printLog
	maxLevelLog
)

var (
	levels = map[int]string{
		debugLog: Color(Green, "[DEBUG]"),
		infoLog:  Color(Green, "[INFO ]"),
		warnLog:  Color(Yellow, "[WARN ]"),
		errorLog: Color(Red, "[ERROR]"),
		fatalLog: Color(Red, "[FATAL]"),
		traceLog: Color(Pink, "[TRACE]"),
	}
	Stdout = os.Stdout
)

const (
	namePrefix        = "LEVEL"
	callDepth         = 2
	defaultMaxLogSize = 20
	byteToMb          = 1024 * 1024
	byteToKb          = 1024
	Path              = "./Log/"
)

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

var Log *Logger

func LevelName(level int) string {
	if name, ok := levels[level]; ok {
		return name
	}
	return namePrefix + strconv.Itoa(level)
}

func NameLevel(name string) int {
	for k, v := range levels {
		if v == name {
			return k
		}
	}
	var level int
	if strings.HasPrefix(name, namePrefix) {
		level, _ = strconv.Atoi(name[len(namePrefix):])
	}
	return level
}

type Logger struct {
	sync.Mutex
	level   int
	logger  *log.Logger
	logFile *os.File
}

func New(out io.Writer, prefix string, flag, level int, file *os.File) *Logger {
	return &Logger{
		level:   level,
		logger:  log.New(out, prefix, flag),
		logFile: file,
	}
}

func (l *Logger) output(level int, s string) error {
	// FIXME enable print GID for all log, should be disable as it effect performance
	if (level == 0) || (level == 1) || (level == 2) || (level == 3) {
		gid := common.GetGID()
		gidStr := strconv.FormatUint(gid, 10)

		// Get file information only
		pc := make([]uintptr, 10)
		runtime.Callers(2, pc)
		f := runtime.FuncForPC(pc[0])
		file, line := f.FileLine(pc[0])
		fileName := filepath.Base(file)
		lineStr := strconv.FormatUint(uint64(line), 10)
		return l.logger.Output(callDepth, AddBracket(LevelName(level))+" "+"GID"+
			" "+gidStr+", "+s+" "+fileName+":"+lineStr)
	} else {
		return l.logger.Output(callDepth, AddBracket(LevelName(level))+" "+s)
	}
}

func (l *Logger) Output(level int, a ...interface{}) error {
	if level >= l.level {
		return l.output(level, fmt.Sprintln(a...))
	}
	return nil
}

func (l *Logger) Trace(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(traceLog, a...)
}

func (l *Logger) Debug(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(debugLog, a...)
}

func (l *Logger) Info(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(infoLog, a...)
}

func (l *Logger) Warn(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(warnLog, a...)
}

func (l *Logger) Error(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(errorLog, a...)
}

func (l *Logger) Fatal(a ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Output(fatalLog, a...)
}

func Trace(a ...interface{}) {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)

	nameFull := f.Name()
	nameEnd := filepath.Ext(nameFull)
	funcName := strings.TrimPrefix(nameEnd, ".")
	Log.Trace(fmt.Sprint(funcName, "() ", fileName, ":", line, " ", fmt.Sprint(a...)))
}

func Debug(a ...interface{}) {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)
	Log.Debug(fmt.Sprint(f.Name(), " ", fileName, ":", line, " ", fmt.Sprint(a...)))
}

func Info(a ...interface{}) {
	Log.Info(fmt.Sprint(a...))
}

func Warn(a ...interface{}) {
	Log.Warn(fmt.Sprint(a...))
}

func Error(a ...interface{}) {
	Log.Error(fmt.Sprint(a...))
}

func Fatal(a ...interface{}) {
	Log.Fatal(fmt.Sprint(a...))
}

func FileOpen(path string) (*os.File, error) {
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("open %s: not a directory", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0766); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var currenttime string = time.Now().Format("2006-01-02_15.04.05")

	logfile, err := os.OpenFile(path+currenttime+"_LOG.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return logfile, nil
}

func Init(a ...interface{}) {
	writers := []io.Writer{}
	var logFile *os.File
	var err error
	if len(a) == 0 {
		writers = append(writers, ioutil.Discard)
	} else {
		for _, o := range a {
			switch o.(type) {
			case string:
				logFile, err = FileOpen(o.(string))
				if err != nil {
					fmt.Println("error: open log file failed")
					os.Exit(1)
				}
				writers = append(writers, logFile)
			case *os.File:
				writers = append(writers, o.(*os.File))
			default:
				fmt.Println("error: invalid log location")
				os.Exit(1)
			}
		}
	}
	fileAndStdoutWrite := io.MultiWriter(writers...)
	var printlevel int = config.Parameters.PrintLevel
	Log = New(fileAndStdoutWrite, "", log.Lmicroseconds, printlevel, logFile)
}

func GetLogFileSize() (int64, error) {
	f, e := Log.logFile.Stat()
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}

func GetMaxLogChangeInterval() int64 {
	if config.Parameters.MaxLogSize != 0 {
		return (config.Parameters.MaxLogSize * byteToMb)
	} else {
		return (defaultMaxLogSize * byteToMb)
	}
}

func CheckIfNeedNewFile() bool {
	logFileSize, err := GetLogFileSize()
	maxLogFileSize := GetMaxLogChangeInterval()
	if err != nil {
		return false
	}
	if logFileSize > maxLogFileSize {
		return true
	} else {
		return false
	}
}

func ClosePrintLog() error {
	var err error
	if Log.logFile != nil {
		err = Log.logFile.Close()
	}
	return err
}
