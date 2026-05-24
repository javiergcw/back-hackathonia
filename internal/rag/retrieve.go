package rag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/javierg/hackathon-bqia/internal/domain"
)

type Client struct {
	chunks     []domain.Chunk
	profiles   []domain.Profile
	scope      domain.Scope
	mu         sync.RWMutex
	filePath   string
}

func NewRetrieve(path string) *Client {
	c := &Client{
		filePath: path,
		scope: domain.Scope{
			StrictMode: true,
			ActiveDocs: []string{},
		},
	}
	c.loadFromFile()
	return c
}

func (c *Client) loadFromFile() {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &c.chunks); err != nil {
		c.chunks = []domain.Chunk{}
	}
}

func (c *Client) saveToFile() error {
	data, err := json.MarshalIndent(c.chunks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.filePath, data, 0644)
}

func (c *Client) Retrieve(query string, k int) []domain.Chunk {
	return c.RetrieveWithTags(query, k, nil)
}

const MIN_SCORE_THRESHOLD = 2

func (c *Client) RetrieveWithTags(query string, k int, allowedTags []string) []domain.Chunk {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.chunks) == 0 {
		return nil
	}

	retrieveQ, intent := ResolveQuery(query)
	queryNorm := normalize(query)
	retrieveNorm := normalize(retrieveQ)
	queryTerms := mergeTerms(
		expandTerms(tokenize(queryNorm)),
		expandTerms(tokenize(retrieveNorm)),
	)

	filteredChunks := c.chunks

	if len(c.scope.ActiveDocs) > 0 {
		filteredChunks = make([]domain.Chunk, 0)
		for _, chunk := range c.chunks {
			if isDocActive(chunk.Doc, c.scope.ActiveDocs) {
				filteredChunks = append(filteredChunks, chunk)
			}
		}
		if len(filteredChunks) == 0 {
			filteredChunks = c.chunks
		}
	}

	if allowedTags != nil && len(allowedTags) > 0 {
		filteredChunks = c.filterByTags(filteredChunks, allowedTags)
	}

	scored := make([]struct {
		chunk *domain.Chunk
		score float64
	}, len(filteredChunks))

	for i := range filteredChunks {
		score := c.scoreChunk(&filteredChunks[i], queryTerms, intent)
		scored[i] = struct {
			chunk *domain.Chunk
			score float64
		}{&filteredChunks[i], score}
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
		if scored[i].score >= MIN_SCORE_THRESHOLD && !IsBoilerplate(scored[i].chunk.Contenido) {
			result = append(result, *scored[i].chunk)
		}
	}

	if len(result) == 0 && intent.ID != "" {
		for i := 0; i < k && i < len(scored); i++ {
			if scored[i].score > 0 && !IsBoilerplate(scored[i].chunk.Contenido) {
				result = append(result, *scored[i].chunk)
			}
		}
	}

	return result
}

