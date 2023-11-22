package batch

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"github.com/44/scalpel/internal/odl"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	log "github.com/sirupsen/logrus"
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

func decompressLog(content []byte, name string) ([]byte, string) {
	var header odl.Header
	reader := bytes.NewReader(content)
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		return content, name
	}
	if header.Magic != odl.HeaderMagicValue {
		return content, name
	}
	if header.Version != 1 && header.Version != 2 && header.Version != 3 {
		return content, name
	}
	if (header.Capabilities & odl.Capabilities_CompressedContents == 0) && (header.Capabilities & odl.Capabilities_CompressedContentsChunked == 0) {
		return content, name
	}
	log.Debugf("Decompressing log with version %d", header.Version)
	gz, err := gzip.NewReader(reader) //bytes.NewReader(content[256:]))
	if err != nil {
		return content, name
	}
	defer gz.Close()
	plain, err := ioutil.ReadAll(gz)
	if err != nil {
		return content, name
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	header.Capabilities &= ^uint32(odl.Capabilities_CompressedContents)
	header.Capabilities &= ^uint32(odl.Capabilities_CompressedContentsChunked)
	err = binary.Write(writer, binary.LittleEndian, &header)
	if err != nil {
		log.Warnf("Failed to write header: %s", err)
		return content, name
	}
	_, err = writer.Write(plain)
	if err != nil {
		log.Warnf("Failed to write content: %s", err)
		return content, name
	}
	log.Debugf("Decompressed log %d to %d bytes", len(content), len(plain))
	if strings.HasSuffix(name, "gz") {
		return buf.Bytes(), name[:len(name)-2]
	}
	return buf.Bytes(), name
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
		new_content, new_name := content, name
		if opts.Unpack {
			new_content, new_name = decompressLog(content, name)
		}
		full_path := path.Join(opts.Dest, new_name)
		if !opts.Force {
			_, err := os.Stat(full_path)
			if !os.IsNotExist(err) {
				log.Warnf("File %s already exists", full_path)
				return nil
			}
		}
		printExtracted(full_path, len(new_content), opts.Long)
		return os.WriteFile(full_path, new_content, 0644)
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
	log.Debugf("Extracting %s", name)
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
			continue
		}
		if fi.IsDir() {
			filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if (!info.IsDir()) && (info.Size() > 256) {
					batches = append(batches, path)
				}
				return nil
			})
		} else {
			if fi.Size() > 256 {
				batches = append(batches, path)
			}
		}
	}
	log.Debugf("Found %d potential batches", len(batches))
	for _, batch := range batches {
		content, err := os.ReadFile(batch)
		if err != nil {
			return err
		}
		err = ExtractFiles(batch, content, opts)
		if err != nil {
			log.Warnf("Failed to extract %s: %s", batch, err)
		}
	}
	return nil
}
