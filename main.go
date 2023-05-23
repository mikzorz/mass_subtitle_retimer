package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/asticode/go-astisub"
)

var offset float64
var filenames []string

func main() {
	parseFlags()
	newSubs, err := offsetSubFiles(filenames, offset)
	if err != nil {
		log.Fatal(err)
	}
	err = moveOldFiles(filenames, "old_subs")
	if err != nil {
		log.Fatal(err)
	}
	err = saveNewFiles(newSubs, filenames, "retimed_subs")
	if err != nil {
		log.Fatal(err)
	}
}

func parseFlags() {
	flag.Float64Var(&offset, "o", 0, "Offset (in seconds). '-o 1' makes subs appear 1 second later, '-o -2.5' 2.5 seconds earlier")
	flag.Parse()
	if offset == 0 {
		fmt.Println("Must provide an offset")
		os.Exit(1)
	}
	filenames = flag.Args()
	if len(filenames) == 0 {
		fmt.Println("Must provide paths to sub files.")
		os.Exit(1)
	}
}

// take paths of subs, try to open, apply offset, return new subs as []byte
func offsetSubFiles(paths []string, secondsOffset float64) ([][]byte, error) {
	ret := [][]byte{}
	for _, s := range paths {
		// read file
		s, err := astisub.OpenFile(s)
		if err != nil {
			return ret, err
		}

		// offset
		s.Add(time.Duration(secondsOffset * float64(time.Second)))

		// save to buffer
		var buf = &bytes.Buffer{}
		err = s.WriteToSRT(buf)
		if err != nil {
			return ret, err
		}

		ret = append(ret, buf.Bytes())
	}
	return ret, nil
}

func moveOldFiles(filenames []string, targetDir string) error {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err := os.Mkdir(targetDir, os.FileMode(0700))
		if err != nil {
			return err
		}
	}

	for _, f := range filenames {
		inputFile, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("Couldn't open source file: %s", err)
		}
		outputFile, err := os.Create(targetDir + "/" + path.Base(f))
		if err != nil {
			inputFile.Close()
			return fmt.Errorf("Couldn't open dest file: %s", err)
		}
		defer outputFile.Close()
		_, err = io.Copy(outputFile, inputFile)
		inputFile.Close()
		if err != nil {
			return fmt.Errorf("Writing to output file failed: %s", err)
		}
	}

	for _, f := range filenames {
		// The copying was successful, so now delete the original files
		err := os.Remove(f)
		if err != nil {
			return fmt.Errorf("Failed removing original file: %s", err)
		}
	}

	return nil
}

func saveNewFiles(newSubs [][]byte, filenames []string, targetDir string) error {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		err := os.Mkdir(targetDir, os.FileMode(0700))
		if err != nil {
			return err
		}
	}

	for i, s := range newSubs {
		name := path.Base(filenames[i])
		out, err := os.Create(targetDir + "/" + name)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = out.Write(s)
		if err != nil {
			return err
		}
	}
	return nil
}
