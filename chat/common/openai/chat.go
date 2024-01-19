package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	copenai "github.com/sashabaranov/go-openai"
	"github.com/zeromicro/go-zero/core/logx"
)

// ChatStream 数据流式传输
func (c *ChatClient) ChatStream(req []ChatModelMessage, channel chan string) (string, error) {

	config := c.buildConfig()

	cli := copenai.NewClientWithConfig(config)

	// 打印请求信息
	logx.Info("Chat stream req: ", req)
	first := 0
	var system ChatModelMessage
	for i, msg := range req {
		if msg.Role == "system" {
			system = msg
		}
		if i%2 == 0 {
			continue
		}
		//估算长度
		if NumTokensFromMessages(req[len(req)-i-1:], ChatModel) < (c.TotalToken - c.MaxToken) {
			first = len(req) - i - 1
		} else {
			break
		}
	}

	var messages []copenai.ChatCompletionMessage

	if first != 0 {
		messages = append(messages, copenai.ChatCompletionMessage{
			Role:    system.Role,
			Content: system.Content,
		})
	}

	for _, message := range req[first:] {
		messages = append(messages, copenai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	fmt.Println("before model:", c.Model)
	if _, ok := Models[c.Model]; !ok {
		c.Model = ChatModel
	}
	request := copenai.ChatCompletionRequest{
		Model:       c.Model,
		Messages:    messages,
		MaxTokens:   c.MaxToken,
		Temperature: c.Temperature,
		TopP:        1,
	}
	fmt.Println("current request:", request)
	stream, err := cli.CreateChatCompletionStream(context.Background(), request)

	if err != nil {
		fmt.Println("req chat stream params:", config)
		return "", err
	}
	defer stream.Close()

	messageText := ""
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logx.Info("Stream finished")
			close(channel)
			return messageText, nil
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			close(channel)
			return messageText, err
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			channel <- content
			// 如果是对内容的进行补充
			messageText += content
			// 结算
			if response.Choices[0].FinishReason != "" {
				close(channel)
				return messageText, nil
			}
		}

		logx.Info("Chat Stream response:", response)
	}
}

func (c *ChatClient) Chat(req []ChatModelMessage) (string, error) {

	config := c.buildConfig()
	cli := copenai.NewClientWithConfig(config)

	// 打印请求信息
	logx.Info("Chat req: ", req)

	first := 0
	var system ChatModelMessage
	for i, msg := range req {
		if msg.Role == "system" {
			system = msg
		}
		if i%2 == 0 {
			continue
		}
		//估算长度
		if NumTokensFromMessages(req[len(req)-i-1:], ChatModel) < (c.TotalToken - c.MaxToken) {
			first = len(req) - i - 1
		} else {
			break
		}
	}

	var messages []copenai.ChatCompletionMessage

	if first != 0 {
		messages = append(messages, copenai.ChatCompletionMessage{
			Role:    system.Role,
			Content: system.Content,
		})
	}

	for _, message := range req[first:] {
		messages = append(messages, copenai.ChatCompletionMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}
	if _, ok := Models[c.Model]; !ok {
		c.Model = ChatModel
	}
	request := copenai.ChatCompletionRequest{
		Model:       c.Model,
		Messages:    messages,
		MaxTokens:   c.MaxToken,
		Temperature: c.Temperature,
		TopP:        1,
	}
	chat, err := cli.CreateChatCompletion(context.Background(), request)
	if err != nil {
		fmt.Println("req chat params:", config)
		return "", err
	}
	txt := ""
	for _, choice := range chat.Choices {
		txt += choice.Message.Content
	}

	logx.Info("Chat resp: ", txt)

	return txt, nil
}

func (c *ChatClient) HasGpt4() bool {

	config := c.buildConfig()
	cli := copenai.NewClientWithConfig(config)

	var messages []copenai.ChatCompletionMessage

	request := copenai.ChatCompletionRequest{
		Model:       ChatModel4,
		Messages:    messages,
		MaxTokens:   c.MaxToken,
		Temperature: c.Temperature,
		TopP:        1,
	}
	chat, _ := cli.CreateChatCompletion(context.Background(), request)

	return strings.Contains(chat.Model, "gpt4")
}
