package rag

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

type Client struct {
	chunks []domain.Chunk
}

func NewRetrieve(path string) *Client {
	c := &Client{}
	data, err := os.ReadFile(path)
	if err != nil {
		return c
	}
	if err := json.Unmarshal(data, &c.chunks); err != nil {
		return c
	}
	return c
}

func (c *Client) Retrieve(query string, k int) []domain.Chunk {
	if len(c.chunks) == 0 {
		return nil
	}

	queryNorm := normalize(query)
	queryTerms := tokenize(queryNorm)

	scored := make([]struct {
		chunk *domain.Chunk
		score float64
	}, len(c.chunks))

	for i := range c.chunks {
		score := c.scoreChunk(&c.chunks[i], queryTerms)
		scored[i] = struct {
			chunk *domain.Chunk
			score float64
		}{&c.chunks[i], score}
	}

	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	result := make([]domain.Chunk, 0, k)
	for i := 0; i < k && i < len(scored); i++ {
		if scored[i].score > 0 {
			result = append(result, *scored[i].chunk)
		}
	}

	if len(result) == 0 {
		return c.chunks[:min(k, len(c.chunks))]
	}

	return result
}

func (c *Client) scoreChunk(chunk *domain.Chunk, queryTerms []string) float64 {
	var score float64

	tagScore := 0
	for _, term := range queryTerms {
		for _, tag := range chunk.Tags {
			if strings.Contains(strings.ToLower(tag), term) {
				tagScore++
			}
		}
	}
	score += float64(tagScore) * 2.0

	contentNorm := normalize(chunk.Contenido)
	for _, term := range queryTerms {
		count := countOccurrences(contentNorm, term)
		score += float64(count)
	}

	if containsSuperCDT(queryTerms) && hasSuperCDTTag(chunk.Tags) {
		score += 5.0
	}

	return score
}

func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsMark(r) || unicode.IsPunct(r) {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

func tokenize(s string) []string {
	return strings.Fields(s)
}

func countOccurrences(text, term string) int {
	re := regexp.MustCompile(`(?i)` + term)
	return len(re.FindAllStringIndex(text, -1))
}

func containsSuperCDT(terms []string) bool {
	for _, t := range terms {
		if t == "supercdt" || t == "cdt" {
			return true
		}
	}
	return false
}

func hasSuperCDTTag(tags []string) bool {
	for _, tag := range tags {
		if tag == "supercdt" || tag == "cdt" {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var profiles []domain.Profile

func (c *Client) LoadProfiles(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &profiles)
}

func (c *Client) GetProfile(id string) (*domain.Profile, error) {
	if len(profiles) == 0 {
		c.LoadProfiles("data/profiles.json")
	}
	for i := range profiles {
		if profiles[i].ID == id {
			return &profiles[i], nil
		}
	}
	return nil, nil
}