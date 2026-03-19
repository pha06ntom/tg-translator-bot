package files

import (
	"path/filepath"
	"strings"
)

const maxExtractedChars = 200_000

func ExtractText(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".txt", ".md", ".csv", ".log":
		return readTextFile(path)
	case ".docx":
		return readDocx(path)
	case ".pdf":
		return readPDF(path)
	case ".png", ".jpg", ".jpeg", ".tif", ".tiff", ".bmp":
		return readImageViaOCR(path)
	default:
		return readTextFile(path)
	}
}

func looksLikeTechnicalBlock(text string) bool {
	t := strings.ToLower(text)

	hits := 0
	keywords := []string{
		"material",
		"coating",
		"gost",
		"dimensions",
		"weld",
		"wire",
		"thickness",
		"note",
		"technical",
		"техничес",
		"материал",
		"покрытие",
		"размер",
		"свар",
		"гост",
	}

	for _, k := range keywords {
		if strings.Contains(t, k) {
			hits++
		}
	}

	lines := strings.Split(text, "\n")
	nonEmpty := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}

	return hits >= 2 || nonEmpty >= 5
}
