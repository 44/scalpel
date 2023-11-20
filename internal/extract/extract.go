package extract

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"fmt"
	"io"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"golang.org/x/text/transform"
	"os"
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
	win16be := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	unicodeReader := transform.NewReader(bytes.NewReader(buf), win16be.NewDecoder())
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
	mac32be := utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM)
	unicodeReader := transform.NewReader(bytes.NewReader(buf), mac32be.NewDecoder())
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
	fmt.Println("header version:", header.Version)
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
		fmt.Println("type:", entryName.Type, "name:", entryName.Name, "content length:", entryName.ContentLength)
		if writer != nil {
			content := make([]byte, entryName.ContentLength)
			err = binary.Read(reader, binary.LittleEndian, &content)
			if err != nil {
				return err
			}
			err = writer(entryName.Name, content)
			if err != nil {
				return err
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

func writeFile(name string, content []byte) error {
	return os.WriteFile(name, content, 0644)
}

func ExtractFiles(content []byte, dest string) error {
	var nameReader func(*bytes.Reader, uint32, *EntryName) error
	err := extractBatch(bytes.NewReader(content), readNameWindows, nil)
	if err == nil {
		nameReader = readNameWindows
	} else {
		err = extractBatch(bytes.NewReader(content), readNameMac, nil)
		if err == nil {
			nameReader = readNameMac
		} else {
			return err
		}
	}
	return extractBatch(bytes.NewReader(content), nameReader, writeFile)
}

// func main() {
// 	// scalpel -vvv --verbose -t --to <dir> -f --force -z --ungzip -m --match <pattern> {file|dir}
// 	entries, err := ioutil.ReadDir(".")
// 	if err != nil {
// 		fmt.Println("ioutil.ReadDir failed:", err)
// 		return
// 	}
// 	for _, entry := range entries {
// 		fmt.Println(entry.Name())
// 		batch := "./" + entry.Name()
// 		data, err := ioutil.ReadFile(batch)
// 		if err != nil {
// 			continue
// 		}
// 		err = extractFiles(data)
// 		if err != nil {
// 			continue
// 		}
// 	}
// }
//
