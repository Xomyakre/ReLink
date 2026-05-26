package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"sync"
)

// ShortCodeGenerator generates unique, URL-safe short codes
// using cryptographically secure randomness.
type ShortCodeGenerator struct {
	charset       string
	length        int
	mu            sync.Mutex
	profanityList []string
	maxRetries    int
}

// NewShortCodeGenerator creates a new ShortCodeGenerator with the specified configuration.
// length: the length of generated short codes (must be between 6 and 10)
// If length is outside the valid range, it defaults to 7.
func NewShortCodeGenerator(length int) *ShortCodeGenerator {
	// Validate and set default length
	if length < 6 || length > 10 {
		length = 7
	}

	return &ShortCodeGenerator{
		charset:       "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		length:        length,
		profanityList: getDefaultProfanityList(),
		maxRetries:    10, // Maximum attempts to generate a non-profane code
	}
}

// getDefaultProfanityList returns a list of profane/offensive terms to block.
// This is a basic blocklist that can be extended or loaded from external sources.
func getDefaultProfanityList() []string {
	return []string{
		"fuck", "shit", "damn", "hell", "ass", "bitch", "crap",
		"piss", "dick", "cock", "pussy", "fag", "slut", "whore",
		"bastard", "cunt", "nigger", "nigga", "retard", "rape",
	}
}

// Generate creates a new random short code using crypto/rand for secure randomness.
// The generated code will contain only alphanumeric characters [a-zA-Z0-9].
// This method is thread-safe and includes profanity filtering.
// If a profane code is generated, it will retry up to maxRetries times.
func (g *ShortCodeGenerator) Generate() (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for attempt := 0; attempt < g.maxRetries; attempt++ {
		code, err := g.generateRandom()
		if err != nil {
			return "", err
		}

		// Check if the generated code contains profanity
		if !g.containsProfanity(code) {
			return code, nil
		}
		// If profane, retry generation
	}

	// If we exhausted all retries, return an error
	return "", errors.New("failed to generate non-profane short code after maximum retries")
}

// generateRandom generates a random short code without profanity checking.
func (g *ShortCodeGenerator) generateRandom() (string, error) {
	result := make([]byte, g.length)
	charsetLen := big.NewInt(int64(len(g.charset)))

	for i := 0; i < g.length; i++ {
		// Generate a random index using crypto/rand
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = g.charset[randomIndex.Int64()]
	}

	return string(result), nil
}

// containsProfanity checks if the given code contains any profane terms.
// It performs case-insensitive substring matching against the profanity list.
func (g *ShortCodeGenerator) containsProfanity(code string) bool {
	codeLower := strings.ToLower(code)

	for _, profaneWord := range g.profanityList {
		if strings.Contains(codeLower, profaneWord) {
			return true
		}
	}

	return false
}

// GetLength returns the configured length of generated short codes.
func (g *ShortCodeGenerator) GetLength() int {
	return g.length
}

// GetCharset returns the charset used for generating short codes.
func (g *ShortCodeGenerator) GetCharset() string {
	return g.charset
}
