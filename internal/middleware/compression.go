package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

const (
	EncodingZstd = "zstd"
	EncodingBr   = "br"
	EncodingGzip = "gzip"
)

var (
	gzipWriterPool = sync.Pool{
		New: func() any {
			w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
			return w
		},
	}
	zstdWriterPool = sync.Pool{
		New: func() any {
			w, _ := zstd.NewWriter(io.Discard, zstd.WithEncoderLevel(zstd.SpeedDefault))
			return w
		},
	}
	brotliWriterPool = sync.Pool{
		New: func() any {
			return brotli.NewWriterLevel(io.Discard, brotli.DefaultCompression)
		},
	}
)

type CompressionConfig struct {
	Encodings []string
	MinBytes  int
}

func Compression(config CompressionConfig) gin.HandlerFunc {
	minBytes := config.MinBytes
	if minBytes < 0 {
		minBytes = 0
	}
	encodings := config.Encodings

	return func(c *gin.Context) {
		if c.Request.Header.Get("Upgrade") != "" {
			c.Next()
			return
		}

		encoding := selectEncoding(c.Request.Header.Get("Accept-Encoding"), encodings)
		if encoding == "" {
			c.Next()
			return
		}

		writer := &compressionWriter{
			ResponseWriter: c.Writer,
			encoding:       encoding,
			minBytes:       minBytes,
		}
		c.Writer = writer

		defer func() {
			writer.finish()
		}()

		c.Next()
	}
}

type compressionWriter struct {
	gin.ResponseWriter
	encoding    string
	minBytes    int
	buffer      bytes.Buffer
	compressor  io.WriteCloser
	compressErr error
	skipped     bool
	started     bool
}

func (w *compressionWriter) shouldSkip() bool {
	status := w.Status()
	if status < 200 || status == http.StatusNoContent || status == http.StatusNotModified {
		return true
	}
	if w.Header().Get("Content-Encoding") != "" || w.Header().Get("Content-Range") != "" {
		return true
	}
	contentType := w.Header().Get("Content-Type")
	if strings.HasPrefix(contentType, "text/event-stream") {
		return true
	}
	return false
}

func (w *compressionWriter) initCompressor() {
	switch w.encoding {
	case EncodingZstd:
		zw := zstdWriterPool.Get().(*zstd.Encoder)
		zw.Reset(w.ResponseWriter)
		w.compressor = &zstdPoolWriter{Encoder: zw}
	case EncodingBr:
		bw := brotliWriterPool.Get().(*brotli.Writer)
		bw.Reset(w.ResponseWriter)
		w.compressor = &brotliPoolWriter{Writer: bw}
	case EncodingGzip:
		gw := gzipWriterPool.Get().(*gzip.Writer)
		gw.Reset(w.ResponseWriter)
		w.compressor = &gzipPoolWriter{Writer: gw}
	}
}

func (w *compressionWriter) startCompression() {
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Encoding", w.encoding)
	addVary(w.Header(), "Accept-Encoding")
	w.initCompressor()
}

func (w *compressionWriter) Write(data []byte) (int, error) {
	if w.skipped {
		return w.ResponseWriter.Write(data)
	}

	if w.compressor != nil {
		return w.compressor.Write(data)
	}

	if w.shouldSkip() {
		w.skipped = true
		if w.buffer.Len() > 0 {
			if _, err := w.ResponseWriter.Write(w.buffer.Bytes()); err != nil {
				return 0, err
			}
			w.buffer.Reset()
		}
		return w.ResponseWriter.Write(data)
	}

	if w.buffer.Len()+len(data) < w.minBytes {
		return w.buffer.Write(data)
	}

	w.started = true
	w.startCompression()

	if w.buffer.Len() > 0 {
		if _, err := w.compressor.Write(w.buffer.Bytes()); err != nil {
			return 0, err
		}
		w.buffer.Reset()
	}

	return w.compressor.Write(data)
}

func (w *compressionWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *compressionWriter) finish() {
	if w.skipped {
		return
	}

	if w.compressor != nil {
		_ = w.compressor.Close()
		return
	}

	if w.shouldSkip() {
		if w.buffer.Len() > 0 {
			_, _ = w.ResponseWriter.Write(w.buffer.Bytes())
			w.buffer.Reset()
		}
		return
	}

	if w.buffer.Len() >= w.minBytes && !w.started {
		w.startCompression()
		_, _ = w.compressor.Write(w.buffer.Bytes())
		_ = w.compressor.Close()
		w.buffer.Reset()
		return
	}

	addVary(w.Header(), "Accept-Encoding")
	if w.buffer.Len() > 0 {
		_, _ = w.ResponseWriter.Write(w.buffer.Bytes())
		w.buffer.Reset()
	}
}

func (w *compressionWriter) Flush() {
	if w.compressor != nil {
		if flusher, ok := w.compressor.(interface{ Flush() error }); ok {
			_ = flusher.Flush()
		}
	}
	w.ResponseWriter.Flush()
}

type gzipPoolWriter struct {
	*gzip.Writer
}

func (g *gzipPoolWriter) Close() error {
	err := g.Writer.Close()
	gzipWriterPool.Put(g.Writer)
	return err
}

type zstdPoolWriter struct {
	*zstd.Encoder
}

func (z *zstdPoolWriter) Close() error {
	err := z.Encoder.Close()
	zstdWriterPool.Put(z.Encoder)
	return err
}

type brotliPoolWriter struct {
	*brotli.Writer
}

func (b *brotliPoolWriter) Close() error {
	err := b.Writer.Close()
	brotliWriterPool.Put(b.Writer)
	return err
}

func addVary(header http.Header, value string) {
	existing := header.Values("Vary")
	for _, v := range existing {
		for _, part := range strings.Split(v, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) || strings.TrimSpace(part) == "*" {
				return
			}
		}
	}
	header.Add("Vary", value)
}

func selectEncoding(acceptEncoding string, serverEncodings []string) string {
	if acceptEncoding == "" || len(serverEncodings) == 0 {
		return ""
	}

	type clientPref struct {
		encoding string
		q        float64
	}

	var prefs []clientPref
	hasWildcard := false
	wildcardQ := 0.0

	for _, clause := range strings.Split(acceptEncoding, ",") {
		clause = strings.TrimSpace(clause)
		if clause == "" {
			continue
		}
		parts := strings.Split(clause, ";")
		enc := strings.ToLower(strings.TrimSpace(parts[0]))
		q := 1.0
		for _, param := range parts[1:] {
			param = strings.TrimSpace(param)
			if strings.HasPrefix(param, "q=") {
				if parsedQ, err := strconv.ParseFloat(strings.TrimPrefix(param, "q="), 64); err == nil {
					q = parsedQ
				}
			}
		}
		if enc == "*" {
			hasWildcard = true
			wildcardQ = q
		} else {
			prefs = append(prefs, clientPref{encoding: enc, q: q})
		}
	}

	for _, serverEnc := range serverEncodings {
		normalized := strings.ToLower(serverEnc)
		matched := false
		var matchedQ float64
		for _, pref := range prefs {
			if pref.encoding == normalized {
				matched = true
				matchedQ = pref.q
				break
			}
		}
		if matched {
			if matchedQ > 0 {
				return normalized
			}
			continue
		}
		if hasWildcard && wildcardQ > 0 {
			return normalized
		}
	}

	return ""
}
