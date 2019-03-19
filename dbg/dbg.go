// This file was AUTOMATICALLY GENERATED by dbg-import (smuggol) from github.com/robfig/dbg

/*
Package dbg is a println/printf/log-debugging utility library.

    import (
        Dbg "github.com/robfig/dbg"
    )

    dbg, dbgf := Dbg.New()

    dbg("Emit some debug stuff", []byte{120, 121, 122, 122, 121}, math.Pi)
    # "2013/01/28 16:50:03 Emit some debug stuff [120 121 122 122 121] 3.141592653589793"

    dbgf("With a %s formatting %.2f", "little", math.Pi)
    # "2013/01/28 16:51:55 With a little formatting (3.14)"

    dbgf("%/fatal//A fatal debug statement: should not be here")
    # "A fatal debug statement: should not be here"
    # ...and then, os.Exit(1)

    dbgf("%/panic//Can also panic %s", "this")
    # "Can also panic this"
    # ...as a panic, equivalent to: panic("Can also panic this")

    dbgf("Any %s arguments without a corresponding %%", "extra", "are treated like arguments to dbg()")
    # "2013/01/28 17:14:40 Any extra arguments (without a corresponding %) are treated like arguments to dbg()"

    dbgf("%d %d", 1, 2, 3, 4, 5)
    # "2013/01/28 17:16:32 Another example: 1 2 3 4 5"

    dbgf("%@: Include the function name for a little context (via %s)", "%@")
    # "2013... github.com/robfig/dbg.TestSynopsis: Include the function name for a little context (via %@)"

By default, dbg uses log (log.Println, log.Printf, log.Panic, etc.) for output.
However, you can also provide your own output destination by invoking dbg.New with
a customization function:

    import (
        "bytes"
        Dbg "github.com/robfig/dbg"
        "os"
    )

    # dbg to os.Stderr
    dbg, dbgf := Dbg.New(func(dbgr *Dbgr) {
        dbgr.SetOutput(os.Stderr)
    })

    # A slightly contrived example:
    var buffer bytes.Buffer
    dbg, dbgf := New(func(dbgr *Dbgr) {
        dbgr.SetOutput(&buffer)
    })

*/
package dbg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"unicode"
)

type _frmt struct {
	ctl          string
	format       string
	operandCount int
	panic        bool
	fatal        bool
	check        bool
}

var (
	ctlTest = regexp.MustCompile(`^\s*%/`)
	ctlScan = regexp.MustCompile(`%?/(panic|fatal|check)(?:\s|$)`)
)

func operandCount(format string) int {
	count := 0
	end := len(format)
	for at := 0; at < end; {
		for at < end && format[at] != '%' {
			at++
		}
		at++
		if at < end {
			if format[at] != '%' && format[at] != '@' {
				count++
			}
			at++
		}
	}
	return count
}

func parseFormat(format string) (frmt _frmt) {
	if ctlTest.MatchString(format) {
		format = strings.TrimLeftFunc(format, unicode.IsSpace)
		index := strings.Index(format, "//")
		if index != -1 {
			frmt.ctl = format[0:index]
			format = format[index+2:] // Skip the second slash via +2 (instead of +1)
		} else {
			frmt.ctl = format
			format = ""
		}
		for _, tmp := range ctlScan.FindAllStringSubmatch(frmt.ctl, -1) {
			for _, value := range tmp[1:] {
				switch value {
				case "panic":
					frmt.panic = true
				case "fatal":
					frmt.fatal = true
				case "check":
					frmt.check = true
				}
			}
		}
	}
	frmt.format = format
	frmt.operandCount = operandCount(format)
	return
}

type Dbgr struct {
	emit _emit
}

type DbgFunction func(values ...interface{})

func NewDbgr() *Dbgr {
	self := &Dbgr{}
	return self
}

/*
New will create and return a pair of debugging functions. You can customize where
they output to by passing in an (optional) customization function:

    import (
        Dbg "github.com/robfig/dbg"
        "os"
    )

    # dbg to os.Stderr
    dbg, dbgf := Dbg.New(func(dbgr *Dbgr) {
        dbgr.SetOutput(os.Stderr)
    })

*/
func New(options ...interface{}) (dbg DbgFunction, dbgf DbgFunction) {
	dbgr := NewDbgr()
	if len(options) > 0 {
		if fn, ok := options[0].(func(*Dbgr)); ok {
			fn(dbgr)
		}
	}
	return dbgr.DbgDbgf()
}

func (self Dbgr) Dbg(values ...interface{}) {
	self.getEmit().emit(_frmt{}, "", values...)
}

func (self Dbgr) Dbgf(values ...interface{}) {
	self.dbgf(values...)
}

func (self Dbgr) DbgDbgf() (dbg DbgFunction, dbgf DbgFunction) {
	dbg = func(vl ...interface{}) {
		self.Dbg(vl...)
	}
	dbgf = func(vl ...interface{}) {
		self.dbgf(vl...)
	}
	return dbg, dbgf // Redundant, but...
}

