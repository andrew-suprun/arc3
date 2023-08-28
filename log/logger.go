package log

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"time"
)

var state = make(chan *logState, 1)

type logState struct {
	loggerName string
	start      time.Time
	writer     *os.File
}

func SetLogger(fileName string) {
	lState := <-state
	defer storeState(lState)

	if lState.loggerName != "" {
		panic("log.SetLogger() already ran")
	}

	lState.loggerName = fileName
}

func CloseLogger() {
	lState := <-state
	defer storeState(lState)

	if lState.writer != nil {
		lState.writer.Close()
	}
}

func Write(b []byte) (int, error) {
	lState := <-state
	defer storeState(lState)

	return lState.writer.Write(b)
}

func Debug(msg string, params ...any) {
	lState := <-state
	defer storeState(lState)

	if lState.loggerName != "" && lState.writer == nil {
		var err any
		lState.writer, err = os.Create(lState.loggerName)
		if err != nil {
			panic(fmt.Sprintf("Failed to create log file %q: %v", lState.loggerName, err))
		}

		lState.start = time.Now()
	}
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%8s - %s", time.Since(lState.start).Truncate(time.Millisecond), msg)
	for i, param := range params {
		sep := ": "
		if i == 0 {
			sep = " - "
		} else if i%2 == 0 {
			sep = ", "
		}
		fmt.Fprint(buf, sep)
		printValue(buf, param)
	}
	fmt.Fprintln(buf)
	_, err := lState.writer.Write(buf.Bytes())
	if err != nil {
		panic(fmt.Sprintf("Failed to write to log file: %v", err))
	}
}

func printValue(buf *bytes.Buffer, value any) {
	if str, ok := value.(fmt.Stringer); ok {
		fmt.Fprint(buf, str)
		return
	}
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.String:
		fmt.Fprintf(buf, "%q", v.String())

	case reflect.Bool:
		fmt.Fprintf(buf, "%v", v.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(buf, v.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Fprint(buf, v.Uint())

	case reflect.Slice:
		if v, ok := value.([]byte); ok {
			fmt.Fprint(buf, string(v))
		} else {
			fmt.Fprintf(buf, "%v", value)

		}

	default:
		fmt.Fprintf(buf, "%v", value)
	}
}

func storeState(lState *logState) {
	state <- lState
}

func init() {
	state <- &logState{}
}
