package rag

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/javierg/hackathon-bqia/internal/domain"
	"github.com/ledongthuc/pdf"
)

func ExtractText(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return extractPDFReal(filePath)
	case ".txt", ".md":
		return extractTextFile(filePath)
	default:
		return "", nil
	}
}

func extractPDFReal(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	totalPages := r.NumPage()
	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		if i < totalPages {
			buf.WriteString("\n\n")
		}
	}

	return strings.TrimSpace(buf.String()), nil
}

func extractTextFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getDocType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return "PDF"
	case ".txt":
		return "TXT"
	case ".md":
		return "MD"
	default:
		return "DOC"
	}
}

func ChunkText(text string, doc string, maxChunkSize int, overlap int) []domain.Chunk {
	docType := getDocType(doc)

	if maxChunkSize <= 0 {
		maxChunkSize = 500
	}
	if overlap < 0 {
		overlap = maxChunkSize / 7
	}
	if overlap >= maxChunkSize {
		overlap = maxChunkSize / 4
	}

	defaultSection := "[" + docType + "] Contenido General"

	paragraphs := splitIntoParagraphs(text)
	var chunks []domain.Chunk
	currentSection := defaultSection
	var currentContent strings.Builder

	flushChunk := func(section string, prevTail string) {
		contentStr := strings.TrimSpace(currentContent.String())
		if contentStr == "" {
			return
		}
		if prevTail != "" {
			contentStr = prevTail + " " + contentStr
		}
		id := generateChunkID(doc, section, len(chunks))
		tags := extractTags(contentStr)

		chunks = append(chunks, domain.Chunk{
			ID:        id,
			Doc:       doc,
			Seccion:   section,
			Tags:      tags,
			Contenido: contentStr,
		})
		currentContent.Reset()
	}

	getTail := func() string {
		content := currentContent.String()
		if len(content) <= overlap {
			return content
		}
		tailStart := len(content) - overlap
		for i := tailStart; i < len(content); i++ {
			if content[i] == ' ' {
				return content[i+1:]
			}
		}
		return content[tailStart:]
	}

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		if isHeader(para) {
			if currentContent.Len() > 0 {
				tail := getTail()
				flushChunk(currentSection, tail)
			}
			currentSection = "[" + docType + "] " + extractHeaderTitle(para)
			continue
		}

		paraLen := len(para) + 1
		if currentContent.Len() == 0 {
			currentContent.WriteString(para)
		} else if currentContent.Len()+paraLen <= maxChunkSize {
			currentContent.WriteString(" ")
			currentContent.WriteString(para)
		} else {
			sentences := splitIntoSentences(para)
			for _, sentence := range sentences {
				sentence = strings.TrimSpace(sentence)
				if sentence == "" {
					continue
				}
				sentenceLen := len(sentence) + 1
				if currentContent.Len() == 0 {
					currentContent.WriteString(sentence)
				} else if currentContent.Len()+sentenceLen <= maxChunkSize {
					currentContent.WriteString(" ")
					currentContent.WriteString(sentence)
				} else {
					tail := getTail()
					flushChunk(currentSection, tail)
					currentContent.WriteString(sentence)
				}
			}
		}
	}

	if currentContent.Len() > 0 {
		tail := getTail()
		flushChunk(currentSection, tail)
	}

	return chunks
}

func splitIntoSentences(text string) []string {
	re := regexp.MustCompile(`([.!?]+[\s]+)`)
	parts := re.Split(text, -1)
	var sentences []string
	var current strings.Builder

	for _, part := range parts {
		if current.Len() == 0 {
			current.WriteString(strings.TrimSpace(part))
		} else {
			current.WriteString(part)
			sentences = append(sentences, current.String())
			current.Reset()
		}
	}

	if current.Len() > 0 {
		sentences = append(sentences, current.String())
	}

	var result []string
	for _, s := range sentences {
		if len(s) <= 500 {
			result = append(result, s)
		} else {
			for len(s) > 0 {
				if len(s) <= 500 {
					result = append(result, s)
					break
				}
				cutAt := 500
				for cutAt > 0 && s[cutAt] != ' ' {
					cutAt--
				}
				if cutAt == 0 {
					cutAt = 500
				}
				result = append(result, s[:cutAt])
				s = strings.TrimSpace(s[cutAt:])
			}
		}
	}

	return result
}

