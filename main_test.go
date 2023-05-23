package main

import (
	"bytes"
	"log"
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
	log.Print(buf.String())
	t.Fail()
	if err != nil {
		log.Print("Cant write to buffer")
	}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("mismatching byte slices, want: \n%s, got: \n%s", string(want), buf.String())
	}
}
