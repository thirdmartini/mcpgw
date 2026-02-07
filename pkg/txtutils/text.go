package txtutils

import (
	"strings"
)

func WrapText(text string, maxLen int) []string {
	if maxLen < 1 {
		maxLen = 80
	}

	var lines []string
	var current strings.Builder

	words := strings.Fields(text) // splits on any whitespace, collapses multiples

	for i, word := range words {
		// Length if we add this word to current line
		addLen := len(word)
		if current.Len() > 0 {
			addLen += 1 // space
		}

		if current.Len()+addLen <= maxLen {
			// Fits → add it
			if current.Len() > 0 {
				current.WriteByte(' ')
			}
			current.WriteString(word)
		} else {
			// Doesn't fit
			if current.Len() > 0 {
				// Flush current line
				lines = append(lines, current.String())
				current.Reset()
			}

			// Now handle the current word
			if len(word) > maxLen {
				// Word itself is too long → have to break it (unfortunate but correct)
				lines = append(lines, word[:maxLen])
				remainder := word[maxLen:]
				// Put remainder back to be processed as next "word"
				words = append([]string{remainder}, words[i+1:]...)
				i-- // re-process this index
				continue
			}

			// Normal case: start new line with this word
			current.WriteString(word)
		}
	}

	// Don't forget the last line
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}

func WrapTextWithNewLines(text string, maxLen int) []string {
	if maxLen <= 0 {
		maxLen = 80
	}

	var lines []string
	var current strings.Builder

	paragraphs := strings.Split(text, "\n")

	for pIdx, paragraph := range paragraphs {
		if pIdx > 0 {
			// Previous paragraph ended → flush any pending line
			if current.Len() > 0 {
				lines = append(lines, current.String())
				current.Reset()
			}
			// Empty line between paragraphs (if original had blank lines)
			if paragraph == "" {
				lines = append(lines, "")
				continue
			}
		}

		words := strings.Fields(paragraph)
		for _, word := range words {
			wordLen := len(word)

			// If word itself is longer than maxLen → must break it
			if wordLen > maxLen {
				if current.Len() > 0 {
					lines = append(lines, current.String())
					current.Reset()
				}
				// Break very long word into chunks
				for wordLen > maxLen {
					lines = append(lines, word[:maxLen])
					word = word[maxLen:]
					wordLen = len(word)
				}
				if wordLen > 0 {
					current.WriteString(word)
				}
				continue
			}

			// Space needed before this word?
			spaceLen := 0
			if current.Len() > 0 {
				spaceLen = 1
			}

			// Does it fit on current line?
			if current.Len()+spaceLen+wordLen <= maxLen {
				if spaceLen == 1 {
					current.WriteByte(' ')
				}
				current.WriteString(word)
			} else {
				// Doesn't fit → commit current line
				if current.Len() > 0 {
					lines = append(lines, current.String())
					current.Reset()
				}
				// Start new line with this word
				current.WriteString(word)
			}
		}

		// After processing all words in paragraph, commit last line (unless empty)
		if current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
	}

	// In case text didn't end with \n but had trailing content
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}
