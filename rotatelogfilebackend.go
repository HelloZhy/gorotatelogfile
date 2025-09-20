package gorotatelogfile

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type rotateLogFileBackendInOut struct {
	LogDir,
	Prefix string
	MaxNumOfLogFiles,
	MaxNumOfLogEntries,
	LogEntryChBufferSize uint32

	LogEntryCh   <-chan []byte
	CloseEventCh chan<- struct{}
}

type rotateLogFileBackend struct {
	inOut rotateLogFileBackendInOut

	f            *os.File
	lineCount    uint32
	logFilePaths *list.List
}

func (b *rotateLogFileBackend) generateLogFilePath() string {
	return filepath.Join(b.inOut.LogDir, fmt.Sprintf("%s-%d.log", b.inOut.Prefix, time.Now().UnixMicro()))
}

func (b *rotateLogFileBackend) closeCurrentLogFile() {
	if b.f != nil {
		b.f.Sync()
		b.f.Close()
		b.f = nil
	}
}

func (b *rotateLogFileBackend) pushBackLogFilePath(logFilePath string) {
	b.logFilePaths.PushBack(logFilePath)
}

func (b *rotateLogFileBackend) removeLogFileOutOfDate() {
	if uint32(b.logFilePaths.Len()) <= b.inOut.MaxNumOfLogFiles {
		return
	}

	elementToBeRemoved := b.logFilePaths.Front()
	logFilePathOutOfDate := elementToBeRemoved.Value.(string)
	os.Remove(logFilePathOutOfDate)
	b.logFilePaths.Remove(elementToBeRemoved)
}

func (b *rotateLogFileBackend) closeCurrentAndOpenNewLogFile() error {
	logFilePath := b.generateLogFilePath()

	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	b.closeCurrentLogFile()
	b.f = f

	b.pushBackLogFilePath(logFilePath)
	b.removeLogFileOutOfDate()

	return nil
}

func (b *rotateLogFileBackend) Run() {
	defer close(b.inOut.CloseEventCh)
	if err := b.closeCurrentAndOpenNewLogFile(); err != nil {
		return
	}
	defer b.closeCurrentLogFile()

	for {
		entry, ok := <-b.inOut.LogEntryCh
		if !ok {
			return
		}

		if b.lineCount < b.inOut.MaxNumOfLogEntries {
			if _, err := b.f.Write(entry); err != nil {
				return
			}
			b.lineCount++
		} else {
			b.closeCurrentAndOpenNewLogFile()
			b.lineCount = 0
		}
	}
}

func newRotateLogFileBackend(inOut rotateLogFileBackendInOut) *rotateLogFileBackend {
	return &rotateLogFileBackend{
		inOut:        inOut,
		f:            nil,
		lineCount:    0,
		logFilePaths: list.New(),
	}
}
