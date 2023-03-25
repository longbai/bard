package bard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Chatbot struct {
	Headers        http.Header
	ReqID          int
	AT             string
	ConversationID string
	ResponseID     string
	ChoiceID       string
	Proxy          *url.URL
}

func NewChatbot(sessionID, proxy string) *Chatbot {
	headers := http.Header{
		"Host":          []string{"bard.google.com"},
		"X-Same-Domain": []string{"1"},
		"User-Agent":    []string{"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"},
		"Content-Type":  []string{"application/x-www-form-urlencoded;charset=UTF-8"},
		"Origin":        []string{"https://bard.google.com"},
		"Referer":       []string{"https://bard.google.com/"},
		"Cookie":        []string{"__Secure-1PSID=" + sessionID + ";"},
	}
	reqID := rand.Intn(10000)
	bot := &Chatbot{
		Headers: headers,
		ReqID:   reqID,
	}
	bot.AT = bot.getAt()

	if proxy != "" {
		bot.Proxy, _ = url.Parse(proxy)
	}
	return bot
}

func (c *Chatbot) getAt() string {
	req, err := http.NewRequest("GET", "https://bard.google.com/", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header = c.Headers
	client := &http.Client{}
	if c.Proxy != nil {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(c.Proxy),
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("从页面中获取 At 失败", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return regexp.MustCompile(`SNlM0e":"(\w+:\d+)"`).FindStringSubmatch(string(body))[1]
}

type Response struct {
	Content           string
	ConversationID    string
	ResponseID        string
	FactualityQueries []interface{}
	TextQuery         string
	Choices           []Choice
}

type Choice struct {
	ID      string
	Content string
}

func readLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func askUrl(params url.Values) string {
	return "https://bard.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate?" + params.Encode()
}

func (c *Chatbot) Ask(message string) (*Response, error) {
	params := url.Values{
		"bl":     []string{"boq_assistant-bard-web-server_20230315.04_p1"},
		"_reqid": []string{fmt.Sprintf("%d", c.ReqID)},
		"rt":     []string{"c"},
	}
	messageStruct := []interface{}{
		[]interface{}{message},
		nil,
		[]interface{}{c.ConversationID, c.ResponseID, c.ChoiceID},
	}
	messageJSON, _ := json.Marshal(messageStruct)
	escapedMessage := strings.Replace(string(messageJSON), "\"", "\\\"", -1)
	data := url.Values{
		"f.req": []string{fmt.Sprintf(`[null, "%s"]`, escapedMessage)},
		"at":    []string{c.AT},
	}

	req, err := http.NewRequest("POST", askUrl(params),
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header = c.Headers
	client := &http.Client{}
	if c.Proxy != nil {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(c.Proxy),
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	lines, err := readLines(resp.Body)
	if err != nil {
		return nil, err
	}

	//log.Println(strings.Join(lines, "\n"))
	chatData := json.RawMessage(lines[3])
	if chatData == nil {
		return &Response{Content: fmt.Sprintf("Google Bard encountered an error: %s.", strings.Join(lines, "\n"))}, nil
	}
	var jsonChatData []interface{}
	err = json.Unmarshal(chatData, &jsonChatData)
	if err != nil {
		return nil, err
	}
	var NewJsonChatData []interface{}
	json.Unmarshal(json.RawMessage(jsonChatData[0].([]interface{})[2].(string)), &NewJsonChatData)

	var choices []Choice
	for _, item := range NewJsonChatData[4].([]interface{}) {
		choices = append(choices, Choice{ID: item.([]interface{})[0].(string), Content: item.([]interface{})[1].([]interface{})[0].(string)})
	}

	results := &Response{
		Content:           NewJsonChatData[0].([]interface{})[0].(string),
		ConversationID:    NewJsonChatData[1].([]interface{})[0].(string),
		ResponseID:        NewJsonChatData[1].([]interface{})[1].(string),
		FactualityQueries: NewJsonChatData[3].([]interface{}),
		TextQuery:         NewJsonChatData[2].([]interface{})[0].([]interface{})[0].(string),
		Choices:           choices,
	}
	c.ConversationID = results.ConversationID
	c.ResponseID = results.ResponseID
	c.ChoiceID = results.Choices[0].ID
	c.ReqID += 100000
	return results, nil
}
