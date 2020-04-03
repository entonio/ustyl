package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/text/unicode/norm"
)

func main() {

	codes := flag.String("styles", asCodes(allStyles...), "the styles to use:"+asDescriptions(allStyles...))
	text := flag.String("text", "", "[alternative] the text to convert")
	file := flag.String("file", "", "[alternative] the file to convert")
	nfd := flag.Bool("nfd", false, "convert letters with diacritics")

	flag.Parse()

	styles := fromCodes(*codes)

	if len(*text) > 0 && len(*file) == 0 {
		doText(*text, *nfd, styles...)
		return
	}

	if len(*text) == 0 && len(*file) > 0 {
		doFile(*file, *nfd, styles...)
		return
	}

	fmt.Fprintln(os.Stderr, "One and only one of -text or -file must be specified")
	os.Exit(1)
}

type Style struct {
	Name   string
	Code   rune
	UpperA rune
	LowerA rune
}

var allStyles = []Style{
	Style{Name: "        Serif normal      ", Code: 'e', UpperA: 'A', LowerA: 'a'},
	Style{Name: "        Serif bold        ", Code: 'E', UpperA: 0x1D400, LowerA: 0x1D41A},
	Style{Name: "        Serif italic      ", Code: 'i', UpperA: 0x1D434, LowerA: 0x1D44E},
	Style{Name: "        Serif bold italic ", Code: 'I', UpperA: 0x1D468, LowerA: 0x1D482},
	Style{Name: "   Sans-serif normal      ", Code: 'a', UpperA: 0x1D5A0, LowerA: 0x1D5BA},
	Style{Name: "   Sans-serif bold        ", Code: 'A', UpperA: 0x1D5D4, LowerA: 0x1D5EE},
	Style{Name: "   Sans-serif italic      ", Code: 'j', UpperA: 0x1D608, LowerA: 0x1D622},
	Style{Name: "   Sans-serif bold italic ", Code: 'J', UpperA: 0x1D63C, LowerA: 0x1D656},
	Style{Name: "  Calligraphy normal      ", Code: 'c', UpperA: 0x1D49C, LowerA: 0x1D4B6},
	Style{Name: "  Calligraphy bold        ", Code: 'C', UpperA: 0x1D4D0, LowerA: 0x1D4EA},
	Style{Name: "      Fraktur normal      ", Code: 'f', UpperA: 0x1D504, LowerA: 0x1D51E},
	Style{Name: "      Fraktur bold        ", Code: 'F', UpperA: 0x1D56C, LowerA: 0x1D586},
	Style{Name: "    Monospace             ", Code: 'm', UpperA: 0x1D670, LowerA: 0x1D68A},
	Style{Name: " Doublestruck             ", Code: 'd', UpperA: 0x1D538, LowerA: 0x1D552},
}

func fromCodes(codes string) (styles []Style) {
	for _, code := range codes {
		for _, style := range allStyles {
			if code == style.Code {
				styles = append(styles, style)
			}
		}
	}
	return
}

func asCodes(styles ...Style) (s string) {
	for _, style := range styles {
		s += string(style.Code)
	}
	return
}

func asDescriptions(styles ...Style) (s string) {
	for _, style := range styles {
		s += "\n\t" + string(style.Code) + " " + strings.TrimSpace(style.Name)
	}
	return
}

func toStyle(input string, style Style) string {
	output := ""
	for _, r := range input {
		var c rune
		if 'A' <= r && r <= 'Z' {
			c = r + style.UpperA - 'A'
		} else if 'a' <= r && r <= 'z' {
			c = r + style.LowerA - 'a'
		} else {
			c = r
		}
		output += string(c)
	}
	return output
}

func doText(text string, nfd bool, styles ...Style) {
	if nfd {
		text = norm.NFD.String(text)
	}
	for _, style := range styles {
		fmt.Println(toStyle(text, style))
	}
}

func doFile(path string, nfd bool, styles ...Style) {

	var writers []*bufio.Writer
	var outputFiles []*os.File

	inputFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer func() {
		inputFile.Close()
		for _, w := range writers {
			w.Flush()
		}
		for _, f := range outputFiles {
			f.Close()
		}
	}()

	for _, style := range styles {
		var ext string
		if style.Code >= 'a' {
			ext = string(style.Code)
		} else {
			ext = string(style.Code + 'a' - 'A')
			ext += ext
		}
		f, err := os.Create(path + "." + ext + ".ustyl")
		if err != nil {
			panic(err)
		}
		outputFiles = append(outputFiles, f)
		w := bufio.NewWriter(f)
		writers = append(writers, w)
		// utf8 bom
		w.Write([]byte{0xEF, 0xBB, 0xBF})
	}

	reader := bufio.NewReader(inputFile)

	var line string
	var eof bool
	for !eof {
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			eof = true
		} else if err != nil {
			panic(err)
		}

		if nfd {
			line = norm.NFD.String(line)
		}

		for i, style := range styles {
			_, err = writers[i].Write([]byte(toStyle(line, style)))
			if err != nil {
				panic(err)
			}
		}
	}
}
