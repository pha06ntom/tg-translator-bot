package files

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	pdf "github.com/ledongthuc/pdf"
)

func readPDF(path string) (string, error) {
	text, err := readPDFTextLayer(path)
	if err == nil && strings.TrimSpace(text) != "" {
		return trimToMax(text), nil
	}
	ocrText, ocrErr := readPDFViaOCR(path)
	if ocrErr != nil {
		if err != nil {
			return "", fmt.Errorf("pdf text extraction failed: %v; ocr failed: %v", err, ocrErr)
		}
		return "", ocrErr
	}

	ocrText = strings.TrimSpace(ocrText)
	if ocrText == "" {
		return "", fmt.Errorf("pdf: extracted empty text after OCR")
	}

	return trimToMax(ocrText), nil
}

func readPDFTextLayer(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(&buf, b); err != nil {
		return "", err
	}

	out := strings.TrimSpace(buf.String())
	if out == "" {
		return "", fmt.Errorf("pdf: extracted empty text")
	}
	return out, nil
}

func readPDFViaOCR(path string) (string, error) {
	workDir, err := os.MkdirTemp("", "pdfocr_*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)

	prefix := filepath.Join(workDir, "page")

	cmd := exec.Command("pdftocairo", "-png", "-r", "300", path, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pdftocairo: %v: %s", err, strings.TrimSpace(string(out)))
	}

	matches, err := filepath.Glob(filepath.Join(workDir, "page-*.png"))
	if err != nil {
		return "", err
	}
	sort.Strings(matches)

	if len(matches) == 0 {
		return "", fmt.Errorf("ocr: no rendered pages found")
	}

	var sb strings.Builder
	for i, imgPath := range matches {
		txt, err := runTesseract(imgPath)
		if err != nil {
			return "", err
		}
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(strings.TrimSpace(txt))
	}

	return sb.String(), nil
}
