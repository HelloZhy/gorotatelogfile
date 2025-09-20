package gorotatelogfile

import (
	"os"
	"strconv"
	"testing"
)

func TestRotateLogFile(t *testing.T) {
	const logDir = "log"
	if err := os.Mkdir(logDir, 0777); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(logDir)

	const userDefinedMaxNumOfLogFiles uint32 = 3
	const userDefinedMaxNumOfLogEntries uint32 = 64
	const userDefinedLogEntryChBufferSize uint32 = 16

	l := NewRotateLogFile(
		logDir,
		"test",
		WithMaxNumOfLogFiles(userDefinedMaxNumOfLogFiles),
		WithMaxNumOfLogEntries(userDefinedMaxNumOfLogEntries),
		WithLogEntryChBufferSize(userDefinedLogEntryChBufferSize),
	)
	if l == nil {
		t.Fatal("newRotateLogFile failed, return value is nil")
	}
	defer l.Close()

	logStrSample := []string{
		"abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd",
		"edfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfgedfg",
		"()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+()-+",
		"<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/<>?/",
	}

	for i := range userDefinedMaxNumOfLogEntries * (userDefinedMaxNumOfLogFiles + 1) {
		logStr := logStrSample[i%uint32(len(logStrSample))]
		if _, err := l.Write([]byte(logStr + strconv.Itoa(int(i)) + "\n")); err != nil {
			t.Fatal(err)
		}
	}

	dirEntries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatal(err)
	}

	lenOfDirEntries := len(dirEntries)
	if lenOfDirEntries != int(userDefinedMaxNumOfLogFiles) {
		t.Fatalf("num of dirEntries: %d, expected: %d", lenOfDirEntries, userDefinedMaxNumOfLogFiles)
	}
}
