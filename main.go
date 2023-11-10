package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/PullRequestInc/go-gpt3"
	config "github.com/anushka2911/ChatGPT_Ctx_Validation/config"
	utils "github.com/anushka2911/ChatGPT_Ctx_Validation/utils"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv(config.ChatGPT_API_KEY)
	if apiKey == "" {
		log.Fatal("ChatGPT API key  missing")
	}

	ctx := context.Background()
	client := gpt3.NewClient(apiKey)
	//create an inputMsg := prompt + input_file_content
	//For creating inputMsg we need to read the input folder and append the content of each file(only with .go extension) to inputMsg

	filePaths, err := utils.GetGoFiles(config.Input)
	if err != nil {
		log.Fatal("Error reading input code file:", err)
	}

	// Create input messages for each file
	var inputMsgs []string
	for _, filePath := range filePaths {
		fileContent, err := utils.ReadFiles(filePath)
		if err != nil {
			log.Printf("Error reading file %s: %v\n", filePath, err)
			continue
		}

		// Create an input message for each file
		inputMsg := fmt.Sprintf("Given this code in file %s, check if context is passed as a parameter. If it's missing, add context as a parameter and provide the correct code:\n\n%s", filePath, fileContent)
		inputMsgs = append(inputMsgs, inputMsg)
	}

	// Process each input message
	for _, inputMsg := range inputMsgs {
		tokenCount, err := utils.GetTokenCount(inputMsg, client)
		if err != nil {
			log.Fatal("Error counting tokens:", err)
		}

		if tokenCount > config.MaxTokensLimit {
			fmt.Println("Input exceeds the maximum token limit.")
			// Since the input exceeds the maximum token limit, we need to split the input into multiple parts and send it to the API

			parts := utils.SplitCode(inputMsg, config.MaxTokensLimit, client) // returns []strings
			err := utils.MakeAPICall(ctx, client, parts, config.Output, config.MaxTokensLimit)
			if err != nil {
				log.Fatal("Error making API call:", err)
			}
		} else {
			// Process the input without splitting
			err := utils.MakeAPICall(ctx, client, []string{inputMsg}, config.Output, config.MaxTokensLimit)
			if err != nil {
				log.Fatal("Error making API call:", err)
			}
		}
	}

}
