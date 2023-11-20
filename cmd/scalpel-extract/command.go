package main

import (
	"flag"
	"fmt"
	"github.com/44/scalpel/internal/extract"
	"io/ioutil"
)

func main() {
	var output string
	flag.StringVar(&output, "o", "", "Output dir")
	flag.Parse()

	// scalpel -vvv --verbose -t --to <dir> -f --force -z --ungzip -m --match <pattern> {file|dir}
	entries, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("ioutil.ReadDir failed:", err)
		return
	}
	for _, entry := range entries {
		fmt.Println(entry.Name())
		batch := "./" + entry.Name()
		data, err := ioutil.ReadFile(batch)
		if err != nil {
			continue
		}
		err = extract.ExtractFiles(data, output)
		if err != nil {
			continue
		}
	}
}
