package terraform

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DebugInfo is the global handler for writing the debug archive. All methods
// are safe to call concurrently. Setting DebugInfo to nil will disable writing
// the debug archive. All methods are safe to call in the nil value.
var DebugInfo *debugInfo

// SetDebugInfo sets the debug options for the terraform package. Currently
// this just sets the path where the archive will be written.
func SetDebugInfo(path string) error {
	if os.Getenv("TF_DEBUG") == "" {
		return nil
	}

	di, err := newDebugInfo(path)
	if err != nil {
		return err
	}

	DebugInfo = di
	return nil
}

func newDebugInfo(dir string) (*debugInfo, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	// FIXME: not guaranteed unique, but good enough for now
	name := fmt.Sprintf("debug-%s", time.Now().Format("2006-01-02-15-04-05.999999999"))
	archivePath := filepath.Join(dir, name)

	f, err := os.OpenFile(archivePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return nil, err
	}

	d := &debugInfo{
		Name:    name,
		file:    f,
		archive: zip.NewWriter(f),
	}
	return d, nil
}

type debugInfo struct {
	Name string
	sync.Mutex
	file    *os.File
	archive *zip.Writer
	step    int
	closed  bool
}

func (d *debugInfo) Close() error {
	if d == nil {
		return nil
	}

	d.Lock()
	defer d.Unlock()

	if d.closed {
		return nil
	}
	d.closed = true

	if err := d.archive.Close(); err != nil {
		return err
	}
	return d.file.Close()
}

func (d *debugInfo) WriteGraph(step string, g *Graph) error {
	if d == nil {
		return nil
	}

	d.Lock()
	defer d.Unlock()

	// If we crash, the central directory won't be written, but we can rebuild
	// the archive if we have to if every file has been flushed and sync'ed.
	defer func() {
		d.archive.Flush()
		d.file.Sync()
	}()

	graphStr, err := GraphDot(g, &GraphDotOpts{
		DrawCycles: true,
		MaxDepth:   -1,
		Verbose:    true,
	})

	dotPath := fmt.Sprintf("debug/%d-%s.dot", d.step, step)
	d.step++

	fw, err := d.archive.Create(dotPath)
	if err != nil {
		return err
	}

	_, err = io.WriteString(fw, graphStr)
	return err
}

// writeFile writes data as a single file to the debug arhive.
func (d *debugInfo) writeFile(name string, data []byte) error {
	if d == nil {
		return nil
	}

	d.Lock()
	defer d.Unlock()

	path := fmt.Sprintf("debug/%d-%s", d.step, name)
	d.step++

	fw, err := d.archive.Create(path)
	if err != nil {
		return err
	}

	_, err = fw.Write(data)
	return err

}

// Return a a buffer we can write to, which will be added as a whole to the
// debug archive when it's closed.
func (d *debugInfo) StepLog(name string) *stepWriter {
	if d == nil {
		return nil
	}
	d.Lock()
	defer d.Unlock()

	name = fmt.Sprintf("%d-%s.log", d.step, name)
	d.step++

	return &stepWriter{
		name:      name,
		debugInfo: d,
	}
}

type stepWriter struct {
	name      string
	buf       bytes.Buffer
	debugInfo *debugInfo
}

func (s *stepWriter) Write(b []byte) (int, error) {
	if s == nil {
		return 0, nil
	}
	return s.buf.Write(b)
}

func (s *stepWriter) WriteString(str string) (int, error) {
	if s == nil {
		return 0, nil
	}
	return io.WriteString(&s.buf, str)
}

func (s *stepWriter) Close() error {
	if s == nil {
		return nil
	}
	return s.debugInfo.writeFile(s.name, s.buf.Bytes())
}

func (s *stepWriter) Printf(f string, args ...interface{}) (int, error) {
	if s == nil {
		return 0, nil
	}
	return fmt.Fprintf(&s.buf, f, args...)
}
