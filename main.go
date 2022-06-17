package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/syzkrash/skol/lexer"
)

var Sanitizer = strings.NewReplacer(
	"\"", "\\\"",
	"\n", "\\n",
	"\r", "\\r",
	"\t", "\\t")

const Version = 1.01
const For = 0.3

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <input file>\nOptionally, use -v to view version information.\n", os.Args[0])
		os.Exit(1)
	}
	if os.Args[1] == "-v" {
		fmt.Printf("skol-minify %.2f for skol %.1f\n", Version, For)
		return
	}

	inf, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("could not open input file: %s", err)
		os.Exit(2)
	}
	defer inf.Close()
	outf, err := os.Create("mini_" + os.Args[1])
	if err != nil {
		fmt.Printf("coult not create output file: %s", err)
		os.Exit(3)
	}
	defer outf.Close()
	code, err := io.ReadAll(inf)
	if err != nil {
		fmt.Printf("could not read input file: %s", err)
		os.Exit(4)
	}
	src := bytes.NewReader(code)
	lex := lexer.NewLexer(src, os.Args[1])
	var prev *lexer.Token
	for {
		tk, err := lex.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			var lerr *lexer.LexerError
			if errors.As(err, &lerr) {
				fmt.Printf("At %s: ", lerr.Where)
			}
			fmt.Println(err)
			os.Exit(5)
		}

		switch tk.Kind {
		case lexer.TkIdent:
			if prev != nil && prev.Kind == lexer.TkIdent {
				outf.Write([]byte{' '})
			}
			outf.WriteString(tk.Raw)
		case lexer.TkConstant, lexer.TkPunct:
			outf.WriteString(tk.Raw)
		case lexer.TkString:
			outf.Write([]byte{'"'})
			outf.WriteString(Sanitizer.Replace(tk.Raw))
			outf.Write([]byte{'"'})
		case lexer.TkChar:
			outf.Write([]byte{'\'', Sanitizer.Replace(tk.Raw)[0], '\''})
		}

		prev = tk
	}
}
