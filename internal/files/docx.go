package files

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func readDocx(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var documentXML []byte
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			documentXML, err = io.ReadAll(rc)
			if err != nil {
				return "", err
			}
			break
		}
	}

	if len(documentXML) == 0 {
		return "", fmt.Errorf("docx: word/document.xml not found")
	}

	text, err := extractTextFromWordXML(documentXML)
	if err != nil {
		return "", err
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("docx: extracted empty text")
	}

	return trimToMax(text), nil
}

func extractTextFromWordXML(data []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var sb strings.Builder

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name.Local {
			case "t":
				var s string
				if err := decoder.DecodeElement(&s, &el); err != nil {
					return "", err
				}
				sb.WriteString(s)
			case "tab":
				sb.WriteString("\t")
			case "br":
				sb.WriteString("\n")
			case "p":
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
			}
		}
	}

	out := sb.String()
	lines := strings.Split(out, "\n")
	clean := make([]string, 0, len(lines))
	for _, line := range lines {
		clean = append(clean, strings.TrimRight(line, " \t"))
	}

	return strings.Join(clean, "\n"), nil
}
