package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"strings"

	. "github.com/stevegt/goadapt"
)

func selfContained() {
	Pl("selfContained:")
	Pl()
	msg := &mail.Message{
		Header: map[string][]string{
			"Content-Type": {"multipart/mixed; boundary=foo"},
		},
		Body: strings.NewReader(
			"--foo\r\nFoo: one\r\n\r\nA section\r\n" +
				"--foo\r\nFoo: two\r\n\r\nAnd another\r\n" +
				"--foo--\r\n"),
	}
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
	fmt.Printf("Body: \n%v\n\n", msg.Body)
	fmt.Println()
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			slurp, err := io.ReadAll(p)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Part %q: %q\n", p.Header.Get("Foo"), slurp)
		}
	}
	Pl()
	Pl()
}

// external parses an external file using the given boundary and shows the parts on stdout.
func external(fn, boundary string) {
	Pl("external: ", fn, boundary)
	Pl()
	rd, err := os.Open(fn)
	Ck(err)
	mr := multipart.NewReader(rd, boundary)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		slurp, err := io.ReadAll(p)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("File %q:\n\n%v\n", p.Header.Get("File"), string(slurp))
	}
}

func main() {
	selfContained()
	external("example.mime", "apple")
}
