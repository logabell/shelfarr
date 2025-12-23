package indexer

import (
	"strings"
)

// QualityScore represents the quality score for a search result
type QualityScore struct {
	Score       int    // Higher is better, -1 means unacceptable
	FormatMatch bool   // True if format is in the profile
	FormatRank  int    // Position in format ranking (0 = best)
	Reason      string // Human-readable explanation
}

// ScoreResult calculates a quality score for a search result based on a quality profile
// formatRanking is a comma-separated list of preferred formats (e.g., "epub,azw3,mobi,pdf")
// minBitrate is the minimum acceptable bitrate for audiobooks (0 means no minimum)
// Returns a QualityScore with higher values being better, -1 for unacceptable
func ScoreResult(result SearchResult, formatRanking string, minBitrate int, isAudiobook bool) QualityScore {
	// Parse format ranking into list
	formats := parseFormatRanking(formatRanking)

	if len(formats) == 0 {
		return QualityScore{
			Score:       0,
			FormatMatch: false,
			Reason:      "No format preferences configured",
		}
	}

	// Normalize the result format
	resultFormat := strings.ToLower(strings.TrimSpace(result.Format))

	// Find position in ranking
	formatRank := -1
	for i, f := range formats {
		if strings.EqualFold(f, resultFormat) {
			formatRank = i
			break
		}
	}

	// If format not in ranking, it's not preferred but might be acceptable
	if formatRank == -1 {
		return QualityScore{
			Score:       0,
			FormatMatch: false,
			FormatRank:  -1,
			Reason:      "Format not in preferred list",
		}
	}

	// Check bitrate for audiobooks
	if isAudiobook && minBitrate > 0 && result.Bitrate > 0 {
		if result.Bitrate < minBitrate {
			return QualityScore{
				Score:       -1,
				FormatMatch: true,
				FormatRank:  formatRank,
				Reason:      "Bitrate below minimum",
			}
		}
	}

	// Calculate score: higher format rank = higher score
	// Base score starts at 100 and decreases by 10 for each rank position
	baseScore := 100 - (formatRank * 10)
	if baseScore < 10 {
		baseScore = 10
	}

	// Bonus points for seeders
	seederBonus := 0
	if result.Seeders >= 10 {
		seederBonus = 20
	} else if result.Seeders >= 5 {
		seederBonus = 10
	} else if result.Seeders >= 1 {
		seederBonus = 5
	}

	// Bonus for freeleech
	freeleechBonus := 0
	if result.Freeleech {
		freeleechBonus = 5
	}

	// Bitrate bonus for audiobooks
	bitrateBonus := 0
	if isAudiobook && result.Bitrate > 0 {
		if result.Bitrate >= 256 {
			bitrateBonus = 15
		} else if result.Bitrate >= 128 {
			bitrateBonus = 10
		} else if result.Bitrate >= 64 {
			bitrateBonus = 5
		}
	}

	totalScore := baseScore + seederBonus + freeleechBonus + bitrateBonus

	return QualityScore{
		Score:       totalScore,
		FormatMatch: true,
		FormatRank:  formatRank,
		Reason:      "Good match",
	}
}

// parseFormatRanking splits a comma-separated format string into a list
func parseFormatRanking(ranking string) []string {
	if ranking == "" {
		return nil
	}

	parts := strings.Split(ranking, ",")
	formats := make([]string, 0, len(parts))

	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			formats = append(formats, p)
		}
	}

	return formats
}

// GetBestResult returns the best result from a list based on quality scoring
// Returns nil if no acceptable results found
func GetBestResult(results []SearchResult, formatRanking string, minBitrate int, isAudiobook bool) *SearchResult {
	var bestResult *SearchResult
	bestScore := -1

	for i := range results {
		score := ScoreResult(results[i], formatRanking, minBitrate, isAudiobook)
		if score.Score > bestScore {
			bestScore = score.Score
			bestResult = &results[i]
		}
	}

	return bestResult
}

// SortResultsByQuality sorts results by quality score (highest first)
func SortResultsByQuality(results []SearchResult, formatRanking string, minBitrate int, isAudiobook bool) []SearchResult {
	// Create scored results
	type scoredResult struct {
		result SearchResult
		score  int
	}

	scored := make([]scoredResult, len(results))
	for i, r := range results {
		s := ScoreResult(r, formatRanking, minBitrate, isAudiobook)
		scored[i] = scoredResult{result: r, score: s.Score}
	}

	// Simple bubble sort (could be optimized for larger lists)
	for i := 0; i < len(scored)-1; i++ {
		for j := 0; j < len(scored)-i-1; j++ {
			if scored[j].score < scored[j+1].score {
				scored[j], scored[j+1] = scored[j+1], scored[j]
			}
		}
	}

	// Extract sorted results
	sorted := make([]SearchResult, len(scored))
	for i, s := range scored {
		sorted[i] = s.result
	}

	return sorted
}
