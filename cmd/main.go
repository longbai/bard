package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/longbai/bard"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println(`
ChatGPT - A command-line interface to Google's Bard (https://bard.google.com/)
Repo: github.com/longbai/Bard

Enter 'exit' or 'reset' to send a command.
`)
	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)
	var sessionID string

	fmt.Print("Enter your session ID: ")
	fmt.Scanln(&sessionID)

	proxy := os.Getenv("HTTP_PROXY")
	chatbot := Bard.NewChatbot(sessionID, proxy)

	for {
		fmt.Print("You: ")
		userPrompt, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}

		userPrompt = strings.TrimSpace(userPrompt)

		if userPrompt == "exit" {
			break
		} else if userPrompt == "reset" {
			chatbot.ConversationID = ""
			chatbot.ResponseID = ""
			chatbot.ChoiceID = ""
			continue
		}
		fmt.Println("Google Bard:")

		response, err := chatbot.Ask(userPrompt)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Println(response.Content)
	}
}
