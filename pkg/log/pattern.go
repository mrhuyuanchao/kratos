package log

import (
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
)

var patternMap = map[string]func(map[string]interface{}) string{
	"T": longTime,
	"t": shortTime,
	"D": longDate,
	"d": shortDate,
	"L": keyFactory(_level),
	"f": keyFactory(_source),
	"i": keyFactory(_instanceID),
	"e": keyFactory(_deplyEnv),
	"z": keyFactory(_zone),
	"S": longSource,
	"s": shortSource,
	"M": message,
}

// newPatternRender new pattern render
func newPatternRender(format string, options ...patternOption) Render {
	p := &pattern{
		bufPool: sync.Pool{New: func() interface{} { return &strings.Builder{} }},
		isJson:  false,
	}
	for _, option := range options {
		option(p)
	}
	b := make([]byte, 0, len(format))
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b = append(b, format[i])
			continue
		}
		if i+1 >= len(format) {
			b = append(b, format[i])
			continue
		}
		f, ok := patternMap[string(format[i+1])]
		if !ok {
			b = append(b, format[i])
			continue
		}
		if len(b) != 0 {
			p.funcs = append(p.funcs, textFactory(string(b)))
			b = b[:0]
		}
		p.funcs = append(p.funcs, f)
		i++
	}
	if len(b) != 0 {
		p.funcs = append(p.funcs, textFactory(string(b)))
	}
	return p
}

type patternOption func(p *pattern)

func RenderWithJson(isJson bool) patternOption {
	return func(p *pattern) {
		p.isJson = isJson
	}
}

type pattern struct {
	funcs   []func(map[string]interface{}) string
	bufPool sync.Pool
	isJson  bool
}

// Render implements Formater
func (p *pattern) Render(w io.Writer, d map[string]interface{}) error {
	var data []byte
	if p.isJson {
		marshalData, marshalError := sonic.Marshal(d)
		if marshalError != nil {
			return marshalError
		}
		data = marshalData
	} else {

		builder := p.bufPool.Get().(*strings.Builder)
		defer func() {
			builder.Reset()
			p.bufPool.Put(builder)
		}()
		for _, f := range p.funcs {
			builder.WriteString(f(d))
		}
		data = []byte(builder.String())
	}

	_, err := w.Write(data)
	return err
}

// Render implements Formater as string
func (p *pattern) RenderString(d map[string]interface{}) string {
	if p.isJson {
		data, _ := sonic.MarshalString(d)
		return data
	}
	builder := p.bufPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		p.bufPool.Put(builder)
	}()
	for _, f := range p.funcs {
		builder.WriteString(f(d))
	}
	return builder.String()

}

func textFactory(text string) func(map[string]interface{}) string {
	return func(map[string]interface{}) string {
		return text
	}
}
func keyFactory(key string) func(map[string]interface{}) string {
	return func(d map[string]interface{}) string {
		if v, ok := d[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprint(v)
		}
		return ""
	}
}

func longSource(d map[string]interface{}) string {
	if fn, ok := d[_source].(string); ok {
		return fn
	}
	return "unknown:0"
}

func shortSource(d map[string]interface{}) string {
	if fn, ok := d[_source].(string); ok {
		return path.Base(fn)
	}
	return "unknown:0"
}

func longTime(map[string]interface{}) string {
	return time.Now().Format("15:04:05.000")
}

func shortTime(map[string]interface{}) string {
	return time.Now().Format("15:04")
}

func longDate(map[string]interface{}) string {
	return time.Now().Format("2006/01/02")
}

func shortDate(map[string]interface{}) string {
	return time.Now().Format("01/02")
}

func isInternalKey(k string) bool {
	switch k {
	case _level, _levelValue, _time, _source, _instanceID, _appID, _deplyEnv, _zone:
		return true
	}
	return false
}

func message(d map[string]interface{}) string {
	var m string
	var s []string
	for k, v := range d {
		if k == _log {
			m = fmt.Sprint(v)
			continue
		}
		if isInternalKey(k) {
			continue
		}
		s = append(s, fmt.Sprintf("%s=%v", k, v))
	}
	s = append(s, m)
	return strings.Join(s, " ")
}
