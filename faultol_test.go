package faultol_test

import (
	"faultol"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndexFileExistence(t *testing.T) {
	faultol.Run()

	var files []string
	var root = os.Getenv("FAULTOL_HTML_PATH")

	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			files = append(files, path)
			return nil
		})

	if err != nil {
		t.Fatalf(err.Error())
	}

	for _, file := range files {
		if strings.Index(file, "index.html") > -1 { // index.html found that was generated from /
			os.Exit(0)
		}
	}

	t.Fatalf("index.html not found - try to regenerate static files with uri = / and the content " +
		"(in data json property) for index.html page")
}