func (self Dbgr) dbgf(values ...interface{}) {

	var frmt _frmt
	if len(values) > 0 {
		tmp := fmt.Sprint(values[0])
		frmt = parseFormat(tmp)
		values = values[1:]
	}

	buffer_f := bytes.Buffer{}
	format := frmt.format
	end := len(format)
	for at := 0; at < end; {
		last := at
		for at < end && format[at] != '%' {
			at++
		}
		if at > last {
			buffer_f.WriteString(format[last:at])
		}
		if at >= end {
			break
		}
		// format[at] == '%'
		at++
		// format[at] == ?
		if format[at] == '@' {
			depth := 2
			pc, _, _, _ := runtime.Caller(depth)
			name := runtime.FuncForPC(pc).Name()
			buffer_f.WriteString(name)
		} else {
			buffer_f.WriteString(format[at-1 : at+1])
		}
		at++
	}

	//values_f := append([]interface{}{}, values[0:frmt.operandCount]...)
	values_f := values[0:frmt.operandCount]
	values_dbg := values[frmt.operandCount:]
	if len(values_dbg) > 0 {
		// Adjust frmt.format:
		// (%v instead of %s because: frmt.check)
		{
			tmp := format
			if len(tmp) > 0 {
				if unicode.IsSpace(rune(tmp[len(tmp)-1])) {
					buffer_f.WriteString("%v")
				} else {
					buffer_f.WriteString(" %v")
				}
			} else if frmt.check {
				// Performing a check, so no output
			} else {
				buffer_f.WriteString("%v")
			}
		}

		// Adjust values_f:
		if !frmt.check {
			tmp := []string{}
			for _, value := range values_dbg {
				tmp = append(tmp, fmt.Sprintf("%v", value))
			}
			// First, make a copy of values_f, so we avoid overwriting values_dbg when appending
			values_f = append([]interface{}{}, values_f...)
			values_f = append(values_f, strings.Join(tmp, " "))
		}
	}

	format = buffer_f.String()
	if frmt.check {
		// We do not actually emit to the log, but panic if
		// a non-nil value is detected (e.g. a non-nil error)
		for _, value := range values_dbg {
			if value != nil {
				if format == "" {
					panic(value)
				} else {
					panic(fmt.Sprintf(format, append(values_f, value)...))
				}
			}
		}
	} else {
		self.getEmit().emit(frmt, format, values_f...)
	}
}

// Idiot-proof &Dbgr{}, etc.
func (self *Dbgr) getEmit() _emit {
	if self.emit == nil {
		self.emit = standardEmit()
	}
	return self.emit
}

// SetOutput will accept the following as a destination for output:
//
//      *log.Logger         Print*/Panic*/Fatal* of the logger
//      io.Writer           -
//      nil                 Reset to the default output (os.Stderr)
//      "log"               Print*/Panic*/Fatal* via the "log" package
//
func (self *Dbgr) SetOutput(output interface{}) {
	if output == nil {
		self.emit = standardEmit()
		return
	}
	switch output := output.(type) {
	case *log.Logger:
		self.emit = _emitLogger{
			logger: output,
		}
		return
	case io.Writer:
		self.emit = _emitWriter{
			writer: output,
		}
		return
	case string:
		if output == "log" {
			self.emit = _emitLog{}
			return
		}
	}
	panic(output)
}

// ======== //
// = emit = //
// ======== //

func standardEmit() _emit {
	return _emitWriter{
		writer: os.Stderr,
	}
}

func ln(tmp string) string {
	length := len(tmp)
	if length > 0 && tmp[length-1] != '\n' {
		return tmp + "\n"
	}
	return tmp
}

type _emit interface {
	emit(_frmt, string, ...interface{})
}

type _emitWriter struct {
	writer io.Writer
}

func (self _emitWriter) emit(frmt _frmt, format string, values ...interface{}) {
	if format == "" {
		fmt.Fprintln(self.writer, values...)
	} else {
		if frmt.panic {
			panic(fmt.Sprintf(format, values...))
		}
		fmt.Fprintf(self.writer, ln(format), values...)
		if frmt.fatal {
			os.Exit(1)
		}
	}
}

type _emitLogger struct {
	logger *log.Logger
}

func (self _emitLogger) emit(frmt _frmt, format string, values ...interface{}) {
	if format == "" {
		self.logger.Println(values...)
	} else {
		if frmt.panic {
			self.logger.Panicf(format, values...)
		} else if frmt.fatal {
			self.logger.Fatalf(format, values...)
		} else {
			self.logger.Printf(format, values...)
		}
	}
}

type _emitLog struct {
}

func (self _emitLog) emit(frmt _frmt, format string, values ...interface{}) {
	if format == "" {
		log.Println(values...)
	} else {
		if frmt.panic {
			log.Panicf(format, values...)
		} else if frmt.fatal {
			log.Fatalf(format, values...)
		} else {
			log.Printf(format, values...)
		}
	}
}
