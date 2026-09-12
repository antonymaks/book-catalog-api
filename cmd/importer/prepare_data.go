package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

const (
	inputFile  = "data/raw/booksummaries.txt"
	outputFile = "data/prepared/books.csv"
	bookLimit  = 100
)

func main() {
	input, err := os.Open(inputFile)
	if err != nil {
		panic(fmt.Errorf("failed to open input file: %w", err))
	}
	defer input.Close()

	reader := csv.NewReader(input)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	output, err := os.Create(outputFile)
	if err != nil {
		panic(fmt.Errorf("failed to create output file: %w", err))
	}
	defer output.Close()

	writer := csv.NewWriter(output)
	defer writer.Flush()

	if err := writer.Write([]string{
		"title",
		"author",
		"description",
	}); err != nil {
		panic(err)
	}

	count := 0
	seen := make(map[string]bool)

	for count < bookLimit {
		record, err := reader.Read()
		if err != nil {
			break
		}

		if len(record) < 7 {
			continue
		}

		title := strings.TrimSpace(record[2])
		author := strings.TrimSpace(record[3])
		description := strings.TrimSpace(record[6])
		description = firstSentence(description)

		if title == "" || author == "" || description == "" {
			continue
		}

		key := strings.ToLower(title + "|" + author)

		if seen[key] {
			continue
		}

		seen[key] = true

		err = writer.Write([]string{
			title,
			author,
			description,
		})
		if err != nil {
			panic(err)
		}

		count++
	}

	if err := writer.Error(); err != nil {
		panic(err)
	}

	fmt.Printf("Prepared %d books\n", count)
	fmt.Printf("Saved to %s\n", outputFile)
}

func firstSentence(text string) string {
	text = strings.TrimSpace(text)

	for i, r := range text {
		if r == '.' || r == '!' || r == '?' {
			return strings.TrimSpace(text[:i+len(string(r))])
		}
	}

	return text
}
