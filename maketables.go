// +build ignore

// Confusables table generator.
// See http://www.unicode.org/reports/tr39/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	flag.Parse()
	loadUnicodeData()
	makeTables()
}

var url = flag.String("url",
	"http://www.unicode.org/Public/security/latest/",
	"URL of Unicode database directory")

var localFiles = flag.Bool("local",
	false,
	"data files have been copied to the current directory; for debugging only")

// confusables.txt has form:
//	309C ;	030A ;	SL	#* ( ゜ → ̊ ) KATAKANA-HIRAGANA SEMI-VOICED SOUND MARK → COMBINING RING ABOVE	# →ﾟ→→゚→
// See http://www.unicode.org/reports/tr39/ for full explanation
// The fields:
const (
	CSourceCodePoint = iota
	CTargetCodePoint
	CType
	NumField

	MaxChar = 0x10FFFF // anything above this shouldn't exist
)

func openReader(file string) (input io.ReadCloser) {
	if *localFiles {
		f, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		input = f
	} else {
		path := *url + file
		log.Println("Downloading " + path)
		resp, err := http.Get(path)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			log.Fatal("bad GET status for "+path, resp.Status)
		}
		input = resp.Body
	}
	return
}

func parsePoint(pointString string, line string) rune {
	x, err := strconv.ParseUint(pointString, 16, 64)
	point := rune(x)
	if err != nil {
		log.Fatalf("%.5s...: %s", line, err)
	}
	if point == 0 {
		log.Fatalf("%5s: Unknown rune %X", line, point)
	}
	if point > MaxChar {
		log.Fatalf("%5s: Rune %X > MaxChar (%X)", line, point, MaxChar)
	}

	return point
}

// Type C encapsulates a line of the confusables.txt files
type C struct {
	k rune
	v []rune
}

var confusables []C

func parseCharacter(line string) {
	if len(line) == 0 || line[0] == '#' {
		return
	}
	field := strings.Split(line, " ;\t")
	if len(field) != NumField {
		log.Fatalf("%5s: %d fields (expected %d)\n", line, len(field), NumField)
	}

	if !strings.HasPrefix(field[2], "MA") {
		// The MA table is a superset anyway
		return
	}

	sourceRune := parsePoint(field[CSourceCodePoint], line)
	var targetRune []rune
	targetCodePoints := strings.Split(field[CTargetCodePoint], " ")
	for _, targetCP := range targetCodePoints {
		targetRune = append(targetRune, parsePoint(targetCP, line))
	}

	confusables = append(confusables, C{sourceRune, targetRune})
}

var originalHeader = []byte(`
// Following is the original header of the source confusables.txt file
//
`)

func parseHeader(line string) bool {
	// strip BOM
	if len(line) > 3 && bytes.Compare(([]byte(line[0:3])), []byte{0xEF, 0xBB, 0xBF}) == 0 {
		line = line[3:]
	}
	if len(line) == 0 || line[0] != '#' {
		return true
	}
	originalHeader = append(originalHeader, "//"+line[1:]+"\n"...)
	return false
}

func loadUnicodeData() {
	f := openReader("confusables.txt")
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if parseHeader(scanner.Text()) {
			break
		}
	}
	for scanner.Scan() {
		parseCharacter(scanner.Text())
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}
}

// Use the Unicode table with the following changes:
// - do not confuse "m" with "rn"
// - confuse "μ" (mu) with "u"
// - confuse "χ" (chi) with "x"
// - confuse "ʀ" (Latin small R) with "R"
// - confuse "ኮ" (Ethiopic syllabel Ko) with "r"
// - various additions as below
func makeTables() {
	out := fmt.Sprintf("%s\n", originalHeader)
	out += fmt.Sprint("var confusablesMap = map[rune]string{\n\n")
	for _, c := range confusables {
		if strings.Contains(string(c.v), "rn") {
			// Avoid m -> m loop
			if string(c.k) != "m" {
				out += fmt.Sprintf("0x%.8X: %+q,\n", c.k, strings.Replace(string(c.v), "rn", "m", -1))
			}
			continue
		} else if strings.Contains(string(c.v), "π") {
			out += fmt.Sprintf("0x%.8X: %+q,\n", c.k, strings.Replace(string(c.v), "π", "n", -1))
			continue
		} else if strings.Contains(string(c.v), "μ") {
			out += fmt.Sprintf("0x%.8X: %+q,\n", c.k, strings.Replace(string(c.v), "μ", "u", -1))
			continue
		} else if strings.Contains(string(c.v), "χ") {
			out += fmt.Sprintf("0x%.8X: %+q,\n", c.k, strings.Replace(string(c.v), "χ", "x", -1))
			continue
		} else if strings.Contains(string(c.v), "ʀ") {
			out += fmt.Sprintf("0x%.8X: %+q,\n", c.k, strings.Replace(string(c.v), "ʀ", "R", -1))
			continue
		} else if strings.Contains(string(c.v), "ኮ") {
			out += fmt.Sprintf("0x%.8X: %+q,\n", c.k, strings.Replace(string(c.v), "ኮ", "r", -1))
			continue
		} else {
			out += fmt.Sprintf("0x%.8X: %+q,\n", c.k, string(c.v))
		}
	}
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ᛒ', "B")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｂ', "b")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｄ', "D")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ḍ', "d")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｄ', "d")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｆ', "F")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｆ', "f")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｇ', "G")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｋ', "k")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｌ', "L")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｍ', "m")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ɴ', "N")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｎ', "n")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ⴍ', "Q")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ⴓ', "Q")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｑ', "Q")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｑ', "q")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｒ', "R")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ʀ', "R")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ᚱ', "R")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｒ', "r")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ⴝ', "S")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｔ', "t")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ա', "U")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｕ', "U")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｕ', "u")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｖ', "V")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'Ｗ', "W")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｗ', "w")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'ｚ', "z")
	out += fmt.Sprintf("0x%.8X: %+q,\n", 'π', "n")
	out += fmt.Sprintln("}")

	WriteGoFile("tables.go", "confusables", []byte(out))
}

const header = `// This file was generated by go generate; DO NOT EDIT

package %s

`

func WriteGoFile(filename, pkg string, b []byte) {
	w, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create file %s: %v", filename, err)
	}
	defer w.Close()
	_, err = fmt.Fprintf(w, header, pkg)
	if err != nil {
		log.Fatalf("Error writing header: %v", err)
	}
	// Strip leading newlines.
	for len(b) > 0 && b[0] == '\n' {
		b = b[1:]
	}
	formatted, err := format.Source(b)
	if err != nil {
		// Print the original buffer even in case of an error so that the
		// returned error can be meaningfully interpreted.
		w.Write(b)
		log.Fatalf("Error formatting file %s: %v", filename, err)
	}
	if _, err := w.Write(formatted); err != nil {
		log.Fatalf("Error writing file %s: %v", filename, err)
	}
}
