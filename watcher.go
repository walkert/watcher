package watcher

import (
	"fmt"
	"io"
	"os"
	"time"
)

var (
	lostError = "File %s is no longer accessible"
)

// Watcher represents a file watcher object
type Watcher struct {
	ByteChannel    chan ([]byte)
	chanInterval   int
	ErrChannel     chan (error)
	fileLost       bool
	fileName       string
	initialModTime time.Time
	readBytes      int
	lastStartBytes int
	lastModified   time.Time
	useChan        bool
}

// WatcherOpt defines a simple funcion for performing config actions on a Watcher object
type WatcherOpt func(w *Watcher)

func (w *Watcher) getModTime() time.Time {
	stat, err := os.Stat(w.fileName)
	if err != nil {
		w.fileLost = true
		return time.Time{}
	}
	return stat.ModTime()
}

// WasModified reports whether the watched file has been modified
func (w *Watcher) WasModified() (bool, error) {
	mod := w.getModTime()
	if w.fileLost {
		return false, fmt.Errorf(lostError, w.fileName)
	}
	var checkTime time.Time
	if w.lastModified.Sub(w.initialModTime) > 0 {
		checkTime = w.lastModified
	} else {
		checkTime = w.initialModTime
	}
	if mod.Sub(checkTime) > 0 {
		w.lastModified = mod
		return true, nil
	}
	return false, nil
}

// GetNewBytes returns a slice containing any newly read bytes if any
func (w *Watcher) GetNewBytes() ([]byte, error) {
	ok, err := w.WasModified()
	if err != nil {
		return nil, err
	}
	if !ok {
		if w.readBytes != 0 {
			return []byte{}, nil
		}
	}
	return w.byteDelta()
}

// startChannelMonitor starts a goroutine that looks for new bytes on an interval and
// returns them to the ByteChannel (or an error to the ErrChannel)
func (w *Watcher) startChannelMonitor() {
	go func() {
		timer := time.NewTicker(time.Duration(w.chanInterval) * time.Second)
		for {
			<-timer.C
			bytes, err := w.GetNewBytes()
			if err != nil {
				w.ErrChannel <- err
				w.ByteChannel <- []byte{}
			}
			if len(bytes) > 0 {
				w.ByteChannel <- bytes
			}
		}
	}()
}

// byteDelta returns a byte slice of bytes read from the modified file from either the
// start or an offset if the file has already been read
func (w *Watcher) byteDelta() ([]byte, error) {
	f, err := os.Open(w.fileName)
	if err != nil {
		return nil, err
	}
	if w.readBytes > 0 {
		f.Seek(int64(w.readBytes), 0)
	}
	buf := make([]byte, 32*1024)
	var contents []byte
top:
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				break top
			}
			return contents, err
		}
		if n > 0 {
			w.lastStartBytes = w.readBytes
			w.readBytes += n
			contents = append(contents, buf[:n]...)
		}
	}
	return contents, nil
}

// WithChannelMonitor is a WatcherOpt that runs the Watcher in 'Channel'
// mode - returning new byes (or an error) to a channel
func WithChannelMonitor(interval int) WatcherOpt {
	return func(w *Watcher) {
		c := make(chan ([]byte), 1)
		e := make(chan (error), 1)
		w.chanInterval = interval
		w.useChan = true
		w.ByteChannel = c
		w.ErrChannel = e
	}
}

// New returns an initialized Watcher object with any options processed
func New(file string, opts ...WatcherOpt) (*Watcher, error) {
	stat, err := os.Stat(file)
	if err != nil {
		return &Watcher{}, err
	}
	watcher := &Watcher{
		fileName:       file,
		initialModTime: stat.ModTime(),
		lastModified:   stat.ModTime(),
	}
	for _, opt := range opts {
		opt(watcher)

	}
	if watcher.useChan {
		watcher.startChannelMonitor()
	}
	return watcher, nil
}
