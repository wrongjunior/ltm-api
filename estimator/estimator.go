package estimator

import (
	"bufio"
	"errors"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

var (
	russianVowels    = "аеёиоуыэюя"
	englishVowels    = "aeiouy"
	wordRegex        = regexp.MustCompile(`[\p{L}\p{N}]+(-[\p{L}\p{N}]+)*`)
	sentenceEndRegex = regexp.MustCompile(`[.!?]+`)
)

// содержит результаты анализа текста
type Result struct {
	ReadingTime        float64
	WordCount          int
	SentenceCount      int
	SyllableCount      int
	FleschKincaidIndex float64
}

func isRussianWord(word string) bool {
	for _, r := range word {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}

func isYotatedVowel(r rune) bool {
	return r == 'е' || r == 'ё' || r == 'ю' || r == 'я'
}

func isVowel(r rune) bool {
	return strings.ContainsRune(russianVowels+englishVowels, r)
}

func CountSyllables(word string) int {
	word = strings.ToLower(word)
	syllables := 0
	isRussian := isRussianWord(word)

	vowels := englishVowels
	if isRussian {
		vowels = russianVowels
	}

	runes := []rune(word)
	lastWasVowel := false
	lastChar := rune(0)

	for i, char := range runes {
		charIsVowel := strings.ContainsRune(vowels, char)

		if charIsVowel {
			// Считаем слог, если это первая гласная или предыдущий символ не был гласной
			if !lastWasVowel || (isRussian && isYotatedVowel(char) && !isVowel(lastChar)) {
				syllables++
			}
			lastWasVowel = true
		} else {
			// Обработка специфических случаев для русского языка
			if isRussian {
				// Буква 'й' может образовывать слог с предыдущей гласной
				if char == 'й' && i > 0 && isVowel(runes[i-1]) {
					syllables++
				}
			}
			lastWasVowel = false
		}
		lastChar = char
	}

	// Корректировка для английских слов
	if !isRussian {
		if strings.HasSuffix(word, "le") && len(word) > 2 && !strings.ContainsRune(vowels, rune(word[len(word)-3])) {
			syllables++
		}
		if strings.HasSuffix(word, "es") || strings.HasSuffix(word, "ed") {
			// Уменьшаем количество слогов, только если это не приводит к нулю
			if syllables > 1 {
				syllables--
			}
		}
	}

	return max(syllables, 1)
}

// CountWords подсчитывает количество слов в тексте
func CountWords(text string) (int, []string) {
	words := wordRegex.FindAllString(text, -1)
	return len(words), words
}

// CountSentences подсчитывает количество предложений в тексте
func CountSentences(text string) int {
	sentences := sentenceEndRegex.Split(strings.TrimSpace(text), -1)
	count := 0
	for _, s := range sentences {
		if strings.TrimSpace(s) != "" {
			count++
		}
	}
	return count
}

// FleschKincaidIndex рассчитывает индекс Флеша-Кинкейда
func FleschKincaidIndex(wordsCount, sentencesCount, syllablesCount float64) float64 {
	if wordsCount == 0 || sentencesCount == 0 {
		return 0
	}
	// Для очень коротких текстов делаем минимальную коррекцию, чтобы избежать слишком больших значений
	if wordsCount < 3 || sentencesCount < 2 {
		return 100
	}
	return 206.835 - 1.015*(wordsCount/sentencesCount) - 84.6*(syllablesCount/wordsCount)
}

// EstimateReadingTimeParallel оценивает время чтения текста с использованием параллельной обработки
func EstimateReadingTimeParallel(text string, readingSpeed float64, hasVisuals bool, workerCount int) (Result, error) {
	wordsCount, words := CountWords(text)
	sentencesCount := CountSentences(text)

	if wordsCount == 0 || sentencesCount == 0 {
		return Result{}, errors.New("text is empty or invalid")
	}

	// Параллельный подсчет слогов
	syllablesChan := make(chan int)
	var wg sync.WaitGroup

	chunkSize := (len(words) + workerCount - 1) / workerCount
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			end := min(start+chunkSize, len(words))
			localSum := 0
			for j := start; j < end; j++ {
				localSum += CountSyllables(words[j])
			}
			syllablesChan <- localSum
		}(i * chunkSize)
	}

	go func() {
		wg.Wait()
		close(syllablesChan)
	}()

	syllablesCount := 0
	for count := range syllablesChan {
		syllablesCount += count
	}

	fkIndex := FleschKincaidIndex(float64(wordsCount), float64(sentencesCount), float64(syllablesCount))

	adjustedSpeed := readingSpeed
	if fkIndex < 60 {
		adjustedSpeed *= 0.8 // Сложный текст
	}

	readingTime := float64(wordsCount) / adjustedSpeed

	if hasVisuals {
		readingTime *= 1.1
	}

	return Result{
		ReadingTime:        math.Round(readingTime*100) / 100,
		WordCount:          wordsCount,
		SentenceCount:      sentencesCount,
		SyllableCount:      syllablesCount,
		FleschKincaidIndex: fkIndex,
	}, nil
}

// ReadTextFromFile читает текст из файла
func ReadTextFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var text strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text.WriteString(scanner.Text())
		text.WriteString(" ")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return text.String(), nil
}
