package lounge

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Log is a highly opinionated, three level log
// suitable for use in production
type Log interface {
	// With is used to attach key value pairs to
	// a Log for later usage
	With(pairs map[string]string) Log

	// Debug is for local development only
	// extremely detailed logs for debugging complex systems
	Debugf(fmt string, args ...interface{})

	// Infof is for normal log statements, nothing unexpected
	// should rarely be written to in production code
	// this is best used for information such as
	// - startup lines
	// - migration information
	// - system state changes
	// - background job notifs (if not tracked elsewhere)
	Infof(fmt string, args ...interface{})

	// Errorf is for unexpected errors
	// in production code. Errorf should be shippable to
	// an error aggregation service
	Errorf(fmt string, args ...interface{})
}

type Option func(l *DefaultLog)

func WithDebugEnabled() Option {
	return func(l *DefaultLog) {
		l.enableDebug = true
	}
}

func WithOutput(w io.Writer) Option {
	return func(l *DefaultLog) {
		l.output = w
	}
}

// NewDefaultLog returns a DefaultLog configured
// according to the options provided
func NewDefaultLog(opts ...Option) Log {
	dl := &DefaultLog{
		output: bufio.NewWriter(os.Stdout),
		pairs:  make(map[string]string),
	}

	for _, opt := range opts {
		opt(dl)
	}

	return dl
}

// DefaultLog is the core lounge Log implementation
type DefaultLog struct {
	pairs map[string]string

	enableDebug bool

	output io.Writer
}

func (dl *DefaultLog) With(pairs map[string]string) Log {
	newPairs := dl.pairs
	for k, v := range pairs {
		newPairs[k] = v
	}

	dl.pairs = newPairs

	return dl
}

func (dl *DefaultLog) Debugf(fmtStr string, args ...interface{}) {
	if !dl.enableDebug {
		return
	}

	dl.printLevel("DEBUG", fmtStr, args...)
}

func (dl *DefaultLog) Infof(fmtStr string, args ...interface{}) {
	dl.printLevel("INFO", fmtStr, args...)
}

func (dl *DefaultLog) Errorf(fmtStr string, args ...interface{}) {
	dl.printLevel("ERROR", fmtStr, args...)
}

func (dl *DefaultLog) printLevel(level string, fmtStr string, args ...interface{}) {
	currentTime := time.Now().In(time.UTC).Format(time.RFC3339)

	var pairs []string
	for k, v := range dl.pairs {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}

	if dl.enableDebug {
		caller := getFrame(2)
		callerWithLine := caller.File + "#" + strconv.Itoa(caller.Line)

		gopath, ok := os.LookupEnv("GOPATH")
		if ok {
			// remove gopath from log lines
			callerWithLine = strings.ReplaceAll(callerWithLine, gopath+"/src/", "")
		}
		fmt.Fprintf(dl.output, currentTime+" |"+level+"| "+callerWithLine+" "+strings.Join(pairs, " ")+fmtStr+"\n", args...)

	} else {
		fmt.Fprintf(dl.output, currentTime+" |"+level+"| "+strings.Join(pairs, " ")+fmtStr+"\n", args...)
	}
}

// https://stackoverflow.com/questions/35212985/is-it-possible-get-information-about-caller-function-in-golang
func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
