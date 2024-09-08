package estimator

import (
	"bufio"
	"errors"
	"log"
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
	sentenceEndRegex = regexp.MustCompile(`[.!?]+\s*`)
)

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
			if !lastWasVowel || (isRussian && isYotatedVowel(char) && !isVowel(lastChar)) {
				syllables++
			}
			lastWasVowel = true
		} else {
			if isRussian && char == 'й' && i > 0 && isVowel(runes[i-1]) {
				syllables++
			}
			lastWasVowel = false
		}
		lastChar = char
	}

	if !isRussian {
		if strings.HasSuffix(word, "le") && len(word) > 2 && !strings.ContainsRune(vowels, rune(word[len(word)-3])) {
			syllables++
		}
		if strings.HasSuffix(word, "es") || strings.HasSuffix(word, "ed") {
			if syllables > 1 {
				syllables--
			}
		}
	}

	return max(syllables, 1)
}

func CountWords(text string) (int, []string) {
	words := wordRegex.FindAllString(text, -1)
	return len(words), words
}

func CountSentences(text string) int {
	// Убедимся, что текст не пуст
	if strings.TrimSpace(text) == "" {
		return 0
	}
	sentences := sentenceEndRegex.Split(strings.TrimSpace(text), -1)
	count := 0
	for _, s := range sentences {
		if strings.TrimSpace(s) != "" {
			count++
		}
	}
	return count
}

func FleschKincaidIndex(wordsCount, sentencesCount, syllablesCount float64) float64 {
	if wordsCount == 0 || sentencesCount == 0 {
		return 0
	}
	if wordsCount < 3 || sentencesCount < 2 {
		return 100
	}
	return 206.835 - 1.015*(wordsCount/sentencesCount) - 84.6*(syllablesCount/wordsCount)
}

func EstimateReadingTimeParallel(text string, readingSpeed float64, hasVisuals bool, workerCount int) (Result, error) {
	wordsCount, words := CountWords(text)
	sentencesCount := CountSentences(text)

	if wordsCount == 0 || sentencesCount == 0 {
		return Result{}, errors.New("text is empty or invalid")
	}

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
		adjustedSpeed *= 0.8
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

func StreamProcessFile(filePath string, readingSpeed float64, hasVisuals bool, workerCount int) (Result, error) {
	log.Println("Opening file:", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return Result{}, err
	}
	defer file.Close()

	var (
		totalWords     int
		totalSentences int
		totalSyllables int
		syllablesChan  = make(chan int)
		linesChan      = make(chan string)
		wg             sync.WaitGroup
	)

	scanner := bufio.NewScanner(file)
	log.Println("Starting file scan")

	// Параллельная обработка слогов
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d started", workerID)
			for line := range linesChan {
				log.Printf("Worker %d processing line: %s", workerID, line)
				line = strings.TrimSpace(line)
				if line == "" {
					log.Printf("Worker %d found an empty line, skipping", workerID)
					continue
				}
				words := wordRegex.FindAllString(line, -1)
				localSyllables := 0
				for _, word := range words {
					localSyllables += CountSyllables(word)
				}
				log.Printf("Worker %d counted %d syllables", workerID, localSyllables)
				syllablesChan <- localSyllables
			}
			log.Printf("Worker %d finished", workerID)
		}(i)
	}

	// Подсчет слов и предложений
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		log.Printf("Read line: %s", line)
		if line == "" {
			log.Println("Skipping empty line")
			continue
		}

		wordsInLine := wordRegex.FindAllString(line, -1)
		totalWords += len(wordsInLine)
		log.Printf("Total words after processing line: %d", totalWords)

		sentences := sentenceEndRegex.Split(line, -1)
		for _, s := range sentences {
			if strings.TrimSpace(s) != "" {
				totalSentences++
			}
		}
		log.Printf("Total sentences after processing line: %d", totalSentences)

		linesChan <- line
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error scanning file:", err)
		return Result{}, err
	}

	close(linesChan)
	log.Println("Finished scanning file, waiting for workers")

	go func() {
		wg.Wait()
		close(syllablesChan)
	}()

	for syllables := range syllablesChan {
		totalSyllables += syllables
		log.Printf("Total syllables after worker results: %d", totalSyllables)
	}

	fkIndex := FleschKincaidIndex(float64(totalWords), float64(totalSentences), float64(totalSyllables))
	log.Printf("Flesch-Kincaid index: %f", fkIndex)

	adjustedSpeed := readingSpeed
	if fkIndex < 60 {
		adjustedSpeed *= 0.8
		log.Println("Adjusting reading speed for complex text")
	}

	readingTime := float64(totalWords) / adjustedSpeed
	if hasVisuals {
		readingTime *= 1.1
		log.Println("Adjusting reading time for visuals")
	}

	result := Result{
		ReadingTime:        math.Round(readingTime*100) / 100,
		WordCount:          totalWords,
		SentenceCount:      totalSentences,
		SyllableCount:      totalSyllables,
		FleschKincaidIndex: fkIndex,
	}
	log.Printf("Final result: %+v", result)
	return result, nil
}
