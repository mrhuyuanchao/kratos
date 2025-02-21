package log

import (
	"context"
	"os"
	"time"
)

type StdoutJsonHandler struct {
	render Render
}

func NewJsonStdout() *StdoutJsonHandler {
	return &StdoutJsonHandler{render: newPatternRender(defaultPattern, RenderWithJson(true))}
}

func (j *StdoutJsonHandler) Log(ctx context.Context, lv Level, args ...D) {
	d := toMap(args...)
	addExtraField(ctx, d)
	d[_time] = time.Now().Format("2006-01-02 15:04:05.999")
	j.render.Render(os.Stderr, d)
	os.Stderr.Write([]byte{'\n'})
}

func (j *StdoutJsonHandler) Close() error {
	return nil
}

func (j *StdoutJsonHandler) SetFormat(format string) {
	j.render = newPatternRender(format, RenderWithJson(true))
}
