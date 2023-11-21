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
	// fmt.Println("header version:", header.Version)
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
		// fmt.Println("type:", entryName.Type, "name:", entryName.Name, "content length:", entryName.ContentLength)
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

func createWriter(dest string, force bool) func(string, []byte) error {
	os.MkdirAll(dest, 0755)
	return func(name string, content []byte) error {
		if !force {
			_, err := os.Stat(path.Join(dest, name))
			if !os.IsNotExist(err) {
				log.Warnf("File %s already exists", path.Join(dest, name))
				return nil
			}
		}
		return os.WriteFile(path.Join(dest, name), content, 0644)
	}
}

func ExtractFiles(name string, content []byte, dest string, force bool) error {
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
	return extractBatch(bytes.NewReader(content), nameReader, createWriter(dest, force))
}

func FindAndExtractBatches(paths []string, dest string, force bool) error {
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
		err = ExtractFiles(batch, content, dest, force)
		if err != nil {
			return err
		}
	}
	return nil
}