func splitIntoParagraphs(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var result []string
	var current strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
			continue
		}

		if current.Len() == 0 {
			current.WriteString(trimmed)
		} else {
			current.WriteString(" ")
			current.WriteString(trimmed)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

func isHeader(line string) bool {
	line = strings.TrimSpace(line)
	if len(line) < 3 || len(line) > 100 {
		return false
	}

	if unicode.IsDigit(rune(line[0])) && strings.Contains(line, ".") {
		prefix := strings.SplitN(line, ".", 2)
		if _, err := strconv.Atoi(strings.TrimSpace(prefix[0])); err == nil {
			return true
		}
	}

	upperCount := 0
	for _, r := range line {
		if unicode.IsUpper(r) {
			upperCount++
		}
	}
	if upperCount > len(line)/2 && !strings.Contains(line, " ") {
		return true
	}

	headerPattern := regexp.MustCompile(`^[A-ZÁÉÍÓÚ0-9\s\-:]+$`)
	if headerPattern.MatchString(line) {
		return true
	}

	if isTitleCase(line) && len(line) < 80 && !strings.Contains(line, " ") {
		return true
	}

	if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "===") || strings.HasPrefix(line, "###") {
		return true
	}

	romanNumerals := regexp.MustCompile(`^(I{1,3}|IV|V|VI{0,3})\.\s*`)
	if romanNumerals.MatchString(line) {
		return true
	}

	if len(line) > 3 && !strings.HasSuffix(line, ".") && !strings.HasSuffix(line, ":") {
		upperCount := 0
		for _, r := range line {
			if unicode.IsUpper(r) {
				upperCount++
			}
		}
		if upperCount > len(line)/2 && upperCount > 3 {
			return true
		}
	}

	return false
}

func isTitleCase(line string) bool {
	if len(line) < 3 {
		return false
	}

	firstRunes := []rune(line)
	firstUpper := unicode.IsUpper(firstRunes[0])

	spaceCount := strings.Count(line, " ")
	if spaceCount < 2 {
		return firstUpper && len(line) < 50
	}

	upperCount := 0
	wordStart := true
	for _, r := range line {
		if wordStart {
			if unicode.IsUpper(r) {
				upperCount++
			}
			wordStart = false
		}
		if r == ' ' || r == '-' {
			wordStart = true
		}
	}

	return firstUpper && upperCount >= spaceCount/2+1
}

func extractHeaderTitle(line string) string {
	line = strings.TrimSpace(line)
	line = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(line, "")
	line = regexp.MustCompile(`^---+|---+$`).ReplaceAllString(line, "")
	line = regexp.MustCompile(`^=+$|^=+$`).ReplaceAllString(line, "")
	line = regexp.MustCompile(`^#+|`).ReplaceAllString(line, "")
	return strings.TrimSpace(line)
}

func extractTags(content string) []string {
	content = strings.ToLower(content)
	content = regexp.MustCompile(`[^\w\s]`).ReplaceAllString(content, " ")

	words := strings.Fields(content)

	stopWords := map[string]bool{
		"el": true, "la": true, "los": true, "las": true,
		"de": true, "del": true, "en": true, "y": true, "a": true,
		"que": true, "un": true, "una": true,
		"por": true, "con": true, "para": true, "se": true, "su": true,
		"este": true, "esta": true, "como": true, "mas": true, "pero": true,
		"si": true, "no": true, "ya": true, "o": true, "al": true,
	}

	wordFreq := make(map[string]int)
	for _, word := range words {
		if len(word) < 4 || stopWords[word] {
			continue
		}
		wordFreq[word]++
	}

	type wordCount struct {
		word  string
		count int
	}
	var sorted []wordCount
	for w, c := range wordFreq {
		if c >= 2 {
			sorted = append(sorted, wordCount{w, c})
		}
	}

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var tags []string
	for i := 0; i < len(sorted) && i < 5; i++ {
		tags = append(tags, sorted[i].word)
	}

	return tags
}

func generateChunkID(doc string, section string, index int) string {
	docName := strings.TrimSuffix(filepath.Base(doc), filepath.Ext(doc))
	docName = regexp.MustCompile(`[^\w]`).ReplaceAllString(docName, "_")

	sectionSlug := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, section)
	sectionSlug = strings.Trim(sectionSlug, "_")

	if sectionSlug == "" {
		sectionSlug = "general"
	}

	return docName + "_" + sectionSlug + "_" + strconv.Itoa(index)
}