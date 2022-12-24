package errco

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

// DebugLvl specify the level of debugging
// (default is LVL_4 so it will log everything)
var DebugLvl LogLvl = LVL_4

type MshLog struct {
	Ori LogOri        // log origin function
	Typ LogTyp        // log type
	Lvl LogLvl        // log debug level
	Cod LogCod        // log code
	Mex string        // log string
	Arg []interface{} // log args
}

type LogOri string
type LogTyp string
type LogLvl int
type LogCod int

const (
	COLOR_RESET  = "\033[0m"
	COLOR_GRAY   = "\033[30m"
	COLOR_RED    = "\033[31m"
	COLOR_GREEN  = "\033[32m"
	COLOR_YELLOW = "\033[33m"
	COLOR_BLUE   = "\033[34m"
	COLOR_PURPLE = "\033[35m"
	COLOR_CYAN   = "\033[36m"
)

// NewLog returns a new msh log object.
//
// When a function fails and returns msh log using NewLog, msh log type must be TYPE_ERR or TYPE_WAR.
// Find bad usage with reg exp: `return (.*)NewLog\((.*)TYPE_(?!ERR|WAR)`
//
// To create a msh log and print it immediately you must use NewLogln()
// If you really want to use NewLog(), use NewLog().Log(false)
// Find bad usage with reg exp: `NewLog\((.*)\).Log\(true`
func NewLog(t LogTyp, l LogLvl, c LogCod, m string, a ...interface{}) *MshLog {
	logMsh := &MshLog{Trace(2), t, l, c, m, a}
	return logMsh
}

// NewLogln prints to terminal msh log struct and returns a new msh log struct.
//
// When a function fails it should not return a msh log struct using NewLogln
// (there is the risk of printing 2 times the same msh log)
// the parent function should handle the logging of msh log struct
// Find bad usage with reg exp: `return (.*)NewLogln\(`
func NewLogln(t LogTyp, l LogLvl, c LogCod, m string, a ...interface{}) *MshLog {
	logMsh := &MshLog{Trace(2), t, l, c, m, a}
	// trace was just set, no need to set it again
	// it would also be wrong:
	// 1) example()               -> Log() -> trace(2) : example
	// 2) example() -> NewLogln() -> trace(2)          : example
	//                            \> Log() -> trace(2) : NewLogln (!)
	logMsh.Log(false)
	return logMsh
}

// Log prints to terminal msh log struct.
//
// if tracing is set to true, Log() will add the caller function to the msh log trace
//
// returns the original log for convenience.
// returns nil if msh log struct is nil.
func (logMsh *MshLog) Log(tracing bool) *MshLog {
	// return immediately if original log is nil
	if logMsh == nil {
		return nil
	}

	// ------- operations on original log -------

	// add trace if requested
	if tracing {
		logMsh.Ori = Trace(2) + LogOri(": ") + logMsh.Ori
	}

	// return original log if log level is not high enough
	if logMsh.Lvl > DebugLvl {
		return logMsh
	}

	// make a copy of original log
	logMod := *logMsh

	// -------- operations on copied log --------

	// set logMod colors depending on logMod level
	switch logMod.Lvl {
	case LVL_0:
		// make important logs more visible
		logMod.Mex = COLOR_CYAN + logMod.Mex + COLOR_RESET
	}

	// set logMod colors depending on logMod type
	var t string
	switch logMod.Typ {
	case TYPE_INF:
		t = COLOR_BLUE + string(logMod.Typ) + COLOR_RESET
	case TYPE_SER:
		t = COLOR_GRAY + string(logMod.Typ) + COLOR_RESET
		logMod.Mex = COLOR_GRAY + logMod.Mex + "\x00" + COLOR_RESET
	case TYPE_BYT:
		t = COLOR_PURPLE + string(logMod.Typ) + COLOR_RESET
	case TYPE_WAR:
		t = COLOR_YELLOW + string(logMod.Typ) + COLOR_RESET
	case TYPE_ERR:
		t = COLOR_RED + string(logMod.Typ) + COLOR_RESET
	}

	// print logMod depending on logMod type
	switch logMod.Typ {
	case TYPE_INF, TYPE_SER, TYPE_BYT:
		log.Printf("[%-16s %-4s] %s\n",
			t,
			strings.Repeat("≡", 4-int(logMod.Lvl)),
			fmt.Sprintf(logMod.Mex, logMod.Arg...))
	case TYPE_WAR, TYPE_ERR:
		log.Printf("[%-16s %-4s] %s %s %s\n",
			t,
			strings.Repeat("≡", 4-int(logMod.Lvl)),
			LogOri(COLOR_YELLOW)+logMod.Ori+":"+LogOri(COLOR_RESET),
			fmt.Sprintf(logMod.Mex, logMod.Arg...),
			fmt.Sprintf("[%06x]", logMod.Cod))
	}

	// return original log
	return logMsh
}

// AddTrace adds the caller function to the msh log trace
func (log *MshLog) AddTrace() *MshLog {
	// return original log if it's nil
	if log == nil {
		return log
	}

	log.Ori = Trace(2) + LogOri(": ") + log.Ori

	return log
}

// Trace returns the parent^(skip) function name
//
// skip == 2: example() -> NewLog() -> trace(): example
func Trace(skip int) LogOri {
	var o string = "?"

	if pc, _, _, ok := runtime.Caller(skip); !ok {
	} else if f := runtime.FuncForPC(pc); f == nil {
	} else {
		fn := f.Name()
		o = fn[strings.LastIndex(fn, ".")+1:]
	}

	return LogOri(o)
}
