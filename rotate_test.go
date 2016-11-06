package logrotate

import (
	"io/ioutil"
	"path"
	"testing"
	"time"
)

func TestRotate(t *testing.T) {
	testDir, err := ioutil.TempDir("", "rotate")
	if err != nil {
		t.Fatal(err)
	}
	c := Config{
		MaxSize:  100,
		MaxFiles: 10,
		Path:     path.Join(testDir, "log"),
	}
	writer, err := New(c)
	if err != nil {
		t.Fatal(err)
	}
	buf := []byte("1234567890")
	for i := 0; i < 9; i++ {
		writer.Write(buf)
	}
	time.Sleep(1 * time.Millisecond)
	assertFileCount(1, testDir, t)
	writer.Write(buf)
	time.Sleep(1 * time.Millisecond)
	assertFileCount(2, testDir, t)
	for i := 0; i < 9; i++ {
		writer.Write(buf)
	}
	assertFileCount(2, testDir, t)
	for i := 0; i < 70; i++ {
		if i%10 == 0 {
			time.Sleep(time.Second)
		}
		writer.Write(buf)
	}
	time.Sleep(time.Second)
	assertFileCount(9, testDir, t)
	for i := 0; i < 10; i++ {
		writer.Write(buf)
	}
	time.Sleep(time.Second)
	assertFileCount(10, testDir, t)
	for i := 0; i < 10; i++ {
		writer.Write(buf)
	}
	time.Sleep(time.Second)
	assertFileCount(11, testDir, t)
	for i := 0; i < 10; i++ {
		writer.Write(buf)
	}
	time.Sleep(time.Second)
	assertFileCount(11, testDir, t)
	for i := 0; i < 10; i++ {
		writer.Write(buf)
	}
	time.Sleep(time.Second)
	assertFileCount(11, testDir, t)
	for i := 0; i < 10; i++ {
		writer.Write(buf)
	}
	time.Sleep(time.Second)
	assertFileCount(11, testDir, t)
	for i := 0; i < 10; i++ {
		writer.Write(buf)
	}
	time.Sleep(time.Second)
	assertFileCount(11, testDir, t)
	writer.Close()
}

func assertFileCount(count int, testDir string, t *testing.T) {
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != count {
		t.Fatal("Expected", count, "files, found", len(files))
	}
}
