package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

type Field struct {
	CountRequirement string // repeated | required | optional
	Type             string
	GlobalType       string
	Name             string
	Values           []interface{}
	Messages         []Message
}

type Message struct {
	Name   string
	Fields map[int]*Field
}

var TypeDict map[string]string
var messages map[string]Message

const MAX_BYTES = int(1e9)

func getVarInt(reader *bufio.Reader) (int64, int) {
	bytes := []byte{}
	for {
		curByte, err := reader.ReadByte()
		if err != nil {
			return -1, 0
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
	return res, len(bytes)
}

func get64bit(reader *bufio.Reader) (float64, int) {
	bytes := []byte{}
	for i := 0; i < 8; i++ {
		curByte, err := reader.ReadByte()
		if err != nil {
			return -1, 0
		}
		bytes = append(bytes, curByte)
	}
	bits := binary.LittleEndian.Uint64(bytes)
	return math.Float64frombits(bits), 8
}

func getLengthDelimited(reader *bufio.Reader) (string, int) {
	length, err := reader.ReadByte()
	result := []byte{}
	if err != nil {
		return "", 0
	}
	for i := 0; i < int(length); i++ {
		curByte, err := reader.ReadByte()
		if err != nil {
			return "", 0
		}
		result = append(result, curByte)
	}
	return string(result), int(length) + 1
}

func get32bit(reader *bufio.Reader) (float32, int) {
	bytes := []byte{}
	for i := 0; i < 4; i++ {
		curByte, err := reader.ReadByte()
		if err != nil {
			return -1, 0
		}
		bytes = append(bytes, curByte)
	}
	bits := binary.LittleEndian.Uint32(bytes)
	return math.Float32frombits(bits), 4
}

func (m *Message) decode(reader *bufio.Reader, bytesLeft int) bool {
	for {
		_, err := reader.Peek(1)
		if bytesLeft < 0 {
			return false
		}
		if bytesLeft == 0 || err != nil {
			return m.filled() // EOF
		}
		tag, len := getVarInt(reader)
		bytesLeft -= len
		if len == 0 {
			return false
		}
		field_number := tag >> 3
		globalType := tag & 7
		types := []string{"Varint", "64-bit", "Length-delimited", "Start group", "End group", "32-bit"}
		if _, exists := m.Fields[int(field_number)]; !exists {
			return false
		}
		if _, exists := messages[m.Fields[int(field_number)].Type]; !exists {
			if types[globalType] != TypeDict[m.Fields[int(field_number)].Type] {
				return false
			}
			m.Fields[int(field_number)].GlobalType = types[globalType]
		} else if globalType != 2 {
			return false // must be Length-delimited
		}
		bytesRead := m.Fields[int(field_number)].decode(reader, bytesLeft)
		if bytesRead == 0 {
			return false
		}
		bytesLeft -= bytesRead
	}
}

func copyField(field *Field) *Field {
	copy := *field
	copy.Messages = []Message{}
	copy.Values = make([]interface{}, 0)
	return &copy
}

func copyMessage(message Message) Message {
	res := Message{
		Name:   message.Name,
		Fields: make(map[int]*Field),
	}
	for key, field := range message.Fields {
		res.Fields[key] = copyField(field)
	}
	return res
}

func (f *Field) decode(reader *bufio.Reader, bytesLeft int) int {
	var val interface{}
	var len int
	switch f.GlobalType {
	case "Varint":
		val, len = getVarInt(reader)
	case "64-bit":
		val, len = get64bit(reader)
	case "Length-delimited":
		val, len = getLengthDelimited(reader)
	case "32-bit":
		val, len = get32bit(reader)
	default:
		message := copyMessage(messages[f.Type])
		lenB, err := reader.ReadByte()
		len = int(lenB) + 1
		if err != nil {
			return 0
		}
		if !message.decode(reader, len-1) {
			return 0
		}
		val, err = message, nil
		f.Messages = append(f.Messages, message)
	}
	f.Values = append(f.Values, val)
	if len > bytesLeft {
		return 0
	}
	return len
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
	field_number_int, err := strconv.Atoi(field_number)
	if err != nil {
		log.Fatal("could not read field_number ", scanner.Text())
	}
	return &field, field_number_int
}

func readMessageFromFile(scanner *bufio.Scanner) *Message {
	text := strings.TrimSpace(scanner.Text())
	name := strings.Split(text, " ")[1]
	message := Message{
		Name:   name,
		Fields: make(map[int]*Field),
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

func readImport(scanner *bufio.Scanner, baseDir string) {
	text := strings.TrimSpace(scanner.Text())
	tokens := strings.Split(text, " ")
	file := strings.Trim(tokens[1], "\";")
	f, err := os.Open(baseDir + file)
	if err != nil {
		log.Fatal("could not open file ", err, " ", baseDir+file)
	}
	readFile(bufio.NewScanner(f), baseDir)
}

func readFile(scanner *bufio.Scanner, baseDir string) *Message {
	if !scanner.Scan() {
		log.Fatal("no proto syntax found")
	}
	messageList := []*Message{}
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "import") {
			readImport(scanner, baseDir)
		} else if strings.HasPrefix(text, "message") {
			message := readMessageFromFile(scanner)
			messageList = append(messageList, message)
			messages[message.Name] = *message
		} else {
			log.Fatal("error occured while parsing proto file (unkown line type)")
		}
	}
	return messageList[len(messageList)-1]
}

func test(protoFile, binFile string) bool {
	messages = make(map[string]Message)

	f, err := os.Open(protoFile)
	if err != nil {
		log.Fatal("could not open file ", protoFile)
	}
	message := readFile(bufio.NewScanner(f), "50/proto/")

	f, err = os.Open(binFile)
	if err != nil {
		log.Fatal("could not open bin file", err)
	}
	if message.decode(bufio.NewReader(f), MAX_BYTES) {
		fmt.Println(protoFile, binFile)
		message.printFields()
		return true
	}
	return false
}

func (f *Field) print() {
	if _, exists := messages[f.Type]; exists {
		for _, message := range f.Messages {
			message.printFields()
		}
	} else {
		for _, val := range f.Values {
			fmt.Println(f.Name, ":", val)
		}
	}
}

func (m Message) printFields() {
	for _, field := range m.Fields {
		field.print()
	}
}

func getAllFilesInDir(dir string) []string {
	fileNames := []string{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}
	return fileNames
}

func init() {
	TypeDict = map[string]string{
		"string":   "Length-delimited",
		"bytes":    "Length-delimited",
		"int64":    "64-bit",
		"int":      "Varint",
		"int32":    "Varint",
		"uint32":   "Varint",
		"uint64":   "Varint",
		"sint32":   "Varint",
		"sint64":   "Varint",
		"bool":     "Varint",
		"enum":     "Varint",
		"fixed64":  "64-bit",
		"sfixed64": "64-bit",
		"double":   "64-bit",
		"fixed32":  "32-bit",
		"sfixed32": "32-bit",
		"float":    "32-bit",
	}
}

func main() {
	protos := getAllFilesInDir("50/proto/")
	pbs := getAllFilesInDir("50/pb/")
	for _, proto := range protos {
		for _, pb := range pbs {
			test("50/proto/"+proto, "50/pb/"+pb)
		}
	}
}
