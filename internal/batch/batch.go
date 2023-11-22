package batch

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path"
	log "github.com/sirupsen/logrus"
	"regexp"
)

type Header struct {
	Version uint32
}

type EntryName struct {
	Type uint16
	Length uint32
	Name string
	ContentLength uint64
}

func readNameWindows(reader *bytes.Reader, length uint32, entry *EntryName) error {
	buf := make([]byte, length * 2)
	err := binary.Read(reader, binary.LittleEndian, &buf)
	if err != nil {
		return err
	}
	enc := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	unicodeReader := transform.NewReader(bytes.NewReader(buf), enc.NewDecoder())
	decoded, err := ioutil.ReadAll(unicodeReader)
	if err != nil {
		return err
	}
	entry.Name = string(decoded)
	return nil
}

func readNameMac(reader *bytes.Reader, length uint32, entry *EntryName) error {
	buf := make([]byte, length * 4)
	err := binary.Read(reader, binary.BigEndian, &buf)
	if err != nil {
		return err
	}
	enc := utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM)
	unicodeReader := transform.NewReader(bytes.NewReader(buf), enc.NewDecoder())
	decoded, err := ioutil.ReadAll(unicodeReader)
	if err != nil {
		return err
	}
	entry.Name = string(decoded)
	return nil
}

func extractBatch(reader *bytes.Reader, nameReader func(*bytes.Reader, uint32, *EntryName)  error, writer func(string, []byte) error) error {
	var header Header
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return err
	}
	if header.Version != 2 {
		return fmt.Errorf("header version %d is not supported", header.Version)
	}

	for {
		var entryName EntryName
		err := binary.Read(reader, binary.LittleEndian, &entryName.Type)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err = binary.Read(reader, binary.LittleEndian, &entryName.Length)
		if err != nil {
			return err
		}
		err = nameReader(reader, entryName.Length, &entryName)
		if err != nil {
			return err
		}
		
		err = binary.Read(reader, binary.LittleEndian, &entryName.ContentLength)
		if err != nil {
			return err
		}
		if entryName.ContentLength > uint64(reader.Len()) {
			return fmt.Errorf("ContentLength %d is larger than remaining data %d", entryName.ContentLength, reader.Len())
		}

		if writer != nil {
			content := make([]byte, entryName.ContentLength)
			err = binary.Read(reader, binary.LittleEndian, &content)
			if err != nil {
				return err
			}
			err = writer(entryName.Name, content)
			if err != nil {
				return err
			} else {
				log.Infof("\t%s", entryName.Name)
			}

		} else {
			_, err := reader.Seek(int64(entryName.ContentLength), io.SeekCurrent)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func printExtracted(name string, size int, printSize bool) {
	if printSize {
		fmt.Printf("%s\t%d\n", name, size)
	} else {
		fmt.Printf("%s\n", name)
	}
}

func createWriter(opts Options) func(string, []byte) error {
	if opts.Test {
		return func(name string, content []byte) error {
			printExtracted(path.Join(opts.Dest, name), len(content), opts.Long)
			return nil
		}
	}
	os.MkdirAll(opts.Dest, 0755)
	return func(name string, content []byte) error {
		full_path := path.Join(opts.Dest, name)
		if !opts.Force {
			_, err := os.Stat(full_path)
			if !os.IsNotExist(err) {
				log.Warnf("File %s already exists", path.Join(opts.Dest, name))
				return nil
			}
		}
		printExtracted(full_path, len(content), opts.Long)
		return os.WriteFile(full_path, content, 0644)
	}
}

func createFilter(opts Options, writer func(string, []byte) error) func(string, []byte) error {
	if len(opts.Match) == 0 {
		return writer
	}
	matchers := make([]*regexp.Regexp, 0)
	for _, match := range opts.Match {
		matcher, err := regexp.Compile(match)
		if err != nil {
			log.Warnf("Failed to compile regex %s: %s", match, err)
			continue
		}
		matchers = append(matchers, matcher)
	}

	return func(name string, content []byte) error {
		found := false
		for _, match := range matchers {
			if match.MatchString(name) {
				found = true
			}
		}
		if found {
			return writer(name, content)
		}
		return nil
	}	
}

func ExtractFiles(name string, content []byte, opts Options) error {
	var nameReader func(*bytes.Reader, uint32, *EntryName) error
	err := extractBatch(bytes.NewReader(content), readNameWindows, nil)
	if err == nil {
		log.Infof("Windows batch: %s", name)
		nameReader = readNameWindows
	} else {
		err = extractBatch(bytes.NewReader(content), readNameMac, nil)
		if err == nil {
			log.Infof("Mac batch: %s", name)
			nameReader = readNameMac
		} else {
			return err
		}
	}
	return extractBatch(bytes.NewReader(content), nameReader, createFilter(opts, createWriter(opts)))
}

type Options struct {
	Dest string
	Force bool
	Unpack bool
	Test bool
	Long bool
	Match []string
}

func FindAndExtractBatches(paths []string, opts Options) error {
	batches := make([]string, 0)
	real_paths := paths
	if len(paths) == 0 {
		real_paths = []string{"."}
	}
	for _, path := range real_paths {
		fi, err := os.Stat(path)
		if err != nil {
			log.Warnf("Failed to accesss %s: %s", path, err)
			continue
		}
		if fi.IsDir() {
		} else {
			if fi.Size() > 256 {
				batches = append(batches, path)
			}
		}
	}
	for _, batch := range batches {
		content, err := os.ReadFile(batch)
		if err != nil {
			return err
		}
		err = ExtractFiles(batch, content, opts)
		if err != nil {
			return err
		}
	}
	return nil
}
