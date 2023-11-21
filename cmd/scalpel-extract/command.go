package main

import (
	"flag"
	"github.com/44/scalpel/internal/extract"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
)

func main() {
	var output string
	flag.StringVar(&output, "o", "", "Output dir")
	flag.Parse()

	// scalpel -vvv --verbose -t --to <dir> -f --force -z --ungzip -m --match <pattern> {file|dir}
	entries, err := ioutil.ReadDir(".")
	if err != nil {
		log.Error("ioutil.ReadDir failed:", err)
		return
	}
	for _, entry := range entries {
		batch := "./" + entry.Name()
		data, err := ioutil.ReadFile(batch)
		if err != nil {
			log.Warnf("Failed to read file: %s. Error: %v", batch, err)
			continue
		}
		err = extract.ExtractFiles(entry.Name(), data, output)
		if err != nil {
			log.Warnf("Failed to extract batch %s. Error: %v", batch, err)
			continue
		}
	}
}
