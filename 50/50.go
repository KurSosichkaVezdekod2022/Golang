package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type protobuf interface {
	decode(reader *bufio.Reader) bool
	filled() bool
}

// Varint | 64-bit | Length-delimited | Start group | End group | 32-bit

type Field struct {
	CountRequirement string // repeated | required | optional
	Type             string
	Name             string
	Values           []interface{}
}

type Message struct {
	Name   string
	Fields map[int]protobuf
}

func getVarInt(reader *bufio.Reader) (int64, error) {
	bytes := []byte{}
	for {
		curByte, err := reader.ReadByte()
		if err != nil {
			return -1, err
		}
		bytes = append(bytes, curByte&((1<<7)-1))
		if curByte&(1<<7) == 0 {
			break
		}
	}
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	var res int64 = 0
	for _, b := range bytes {
		res = (res << 7) | int64(b)
	}
	return res, nil
}

func get64bit(reader *bufio.Reader) (int64, error) {
	var n int64 = 0
	for i := 0; i < 8; i++ {
		curByte, err := reader.ReadByte()
		if err != nil {
			return -1, err
		}
		n = (n << 8) | int64(curByte)
	}

	return (n >> 1) ^ (n << 63), nil
}

func getLengthDelimited(reader *bufio.Reader) ([]byte, error) {
	length, err := reader.ReadByte()
	result := []byte{}
	if err != nil {
		return result, err
	}
	for i := 0; i < int(length); i++ {
		curByte, err := reader.ReadByte()
		if err != nil {
			return result, err
		}
		result = append(result, curByte)
	}
	return result, nil
}

func get32bit(reader *bufio.Reader) (int32, error) {
	var n int32 = 0
	for i := 0; i < 4; i++ {
		curByte, err := reader.ReadByte()
		if err != nil {
			return -1, err
		}
		n = (n << 8) | int32(curByte)
	}

	return (n >> 1) ^ (n << 31), nil
}

func (m *Message) decode(reader *bufio.Reader) bool {
	for {
		tag, err := getVarInt(reader)
		if err != nil {
			log.Fatal(err)
		}
		field_number := tag >> 3
		if _, exists := m.Fields[int(field_number)]; !exists {
			return false
		}
		if !m.Fields[int(field_number)].decode(reader) {
			return false
		}
	}
	return true
}

func (f *Field) decode(reader *bufio.Reader) bool {
	var err error
	var val interface{}
	switch f.Type {
	case "Varint":
		val, err = getVarInt(reader)
	case "64-bit":
		val, err = get64bit(reader)
	case "Length-delimited":
		val, err = getLengthDelimited(reader)
	case "32-bit":
		val, err = get32bit(reader)
	}
	f.Values = append(f.Values, val)
	return err != nil
}

func (m *Message) filled() bool {
	for _, field := range m.Fields {
		if !field.filled() {
			return false
		}
	}
	return true
}

func (f *Field) filled() bool {
	curCount := len(f.Values)
	if f.CountRequirement == "required" {
		return curCount == 1
	} else if f.CountRequirement == "optional" {
		return curCount <= 1
	}
	return true
}

func read(fname string) *bufio.Reader {
	f, err := os.Open(fname)
	if err != nil {
		return nil
	}
	return bufio.NewReader(f)
}

func readFieldFromFile(scanner *bufio.Scanner) (*Field, int) {
	tokens := strings.Split(strings.TrimSpace(scanner.Text()), " ")
	field := Field{
		CountRequirement: tokens[0],
		Name:             tokens[2],
		Type:             tokens[1],
	}
	field_number := tokens[4]
	field_number = field_number[:len(field_number)-1]
	log.Print(field_number)
	field_number_int, err := strconv.Atoi(field_number)
	if err != nil {
		log.Fatal("could not read field_number ", scanner.Text())
	}
	return &field, field_number_int
}

func readMessageFromFile(scanner *bufio.Scanner) *Message {
	text := scanner.Text()
	name := strings.Split(text, " ")[1]
	message := Message{
		Name:   name,
		Fields: make(map[int]protobuf),
	}

	scanner.Scan() // {
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "}" {
			break
		}
		field, fieldNumber := readFieldFromFile(scanner)
		message.Fields[fieldNumber] = field
	}
	return &message
}

func readFile(scanner *bufio.Scanner) *Message {
	if !scanner.Scan() {
		log.Fatal("no proto syntax found")
	}
	messages := []*Message{}
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "import") {

		} else if strings.HasPrefix(text, "message") {
			messages = append(messages, readMessageFromFile(scanner))
		} else {
			log.Fatal("error occured while parsing proto file (unkown line type)")
		}
	}
	return messages[0]
}

func main() {
	filePath := "50/proto/teams.proto"
	binPath := "50/pb/example4.pb"
	fmt.Scanf("%s %s", &filePath, &binPath)

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("could not ope file")
	}
	message := readFile(bufio.NewScanner(f))
	log.Print(message)

	f, err = os.Open(binPath)
	if err != nil {
		log.Fatal("could not open bin file", err)
	}
	log.Print(message.decode(bufio.NewReader(f)))
}
