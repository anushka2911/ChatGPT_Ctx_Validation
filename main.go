package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/PullRequestInc/go-gpt3"
	config "github.com/anushka2911/ChatGPT_Ctx_Validation/config"
	utils "github.com/anushka2911/ChatGPT_Ctx_Validation/utils"
	"github.com/joho/godotenv"
	copy "github.com/otiai10/copy"
)

const (
	MaxExecutionTime = 1000 * time.Second
	Prompt           = "For given code, check if context is passed as a parameter. If missing, add context as a parameter and provide the correct code:\n\n"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
		return
	}

	apiKey := os.Getenv(config.ChatGPT_API_KEY)
	if apiKey == "" {
		log.Println("ChatGPT API key missing")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), MaxExecutionTime)
	defer cancel()

	client := gpt3.NewClient(apiKey)

	if err := copyFiles(); err != nil {
		log.Println("Error copying input folder:", err)
		return
	}

	filePaths, err := utils.GetGoFiles(config.Output)
	if err != nil {
		log.Println("Error reading input code file:", err)
		return
	}
	//path : code
	fileContentMap, err := utils.ReadFiles(filePaths)
	if err != nil {
		log.Println("Error reading input code file:", err)
		return
	}

	for filePath, code := range fileContentMap {
		if err := processFile(ctx, client, filePath, code); err != nil {
			log.Println(err)
		}
	}
}

func copyFiles() error {
	return copy.Copy(config.Input, config.Output)
}

func processFile(ctx context.Context, client gpt3.Client, filePath, code string) error {
	tokenCount, err := utils.GetTokenCount(Prompt+code, client)
	if err != nil {
		return fmt.Errorf("error counting tokens for file %s: %v", filePath, err)
	}

	if tokenCount < config.MaxTokensLimit {
		fmt.Printf("TOKEN COUNT FOR FILE %s: %d\n", filePath, tokenCount)
		inputMsg := Prompt + code

		resp, err := utils.MakeAPICall(ctx, client, []string{inputMsg}, config.MaxTokensLimit)
		if err != nil {

			log.Fatal(err)
		}

		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("error opening file %s: %v", filePath, err)
		}
		defer file.Close()

		_, err = file.WriteString(resp)
		if err != nil {
			return fmt.Errorf("error writing corrected code to file %s: %v", filePath, err)
		}

	} else {
		fmt.Printf("MAXIMA Token Count for input file(%d\n) >= MaxTokensLimit: ", tokenCount)
	}
	return nil
}
