package main

import (
	"bytes"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/asticode/go-astisub"
)

var bytesBOM = []byte{239, 187, 191}

func TestApply1SecondOffset(t *testing.T) {
	in := bytes.NewReader([]byte("\n00:01:00.000 --> 00:02:00.000\nCredits"))
	s2, err := astisub.ReadFromSRT(in)
	if err != nil {
		log.Print(err)
	}
	s2.Add(time.Second)

	// astisub adds extra bytes, must account for that
	want := bytesBOM
	want = append(want, []byte("1")...) // Line number
	want = append(want, []byte("\n00:01:01,000 --> 00:02:01,000\nCredits")...)
	want = append(want, []byte("\n")...)

	var buf = &bytes.Buffer{}
	err = s2.WriteToSRT(buf)
	if err != nil {
		t.Error("Cant write to buffer")
	}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("mismatching byte slices, want: \n%s, got: \n%s", string(want), buf.String())
	}
}

func TestOffsetMultipleFiles(t *testing.T) {
	inPaths := []string{"testfiles/test1.srt", "testfiles/test2.srt"}

	t.Run("1 second offset", func(t *testing.T) {
		newSubs, err := offsetSubFiles(inPaths, 1)
		if err != nil {
			t.Fatal(err)
		}

		want := []string{"00:01:33,109 --> 00:01:38,047", "00:05:21,000 --> 00:05:26,500"}
		for i, s := range newSubs {

			if !bytes.Contains(s, []byte(want[i])) {
				t.Fatalf("can't find correct times in file %s, want: \n%s\ngot: \n%s", inPaths[i], []byte(want[i]), string(s))
			}
		}
	})

	t.Run("minus 1 second offset", func(t *testing.T) {
		newSubs, err := offsetSubFiles(inPaths, -1)
		if err != nil {
			t.Fatal(err)
		}

		want := []string{"00:01:31,109 --> 00:01:36,047", "00:05:19,000 --> 00:05:24,500"}
		for i, s := range newSubs {

			if !bytes.Contains(s, []byte(want[i])) {
				t.Fatalf("can't find correct times, want: \n%s\ngot: \n%s", []byte(want[i]), string(s))
			}
		}
	})

	t.Run("minus 2.5 second offset", func(t *testing.T) {
		newSubs, err := offsetSubFiles(inPaths, -2.5)
		if err != nil {
			t.Fatal(err)
		}

		want := []string{"00:01:29,609 --> 00:01:34,547", "00:05:17,500 --> 00:05:23,000"}
		for i, s := range newSubs {

			if !bytes.Contains(s, []byte(want[i])) {
				t.Fatalf("can't find correct times, want: \n%s\ngot: \n%s", []byte(want[i]), string(s))
			}
		}
	})

	t.Run("try to offset non-existent file", func(t *testing.T) {
		ret, err := offsetSubFiles([]string{"does_not_exist.srt"}, 1)
		if err == nil {
			t.Error("offsetSubFiles did not return error for non-existent file")
		}

		if len(ret) > 0 {
			t.Errorf("byte slice return by offsetSubFiles should be empty, got len: %d", len(ret))
		}
	})
}

func TestMoveOldFiles(t *testing.T) {
	fileToMove := "testfiles/move_me.srt"
	oldSubsDir := "testfiles/old_subs"

	cleanup := func() {
		os.Remove(fileToMove)
		os.RemoveAll(oldSubsDir)
	}
	defer cleanup()

	_, err := os.Create(fileToMove)
	if err != nil {
		t.Fatal(err)
	}

	err = moveOldFiles([]string{fileToMove}, oldSubsDir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(oldSubsDir + "/" + path.Base(fileToMove)); os.IsNotExist(err) {
		// testfiles/old_subs/move_me.srt does not exist
		t.Fatal(err)
	}
}

func TestSaveNewFiles(t *testing.T) {
	targetDir := "testfiles/retimed_subs"

	filenames := []string{"testfiles/test1.srt", "testfiles/test2.srt"}

	cleanup := func() {
		os.RemoveAll(targetDir)
	}
	defer cleanup()

	newSubs, err := offsetSubFiles(filenames, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = saveNewFiles(newSubs, filenames, targetDir)

	for _, f := range filenames {
		if _, err := os.Stat(targetDir + "/" + path.Base(f)); os.IsNotExist(err) {
			// testfiles/retimed_subs/$f does not exist
			t.Fatal(err)
		}
	}

}
