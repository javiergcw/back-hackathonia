package rag

import (
	"encoding/json"
	"os"
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
	return c.RetrieveForClient(query, nil, k).Chunks
}

func (c *Client) RetrieveForClient(query string, profile *domain.Profile, k int) RetrieveResult {
	if len(c.chunks) == 0 {
		return RetrieveResult{}
	}

	queryNorm := normalize(query)
	queryTerms := expandTerms(tokenize(queryNorm))
	profileTags := ProfileBoostTags(profile)

	scored := make([]struct {
		chunk *domain.Chunk
		score float64
	}, len(c.chunks))

	for i := range c.chunks {
		score := c.scoreChunk(&c.chunks[i], queryTerms, profileTags)
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

	topScore := 0.0
	if len(scored) > 0 {
		topScore = scored[0].score
	}

	result := make([]domain.Chunk, 0, k)
	for i := 0; i < k && i < len(scored); i++ {
		if scored[i].score > 0 {
			result = append(result, *scored[i].chunk)
		}
	}

	return RetrieveResult{Chunks: result, TopScore: topScore}
}

func (c *Client) scoreChunk(chunk *domain.Chunk, queryTerms []string, profileTags []string) float64 {
	var score float64

	for _, term := range queryTerms {
		for _, tag := range chunk.Tags {
			tagLower := strings.ToLower(tag)
			if strings.Contains(tagLower, term) || strings.Contains(term, tagLower) {
				score += 2.0
			}
		}
	}

	contentNorm := normalize(chunk.Contenido)
	for _, term := range queryTerms {
		if len(term) < 3 {
			continue
		}
		count := countOccurrences(contentNorm, term)
		score += float64(count) * 1.5
	}

	if containsSuperCDT(queryTerms) && hasSuperCDTTag(chunk.Tags) {
		score += 5.0
	}

	for _, pTag := range profileTags {
		for _, tag := range chunk.Tags {
			if strings.Contains(strings.ToLower(tag), pTag) {
				score += 3.0
				break
			}
		}
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
	if term == "" {
		return 0
	}
	lower := strings.ToLower(text)
	needle := strings.ToLower(term)
	n := 0
	for {
		i := strings.Index(lower, needle)
		if i < 0 {
			return n
		}
		n++
		start := i + len(needle)
		if start >= len(lower) {
			return n
		}
		lower = lower[start:]
	}
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