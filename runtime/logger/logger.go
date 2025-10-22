package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

type Logger struct {
	mu          sync.Mutex
	currentFile *os.File
	logDir      string
	maxSize     int
	currentDate string
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
	panicLogger *log.Logger
	level       LogLevel
}

func NewLogger(dirName string, maxSizeMB int, logLevel LogLevel) (*Logger, error) {
	// создаём директорию если нету
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return nil, err
	}
	// записываем в структуру логгера полученные данные
	l := &Logger{
		logDir:  dirName,
		maxSize: maxSizeMB,
		level:   logLevel,
	}
	// создаём файл если нету или создаём файл если больше дефолтного значения памяти файла
	if err := l.createNewLogFileIfNeed(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Logger) getCurrentLogFile() string {
	currentDate := time.Now().Format("2006-01-02")
	return filepath.Join(l.logDir, fmt.Sprintf("app-%s.log", currentDate))
}

func (l *Logger) createNewLogFile(filePath, date string) error {
	// если файл есть, закроем
	if l.currentFile != nil {
		l.currentFile.Close()
	}
	// создаём новый
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	l.currentFile = file
	l.currentDate = date

	flags := log.Ldate | log.Ltime | log.Lshortfile
	l.debugLogger = log.New(file, "DEBUG: ", flags)
	l.infoLogger = log.New(file, "INFO: ", flags)
	l.warnLogger = log.New(file, "WARN: ", flags)
	l.errorLogger = log.New(file, "ERROR: ", flags)
	l.fatalLogger = log.New(file, "FATAL: ", flags)
	l.panicLogger = log.New(file, "PANIC: ", flags)

	return nil
}

func (l *Logger) shouldRotateBySize() bool {
	if l.currentFile == nil {
		return true
	}
	info, err := l.currentFile.Stat()
	if err != nil {
		return true
	}
	return info.Size() >= int64(l.maxSize)*1024*1024
}

// создаём нвоый файл с суффикосом
func (l *Logger) rotateBySize() error {
	//убераем суффикс
	basePath := l.getCurrentLogFile()
	baseName := strings.TrimSuffix(basePath, ".log")

	// Ищем следующий доступный суффикс (1,2,3,4 и т.п.) и добавляем суффикс
	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s-%d.log", baseName, i)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return l.createNewLogFile(newPath, l.currentDate)
		}
	}
}

func (l *Logger) createNewLogFileIfNeed() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	currentDate := time.Now().Format("2006-01-02")

	logFile := l.getCurrentLogFile()

	// если дата не текущая или нету самого файла, то создаём его
	if l.currentDate != currentDate || l.currentFile == nil {
		return l.createNewLogFile(logFile, currentDate)
	}

	// проверяем размер файла
	if l.shouldRotateBySize() {
		return l.rotateBySize()
	}
	return nil
}

func (l *Logger) logInternal(level LogLevel, service, message string, err error) {
	if level < l.level {
		return
	}
	if err := l.createNewLogFileIfNeed(); err != nil {
		log.Printf("Failed to rotate log: %v", err)
	}
	fullMessage := fmt.Sprintf("[%s] %s", service, message)
	if err != nil {
		fullMessage += fmt.Sprintf(": %v", err)
	}

	switch level {
	case DEBUG:
		l.debugLogger.Output(3, fullMessage)
	case INFO:
		l.infoLogger.Output(3, fullMessage)
	case WARN:
		l.warnLogger.Output(3, fullMessage)
	case ERROR:
		l.errorLogger.Output(3, fullMessage)
	case FATAL:
		l.fatalLogger.Output(3, fullMessage)
		l.Close()
		os.Exit(1)
	case PANIC:
		l.panicLogger.Output(3, fullMessage)
		l.Close()
		panic(fullMessage)
	}
}

// Публичные методы
func (l *Logger) Debug(service, message string) {
	l.logInternal(DEBUG, service, message, nil)
}

func (l *Logger) Info(service, message string) {
	l.logInternal(INFO, service, message, nil)
}

func (l *Logger) Warn(service, message string) {
	l.logInternal(WARN, service, message, nil)
}

func (l *Logger) Error(service, message string, err error) {
	l.logInternal(ERROR, service, message, err)
}

func (l *Logger) Fatal(service, message string, err error) {
	l.logInternal(FATAL, service, message, err)
}

func (l *Logger) Panic(service, message string, err error) {
	l.logInternal(PANIC, service, message, err)
}

func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.currentFile != nil {
		l.currentFile.Close()
		l.currentFile = nil
	}
}

// SetLevel изменяет уровень логирования на лету
func (l *Logger) SetLevel(newLevel LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = newLevel
}

// GetLevel возвращает текущий уровень логирования
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// String representation для уровней логирования
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	case PANIC:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// Вспомогательные функции для создания логгера с настройками по умолчанию
func NewDefaultLogger() (*Logger, error) {
	return NewLogger("./runtime/logs", 10, INFO)
}

func NewProductionLogger() (*Logger, error) {
	return NewLogger("./runtime/logs", 100, WARN)
}
