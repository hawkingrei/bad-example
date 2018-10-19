package log

import (
	"context"
	"math"
	"path"
	"time"

	"github.com/alecthomas/log4go"
)

// FileHandler .
type FileHandler struct {
	writers []*log4go.FileLogWriter
	logger  log4go.Logger
	render  Render
}

// NewFile crete a file logger.
func NewFile(dir string) *FileHandler {
	format := "%M"
	var l = log4go.Logger{}
	log4go.LogBufferLength = 10240

	// new info writer
	iw := log4go.NewFileLogWriter(path.Join(dir, "info.log"), true)
	iw.SetRotateDaily(true)
	iw.SetRotateSize(math.MaxInt32)
	iw.SetFormat(format)
	l["info"] = &log4go.Filter{
		Level:     log4go.INFO,
		LogWriter: iw,
	}
	// new warning writer
	ww := log4go.NewFileLogWriter(path.Join(dir, "warning.log"), true)
	ww.SetRotateDaily(true)
	ww.SetRotateSize(math.MaxInt32)
	ww.SetFormat(format)
	l["warning"] = &log4go.Filter{
		Level:     log4go.WARNING,
		LogWriter: ww,
	}
	// new error writer
	ew := log4go.NewFileLogWriter(path.Join(dir, "error.log"), true)
	ew.SetRotateDaily(true)
	ew.SetRotateSize(math.MaxInt32)
	ew.SetFormat(format)
	l["error"] = &log4go.Filter{
		Level:     log4go.ERROR,
		LogWriter: ew,
	}
	return &FileHandler{
		logger:  l,
		writers: []*log4go.FileLogWriter{iw, ww, ew},
		render:  newPatternRender("[%D %T] [%L] [%S] %M"),
	}
}

// Log loggint to file .
func (h *FileHandler) Log(ctx context.Context, lv Level, args ...D) {
	d := make(map[string]interface{}, 10+len(args))
	for _, arg := range args {
		d[arg.Key] = arg.Value
	}
	// add extra fields
	addExtraField(ctx, d)
	fn := funcName()
	d[_levelValue] = lv
	d[_level] = lv.String()
	d[_time] = time.Now().Format(_timeFormat)
	d[_source] = fn
	c.filter(d)
	errIncr(lv, fn)
	msg := h.render.RenderString(d)
	switch lv {
	case _debugLevel:
		h.logger.Log(log4go.DEBUG, fn, msg)
	case _infoLevel:
		h.logger.Log(log4go.INFO, fn, msg)
	case _warnLevel:
		h.logger.Log(log4go.WARNING, fn, msg)
	case _errorLevel:
		h.logger.Log(log4go.ERROR, fn, msg)
	case _fatalLevel:
		h.logger.Log(log4go.CRITICAL, fn, msg)
	default:
	}
}

// Close .
func (h *FileHandler) Close() (err error) {
	h.logger.Close()
	return nil
}

// SetFormat .
func (h *FileHandler) SetFormat(format string) {
	h.render = newPatternRender(format)
}
