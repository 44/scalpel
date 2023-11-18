package main

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"fmt"
	"io"
)

type Header struct {
	Version uint32
}

type EntryName struct {
	Type uint16
	Length uint32
	NameBytes []byte
	ContentLength uint64
}


func extractBatch(reader *bytes.Reader) {
	// TODO
	var header Header
	err := binary.Read(reader, binary.LittleEndian, &header)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
		return
	}
	fmt.Println("header version:", header.Version)
	for {
		var entryName EntryName
		err := binary.Read(reader, binary.LittleEndian, &entryName.Type)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("binary.Read failed:", err)
			break
		}
		err = binary.Read(reader, binary.LittleEndian, &entryName.Length)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
			break
		}
		entryName.NameBytes = make([]byte, entryName.Length * 2)
		err = binary.Read(reader, binary.BigEndian, &entryName.NameBytes)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
			break
		}
		err = binary.Read(reader, binary.LittleEndian, &entryName.ContentLength)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
			break
		}
		fmt.Println("type:", entryName.Type, "name:", string(entryName.NameBytes), "content length:", entryName.ContentLength)
		// fmt.Println("entryName length:", entryName.Length)
		// fmt.Println("entryName name:", string(entryName.NameBytes))
		// fmt.Println("entryName content length:", entryName.ContentLength)
		newpos, err := reader.Seek(int64(entryName.ContentLength), io.SeekCurrent)
		if err != nil {
			fmt.Println("binary.Read failed:", err)
			break
		}
		fmt.Println("newpos:", newpos)
		// io.CopyN(ioutil.Discard, reader, int64(entryName.ContentLength))
	}
}

func main() {
	entries, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("ioutil.ReadDir failed:", err)
		return
	}
	for _, entry := range entries {
		fmt.Println(entry.Name())
		batch := "./" + entry.Name()
		data, _ := ioutil.ReadFile(batch)
		reader := bytes.NewReader(data)
		extractBatch(reader)
	}

	// TODO
}

