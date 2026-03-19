package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readImageViaOCR(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	ocrPath := path
	var cleanup func()

	if ext == ".tif" || ext == ".tiff" {
		convertedPath, err := convertTIFFToPNG(path)
		if err != nil {
			return "", err
		}
		ocrPath = convertedPath
		cleanup = func() { _ = os.Remove(convertedPath) }
	}

	if cleanup != nil {
		defer cleanup()
	}

	// 1. OCR всего листа
	fullText, fullErr := runTesseract(ocrPath)
	if fullErr != nil {
		fullText = ""
	}

	// 2. OCR нижнего блока
	bottomPath, err := cropBottomArea(ocrPath)
	if err == nil {
		defer os.Remove(bottomPath)

		bottomText, err := runTesseract(bottomPath)
		if err == nil {
			bottomText = strings.TrimSpace(bottomText)
			if looksLikeTechnicalBlock(bottomText) {
				return trimToMax(bottomText), nil
			}
			if len([]rune(bottomText)) > len([]rune(fullText)) && bottomText != "" {
				return trimToMax(bottomText), nil
			}
		}
	}

	fullText = strings.TrimSpace(fullText)
	if fullText == "" {
		return "", fmt.Errorf("image: extracted empty text")
	}

	return trimToMax(fullText), nil
}
