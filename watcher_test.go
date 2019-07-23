package watcher

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBadFile(t *testing.T) {
	_, err := New("/rot/random")
	if err == nil {
		t.Fatal("Expected error but found none")
	}
}

func TestBasicWatch(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Fatalf("unable to create temp file: %s\n", err)
	}
	defer os.Remove(file.Name())
	watcher, err := New(file.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	time.Sleep(time.Second)
	if ok, _ := watcher.WasModified(); ok {
		t.Fatalf("File %s reports as modified but should not!", file.Name())
	}
	file.WriteString("some data")
	if ok, _ := watcher.WasModified(); !ok {
		t.Fatalf("File %s does not report as modified but should!", file.Name())
	}
}

func TestLostFile(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Fatalf("unable to create temp file: %s\n", err)
	}
	defer os.Remove(file.Name())
	watcher, err := New(file.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	time.Sleep(time.Second)
	os.Remove(file.Name())
	_, err = watcher.WasModified()
	if err == nil {
		t.Fatalf("File %s reports as modified but should not!", file.Name())
	}
	if !strings.Contains(err.Error(), "no longer accessible") {
		t.Fatalf("Unexpected error '%v'", err)
	}
}

func TestGetNewBytes(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Fatalf("unable to create temp file: %s\n", err)
	}
	defer os.Remove(file.Name())
	want := "a line"
	file.WriteString(want)
	watcher, err := New(file.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	got, _ := watcher.GetNewBytes()
	if len(got) == 0 {
		t.Fatal("Expected the first line of bytes but got none")
	}
	if string(got) != want {
		t.Fatalf("Wanted string: '%s' but got '%s'", want, string(got))
	}
}

func TestGetNewBytesFromEmpty(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Fatalf("unable to create temp file: %s\n", err)
	}
	defer os.Remove(file.Name())
	watcher, err := New(file.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	got, _ := watcher.GetNewBytes()
	if len(got) > 0 {
		t.Fatal("Unexpected receipt of new bytes!")
	}
	time.Sleep(time.Second)
	want := "new line"
	file.WriteString(want)
	got, _ = watcher.GetNewBytes()
	if len(got) == 0 {
		t.Fatal("Expected new bytes but got none")
	}
	if string(got) != want {
		t.Fatalf("Wanted string: '%s' but got '%s'", want, string(got))
	}
	time.Sleep(time.Second)
	want = "next line"
	file.WriteString(want)
	got, _ = watcher.GetNewBytes()
	if len(got) == 0 {
		t.Fatal("Expected new bytes but got none")
	}
	if string(got) != want {
		t.Fatalf("Wanted string: '%s' but got '%s'", want, string(got))
	}
}

func TestGetChannelBytes(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		t.Fatalf("unable to create temp file: %s\n", err)
	}
	defer os.Remove(file.Name())
	want := "first line"
	file.WriteString(want)
	watcher, err := New(file.Name(), WithChannelMonitor(2))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	got := <-watcher.ByteChannel
	if len(got) == 0 {
		t.Fatal("Expected new bytes but got none")
	}
	if string(got) != want {
		t.Fatalf("Wanted string: '%s' but got '%s'", want, string(got))
	}
	time.Sleep(time.Second)
	want = "second line"
	file.WriteString(want)
	got = <-watcher.ByteChannel
	if len(got) == 0 {
		t.Fatal("Expected new bytes but got none")
	}
	if string(got) != want {
		t.Fatalf("Wanted string: '%s' but got '%s'", want, string(got))
	}
	os.Remove(file.Name())
	err = <-watcher.ErrChannel
	if !strings.Contains(err.Error(), "no longer accessible") {
		t.Fatalf("Unexpected error '%v'", err)
	}
}
