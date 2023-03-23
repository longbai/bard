package main

import (
    "fmt"
    "math/rand"
    "os"
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
    var sessionID, at string
    fmt.Print("Enter your session ID: ")
    fmt.Scanln(&sessionID)
    fmt.Print("Enter your AT value: ")
    fmt.Scanln(&at)

    proxy := os.Getenv("HTTP_PROXY")

    chatbot := bard.NewChatbot(sessionID, at, proxy)
    for {
        fmt.Print("You: ")
        var userPrompt string
        fmt.Scanln(&userPrompt)

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
