package utils

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	tokenizer "github.com/tiktoken-go/tokenizer"
)

func GetGoFiles(output string) ([]string, error) {
	var filePaths []string
	err := filepath.WalkDir(output, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("Error while walking the directory:", err)
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") {
			filePaths = append(filePaths, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}

func GetTokenCount(text string, client gpt3.Client) (int, error) {
	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		return 0, err
	}

	_, tokens, err := enc.Encode(text)
	if err != nil {
		return 0, err
	}

	return len(tokens), nil
}

func ReadFiles(filePaths []string) (map[string]string, error) {
	content := make(map[string]string)

	for _, filePath := range filePaths {
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			return nil, err
		}

		content[filePath] = string(fileContent)
	}

	return content, nil
}

func SplitCode(input string, maxTokens int, client gpt3.Client) []string {
	var parts []string
	lines := strings.Split(input, "\n")
	currentPart := ""
	currentTokenCount := 0

	for _, line := range lines {
		lineTokens, err := GetTokenCount(line, client)
		if err != nil {
			// Handle the error as needed
			continue
		}

		// Check token count before adding a new line
		if currentTokenCount+lineTokens > maxTokens {
			parts = append(parts, currentPart)
			currentPart = ""
			currentTokenCount = 0
		}

		// Add the line to the current part
		currentPart += line + "\n"
		currentTokenCount += lineTokens
	}

	parts = append(parts, currentPart)

	return parts
}

func MakeAPICall(ctx context.Context, client gpt3.Client, inputMsg []string, MaxTokenLimit int) (code string, err error) {
	err = client.CompletionStreamWithEngine(ctx, gpt3.TextDavinci003Engine, gpt3.CompletionRequest{
		Prompt:      inputMsg,
		Temperature: gpt3.Float32Ptr(0),
		MaxTokens:   gpt3.IntPtr(3000),
		N:           gpt3.IntPtr(1),
		Echo:        false,
	}, func(resp *gpt3.CompletionResponse) {
		for _, choice := range resp.Choices {
			code += choice.Text
		}
	})

	if err != nil {
		return "", err
	}

	return code, nil
}
