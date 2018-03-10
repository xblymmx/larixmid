package larixmid

import (
	"net/http"
	"fmt"
	"html/template"
	"log"
	"os"
	"runtime"
	"runtime/debug"
)

const (
	nilRequestMessage = "Request is nil"
	panicTextFormat   = "PANIC: %s\n%s"
	panicHTML         = `
<body>
<h1>Panic</h1>

<div>
<h3> {{.RequestDescription}} </h3>
<h4> {{.RecoveredPanic}} </h4>

{{ if .Stack }}
<div>
<h3> Runtime stack</h3>
<h4> {{.StackToString}} </h4>
</div>
</body>
`
)

var panicHTMLTemplate = template.Must(template.New("PanicPage").Parse(panicHTML))

type PanicInformation struct {
	RecoveredPanic interface{}
	Stack          []byte
	Request        *http.Request
}

func (p *PanicInformation) StackToString() string {
	return string(p.Stack)
}

func (p *PanicInformation) RequestDescription() string {
	if p.Request == nil {
		return nilRequestMessage
	}

	queryStr := ""
	if p.Request.URL.RawQuery != "" {
		queryStr = "?" + p.Request.URL.RawQuery
	}

	return fmt.Sprintf("%s %s%s", p.Request.Method, p.Request.URL.Path, queryStr)
}

type PanicFormatter interface {
	FormatPanicError(w http.ResponseWriter, r *http.Request, p *PanicInformation)
}

type TextPanicFormatter struct{}

func (t *TextPanicFormatter) FormatPanicError(w http.ResponseWriter, r *http.Request, p *PanicInformation) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf8")
	}

	fmt.Fprintf(w, panicTextFormat, p.RecoveredPanic, p.Stack)
}

type HTMLPanicFormatter struct{}

func (h *HTMLPanicFormatter) FormatPanicError(w http.ResponseWriter, r *http.Request, p *PanicInformation) {
	if w.Header().Get("Content-Type") != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf8")
	}

	panicHTMLTemplate.Execute(w, p)
}

type Recovery struct {
	Logger           ALogger
	PrintStack       bool
	PanicHandlerFunc func(*PanicInformation)
	StackAll         bool
	StackSize        int
	Formatter        PanicFormatter
}

func (rec *Recovery) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			stack := make([]byte, rec.StackSize)
			stack = stack[:runtime.Stack(stack, rec.StackAll)]
			infos := &PanicInformation{RecoveredPanic: err, Request: r}

			if rec.PrintStack {
				infos.Stack = stack
			}

			rec.Logger.Printf(panicTextFormat, err, stack)
			rec.Formatter.FormatPanicError(w, r, infos)

			if rec.PanicHandlerFunc != nil {
				func() {
					defer func() {
						if err := recover(); err != nil {
							rec.Logger.Printf("Provided PanicHandlerFunc panic-ed %s, trace:\n %s", err, debug.Stack())
							rec.Logger.Printf("%s\n", debug.Stack())
						}
						rec.PanicHandlerFunc(infos)
					}()
				}()
			}
		}
	}()

	next(w, r)
}

func NewRecovery() *Recovery {
	return &Recovery{
		Logger:     log.New(os.Stdout, "[Larixmid] ", 0),
		PrintStack: true,
		StackAll:   false,
		StackSize:  1024 * 8,
		Formatter:  &TextPanicFormatter{},
	}
}
