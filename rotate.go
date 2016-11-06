package logrotate

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

// Config configures a RotatingWriter.
type Config struct {
	// MaxSize is the size after which log files will be rotated.
	MaxSize int64 `json:"max_size"`
	// MaxFiles is the maximum number of backup files that will be kept.
	MaxFiles int `json:"max_files"`
	// Path is the root file name of the output log file.
	Path string `json:"path"`
}

// RotatingWriter is a writer that will rotate log files when the output
// reaches a certain size.
type RotatingWriter struct {
	outputFile  *os.File
	currentSize int64
	config      Config
	quitChan    chan chan struct{}
	dataChan    chan []byte
}

// New creates a new RotatingWriter with the supplied config.
func New(config Config) (w *RotatingWriter, err error) {
	if err = os.MkdirAll(path.Dir(config.Path), 0666); err != nil {
		return
	}
	w = &RotatingWriter{
		config:   config,
		quitChan: make(chan chan struct{}),
		dataChan: make(chan []byte, 1024),
	}
	if err = w.openFile(); err != nil {
		return nil, err
	}
	go w.listen()
	return
}

func (w *RotatingWriter) openFile() (err error) {
	if w.outputFile != nil {
		w.outputFile.Close()
		w.outputFile = nil
	}
	if w.outputFile, err = os.OpenFile(w.config.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666); err != nil {
		return
	}
	fileInfo, err := w.outputFile.Stat()
	if err != nil {
		return
	}
	w.currentSize = fileInfo.Size()
	if w.currentSize > w.config.MaxSize {
		return w.rotate()
	}
	return
}

func (w *RotatingWriter) rotate() (err error) {
	if w.outputFile != nil {
		w.outputFile.Close()
		w.outputFile = nil
	}
	fileTime := time.Now().UTC()
	newName := w.config.Path + "." + strings.Replace(fileTime.Format(time.RFC3339), ":", "-", -1)
	os.Rename(w.config.Path, newName)
	if err = w.openFile(); err != nil {
		return
	}
	baseName := path.Base(w.config.Path)
	logDir := path.Dir(w.config.Path)
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return err
	}
	logFileCount := 0
	for i := len(files) - 1; i >= 0; i-- {
		f := files[i]
		if strings.HasPrefix(f.Name(), baseName) {
			if logFileCount >= w.config.MaxFiles && f.Name() != baseName {
				os.Remove(path.Join(logDir, f.Name()))
			} else {
				logFileCount++
			}
		}
	}
	return
}

func (w *RotatingWriter) listen() {
	for {
		select {
		case b := <-w.dataChan:
			n, err := w.outputFile.Write(b)
			if err != nil {
				w.openFile()
				continue
			}
			w.currentSize += int64(n)
			if w.currentSize >= w.config.MaxSize {
				w.rotate()
			}
		case c := <-w.quitChan:
			if w.outputFile != nil {
				w.outputFile.Close()
			}
			close(c)
		}
	}
}

// Write writes data to the RotatingWriter
func (w *RotatingWriter) Write(b []byte) (int, error) {
	w.dataChan <- b
	return len(b), nil
}

// Close shuts down the writer
func (w *RotatingWriter) Close() {
	c := make(chan struct{})
	w.quitChan <- c
	<-c
}
