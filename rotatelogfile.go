package gorotatelogfile

import (
	"errors"
	"slices"
	"time"
)

type config struct {
	MaxNumOfLogFiles,
	MaxNumOfLogEntries,
	LogEntryChBufferSize uint32
}

type configOption func(*config)

func WithMaxNumOfLogFiles(val uint32) configOption {
	f := func(cfg *config) {
		cfg.MaxNumOfLogFiles = val
	}

	return f
}

func WithMaxNumOfLogEntries(val uint32) configOption {
	f := func(cfg *config) {
		cfg.MaxNumOfLogEntries = val
	}

	return f
}

func WithLogEntryChBufferSize(val uint32) configOption {
	f := func(cfg *config) {
		cfg.LogEntryChBufferSize = val
	}

	return f
}

const (
	defaultMaxNumOfLogFiles     uint32 = 10
	defaultMaxNumOfLogEntries   uint32 = 1024 * 16
	defaultLogEntryChBufferSize uint32 = 4096
)

type RotateLogFile struct {
	logEntryCh          chan<- []byte
	backendCloseEventCh <-chan struct{}
	backend             *rotateLogFileBackend
}

func (f *RotateLogFile) Write(p []byte) (n int, err error) {
	// NOTE: need a clone for slice, because input p []byte might be modified and reused
	f.logEntryCh <- slices.Clone(p)

	return len(p), nil
}

func (f *RotateLogFile) Close() error {
	close(f.logEntryCh)

	var err error
	select {
	case <-f.backendCloseEventCh:
	case <-time.After(time.Second):
		err = errors.New("close backendWriter timeout, shutdown immediately")
	}

	return err
}

func NewRotateLogFile(logDir, prefix string, opts ...configOption) *RotateLogFile {
	config := config{
		MaxNumOfLogFiles:     defaultMaxNumOfLogFiles,
		MaxNumOfLogEntries:   defaultMaxNumOfLogEntries,
		LogEntryChBufferSize: defaultLogEntryChBufferSize,
	}

	for _, opt := range opts {
		opt(&config)
	}

	logEntryCh := make(chan []byte, config.LogEntryChBufferSize)
	backendCloseEventCh := make(chan struct{}, 1)

	inOut := rotateLogFileBackendInOut{
		LogDir:               logDir,
		Prefix:               prefix,
		MaxNumOfLogFiles:     config.MaxNumOfLogFiles,
		MaxNumOfLogEntries:   config.MaxNumOfLogEntries,
		LogEntryChBufferSize: config.LogEntryChBufferSize,
		LogEntryCh:           logEntryCh,
		CloseEventCh:         backendCloseEventCh,
	}

	f := &RotateLogFile{
		logEntryCh:          logEntryCh,
		backendCloseEventCh: backendCloseEventCh,
		backend:             newRotateLogFileBackend(inOut),
	}

	go f.backend.Run()

	return f
}
