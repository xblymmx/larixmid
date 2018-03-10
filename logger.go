package larixmid

import (
	"net/http"
	"time"
	"text/template"
	"log"
	"os"
	"bytes"
)

type LoggerEntry struct {
	StartTime string
	Status int
	Duration time.Duration
	Hostname string
	Method string
	Path string
	Request *http.Request
}


var LoggerDefaultFormat = "{{.StartTime}} | {{.Status}} | \t {{.Duration}} | {{.Hostname}} | {{.Method}} {{.Path}} \n"

var LoggerDefaultDateFormat = time.RFC3339


type ALogger interface {
	Println(...interface{})
	Printf(format string, v ...interface{})
}

type Logger struct {
	ALogger
	dateFormat string
	template *template.Template
}


func NewLogger() *Logger {
	logger := &Logger{
		ALogger: log.New(os.Stdout, "[Larixmid]", 0),
		dateFormat: LoggerDefaultDateFormat,
	}
	logger.SetFormat(LoggerDefaultFormat)
	return logger
}

func (logger *Logger) SetFormat(format string) {
	logger.template = template.Must(template.New("larixmid_parser").Parse(format))
}

func (logger *Logger) SetDateFormat(format string) {
	logger.dateFormat = format
}

func (logger *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(w, r)

	res := w.(ResponseWriter)
	log := LoggerEntry{
		StartTime: start.Format(logger.dateFormat),
		Status: res.Status(),
		Duration: time.Since(start),
		Method: r.Method,
		Path: r.URL.Path,
		Request: r,
		Hostname: r.Host,
	}

	buf := &bytes.Buffer{}
	logger.template.Execute(buf, log)
	logger.Printf(buf.String())

}