func mergeTerms(a, b []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, list := range [][]string{a, b} {
		for _, t := range list {
			if !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	return out
}

func (c *Client) filterByTags(chunks []domain.Chunk, allowedTags []string) []domain.Chunk {
	var result []domain.Chunk
	for _, chunk := range chunks {
		if chunkAllowedForTags(allowedTags, chunk) {
			result = append(result, chunk)
		}
	}
	return result
}

func (c *Client) scoreChunk(chunk *domain.Chunk, queryTerms []string, intent QueryIntent) float64 {
	if IsBoilerplate(chunk.Contenido) {
		return -20
	}

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
	seccionNorm := normalize(chunk.Seccion)
	for _, term := range queryTerms {
		if len(term) < 3 {
			continue
		}
		count := countOccurrences(contentNorm, term)
		score += float64(count)
		if strings.Contains(seccionNorm, term) {
			score += 3
		}
	}

	docLower := strings.ToLower(chunk.Doc)
	for _, hint := range intent.DocHints {
		if strings.Contains(docLower, strings.ToLower(hint)) {
			score += 8
		}
	}
	for _, kw := range intent.Keywords {
		if strings.Contains(contentNorm, strings.ToLower(kw)) {
			score += 2
		}
	}

	if containsSuperCDT(queryTerms) && hasSuperCDTTag(chunk.Tags) {
		score += 5.0
	}
	if intent.ID == "cdt_beneficios" && strings.Contains(docLower, "03_cdt") {
		score += 6
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

func isDocActive(doc string, activeDocs []string) bool {
	for _, active := range activeDocs {
		if strings.Contains(doc, active) {
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

func (c *Client) ListKnowledge() []domain.KnowledgeItem {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := make([]domain.KnowledgeItem, len(c.chunks))
	for i, chunk := range c.chunks {
		items[i] = domain.KnowledgeItem{
			ID:        chunk.ID,
			Doc:       chunk.Doc,
			Seccion:   chunk.Seccion,
			Tags:      chunk.Tags,
			Contenido: chunk.Contenido,
		}
	}
	return items
}

func (c *Client) AddKnowledge(req domain.AddKnowledgeRequest) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := generateID(req.Doc, req.Seccion)

	for _, chunk := range c.chunks {
		if chunk.ID == id {
			return "", fmt.Errorf("chunk ya existe para ese doc y sección")
		}
	}

	newChunk := domain.Chunk{
		ID:        id,
		Doc:       req.Doc,
		Seccion:   req.Seccion,
		Tags:      req.Tags,
		Contenido: req.Contenido,
	}

	c.chunks = append(c.chunks, newChunk)

	if err := c.saveToFile(); err != nil {
		c.chunks = c.chunks[:len(c.chunks)-1]
		return "", err
	}

	return id, nil
}

func (c *Client) AddChunk(chunk domain.Chunk) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chunks = append(c.chunks, chunk)

	return c.saveToFile()
}

func (c *Client) UpdateKnowledge(id string, req domain.UpdateKnowledgeRequest) (string, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var before string
	found := false

	for i, chunk := range c.chunks {
		if chunk.ID == id {
			before = chunk.Contenido
			if req.Contenido != "" {
				c.chunks[i].Contenido = req.Contenido
			}
			if req.Tags != nil {
				c.chunks[i].Tags = req.Tags
			}
			found = true
			break
		}
	}

	if !found {
		return "", "", fmt.Errorf("chunk no encontrado")
	}

	if err := c.saveToFile(); err != nil {
		return "", "", err
	}

	return id, before, nil
}

func (c *Client) DeleteKnowledge(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	found := false
	var idx int
	for i, chunk := range c.chunks {
		if chunk.ID == id {
			found = true
			idx = i
			break
		}
	}

	if !found {
		return fmt.Errorf("chunk no encontrado")
	}

	c.chunks = append(c.chunks[:idx], c.chunks[idx+1:]...)

	return c.saveToFile()
}

func (c *Client) RemoveChunksByDoc(docName string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	initial := len(c.chunks)
	var remaining []domain.Chunk
	for _, chunk := range c.chunks {
		if chunk.Doc != docName {
			remaining = append(remaining, chunk)
		}
	}
	c.chunks = remaining

	if err := c.saveToFile(); err != nil {
		return 0, err
	}
	return initial - len(c.chunks), nil
}

func (c *Client) ClearAllChunks() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := len(c.chunks)
	c.chunks = []domain.Chunk{}

	if err := c.saveToFile(); err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Client) ReloadKnowledge() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.loadFromFile()

	return len(c.chunks), nil
}

func (c *Client) GetScope() domain.Scope {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return domain.Scope{
		StrictMode: c.scope.StrictMode,
		ActiveDocs:  c.scope.ActiveDocs,
	}
}

func (c *Client) SetScope(req domain.SetScopeRequest) domain.Scope {
	c.mu.Lock()
	defer c.mu.Unlock()

	if req.StrictMode != nil {
		c.scope.StrictMode = *req.StrictMode
	}
	if req.ActiveDocs != nil {
		c.scope.ActiveDocs = req.ActiveDocs
	}

	return domain.Scope{
		StrictMode: c.scope.StrictMode,
		ActiveDocs:  c.scope.ActiveDocs,
	}
}

func generateID(doc, seccion string) string {
	docName := strings.TrimSuffix(filepath.Base(doc), filepath.Ext(doc))
	seccionSlug := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, seccion)
	return docName + "_" + seccionSlug
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

func (c *Client) GetAllProfiles() []domain.Profile {
	if len(profiles) == 0 {
		c.LoadProfiles("data/profiles.json")
	}
	return profiles
}

func (c *Client) ListChunks() []domain.Chunk {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]domain.Chunk, len(c.chunks))
	copy(result, c.chunks)
	return result
}