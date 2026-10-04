package importer

import (
	"regexp"
	"strings"
	"unicode"
)

// accentFolder maps the Latin letters that carry diacritics to their base
// letter, so "Beyoncé" and "Beyonce" compare equal.
var accentFolder = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ã", "a", "ä", "a", "å", "a", "ā", "a", "ă", "a", "ą", "a",
	"ç", "c", "ć", "c", "č", "c", "ď", "d", "đ", "d",
	"è", "e", "é", "e", "ê", "e", "ë", "e", "ē", "e", "ė", "e", "ę", "e", "ě", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i", "ī", "i", "ı", "i",
	"ł", "l", "ñ", "n", "ń", "n", "ň", "n",
	"ò", "o", "ó", "o", "ô", "o", "õ", "o", "ö", "o", "ø", "o", "ō", "o", "ő", "o", "œ", "oe",
	"ř", "r", "ś", "s", "š", "s", "ş", "s", "ß", "ss", "ť", "t",
	"ù", "u", "ú", "u", "û", "u", "ü", "u", "ū", "u", "ů", "u", "ű", "u",
	"ý", "y", "ÿ", "y", "ź", "z", "ż", "z", "ž", "z", "æ", "ae",
)

var (
	bracketedPattern = regexp.MustCompile(`[\(\[\{][^\)\]\}]*[\)\]\}]`)
	dashSuffix       = regexp.MustCompile(`\s[-–—]\s.*$`)
	featuringPattern = regexp.MustCompile(`\s(?:feat\.?|ft\.?|featuring)\s.*$`)
)

// variantWords are version markers that change which recording a title
// means; a mismatch between wanted and found lowers the score.
var variantWords = []string{
	"live", "remix", "acoustic", "instrumental", "karaoke", "cover", "demo",
	"sped up", "slowed", "reverb", "orchestral", "unplugged", "a cappella", "mashup", "8d",
}

// noiseWords mark release details that do not change the recording.
var noiseWords = []string{
	"remaster", "remastered", "mono", "stereo", "radio edit", "single version", "album version",
	"deluxe", "expanded", "explicit", "bonus track", "anniversary", "version", "edition", "edit",
}

// foldText lowercases, folds accents and reduces punctuation to single spaces.
func foldText(text string) string {
	folded := accentFolder.Replace(strings.ToLower(text))
	var builder strings.Builder
	pendingSpace := false
	for _, character := range folded {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			if pendingSpace && builder.Len() > 0 {
				builder.WriteByte(' ')
			}
			pendingSpace = false
			builder.WriteRune(character)
			continue
		}
		pendingSpace = true
	}
	return builder.String()
}

// titleParts is a title split into its comparable core and its version markers.
type titleParts struct {
	core     string
	variants []string
}

// splitTitle removes featuring credits and release noise from a title and
// collects the version markers (live, remix, ...) found in it.
func splitTitle(title string) titleParts {
	lowered := strings.ToLower(title)
	annotations := bracketedPattern.FindAllString(lowered, -1)
	suffix := dashSuffix.FindString(lowered)
	remainder := bracketedPattern.ReplaceAllString(lowered, " ")
	remainder = dashSuffix.ReplaceAllString(remainder, "")
	remainder = featuringPattern.ReplaceAllString(" "+remainder, "")

	annotationText := foldText(strings.Join(annotations, " ") + " " + suffix)
	variants := findVariants(annotationText + " " + foldText(remainder))
	return titleParts{core: foldText(remainder), variants: variants}
}

// findVariants returns the variant markers present in already folded text.
func findVariants(folded string) []string {
	padded := " " + folded + " "
	found := []string{}
	for _, word := range variantWords {
		if strings.Contains(padded, " "+word+" ") {
			found = append(found, word)
		}
	}
	return found
}

// stripNoise drops release-detail words from an album or show title.
func stripNoise(title string) string {
	parts := splitTitle(title)
	words := strings.Fields(parts.core)
	kept := words[:0]
	for _, word := range words {
		if !isNoiseWord(word) {
			kept = append(kept, word)
		}
	}
	return strings.Join(kept, " ")
}

func isNoiseWord(word string) bool {
	for _, noise := range noiseWords {
		if word == noise {
			return true
		}
	}
	return false
}

// similarity is 1 for equal strings and falls with the edit distance.
func similarity(left, right string) float64 {
	if left == right {
		return 1
	}
	longest := max(len([]rune(left)), len([]rune(right)))
	if longest == 0 {
		return 0
	}
	return 1 - float64(editDistance(left, right))/float64(longest)
}

func editDistance(left, right string) int {
	leftRunes, rightRunes := []rune(left), []rune(right)
	previous := make([]int, len(rightRunes)+1)
	for column := range previous {
		previous[column] = column
	}
	for row := 1; row <= len(leftRunes); row++ {
		current := make([]int, len(rightRunes)+1)
		current[0] = row
		for column := 1; column <= len(rightRunes); column++ {
			cost := 1
			if leftRunes[row-1] == rightRunes[column-1] {
				cost = 0
			}
			current[column] = min(previous[column]+1, current[column-1]+1, previous[column-1]+cost)
		}
		previous = current
	}
	return previous[len(rightRunes)]
}
