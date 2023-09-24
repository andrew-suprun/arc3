package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Type   string
	Params map[string]string
}

func (msg *Message) String() string {
	buf := &strings.Builder{}
	fmt.Fprint(buf, msg.Type)
	sep := ": "
	for name, value := range msg.Params {
		fmt.Fprintf(buf, "%s%q = %q", sep, name, value)
		sep = ", "
	}
	return buf.String()
}

func (msg *Message) StringValue(param string) string {
	return msg.Params[param]
}

func (msg *Message) Int(param string) int {
	res, err := strconv.ParseInt(msg.Params[param], 10, 64)
	if err != nil {
		panic(err)
	}
	return int(res)
}

func (msg *Message) Time(param string) time.Time {
	res, err := time.Parse(time.RFC3339, msg.Params[param])
	if err != nil {
		panic(err)
	}
	return res
}

func Parse(cmd string) *Message {
	res := &Message{
		Params: map[string]string{},
	}
	if cmd[len(cmd)-1] == '\n' {
		cmd = cmd[:len(cmd)-1]
	}
	parts := strings.Split(cmd, "\t")
	res.Type = parts[0]
	for _, param := range parts[1:] {
		pParts := strings.SplitN(param, "=", 2)
		if len(pParts) == 2 {
			res.Params[pParts[0]] = pParts[1]
		}
	}
	return res
}

func String(kind string, params ...any) string {
	buf := &strings.Builder{}
	buf.WriteString(kind)
	for i := 0; i < len(params)-1; i += 2 {
		buf.WriteRune('\t')
		printValue(buf, params[i])
		buf.WriteRune('=')
		printValue(buf, params[i+1])
	}
	buf.WriteRune('\n')
	return buf.String()
}

func printValue(buf *strings.Builder, value any) {
	switch v := value.(type) {
	case string:
		buf.WriteString(v)
	case []byte:
		buf.WriteString(string(v))
	case int:
		fmt.Fprintf(buf, "%d", v)
	case time.Time:
		buf.WriteString(v.Format(time.RFC3339))
	default:
		panic(fmt.Sprintf("Cannot print value %v of type %T", v, v))
	}
}
