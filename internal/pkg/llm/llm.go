package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/utils"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

var chatCache = cache.New(cache.NoExpiration, 365*24*time.Hour)

const apiKey = "EMPTY"

const chatAPIURL = "http://localhost:3040/v1"

const searchAPIURL = "http://localhost:8000/v1"

func Setup(ctx context.Context, c *config.Config) error {
	chatConfig := openai.DefaultConfig(apiKey)
	chatConfig.BaseURL = chatAPIURL
	searchConfig := openai.DefaultConfig(apiKey)
	searchConfig.BaseURL = searchAPIURL
	sharedInstance = &LLMClient{
		m:            openai.GPT3Dot5Turbo,
		chatClient:   openai.NewClientWithConfig(chatConfig),
		searchClient: openai.NewClientWithConfig(searchConfig),
	}
	return nil
}

func Serve() error {
	return nil
}

func Shutdown() error {
	return nil
}

var sharedInstance *LLMClient

type LLMClient struct {
	m            string
	chatClient   *openai.Client
	searchClient *openai.Client
}

type Chat struct {
	ID           int64
	UserID       int64
	LanguageCode utils.LanguageCode
	system       []openai.ChatCompletionMessage
	history      []openai.ChatCompletionMessage
}

type SearchCompletion struct {
	Content   string
	Summarize string
	URL       []string
}

type ChatCompletion struct {
	Content string
}

type Record struct {
	Save      bool
	User      string
	Assistant string
}

func GetChat(id int64, userID int64, languageCode string) (*Chat, error) {
	key := strconv.FormatInt(id, 10)
	if v, ok := chatCache.Get(key); ok {
		return v.(*Chat), nil
	}
	v := &Chat{
		ID:           id,
		UserID:       userID,
		LanguageCode: utils.LanguageCodeZhHans,
		system:       []openai.ChatCompletionMessage{},
		history:      []openai.ChatCompletionMessage{},
	}
	chatCache.SetDefault(key, v)
	return v, nil
}

func (c *Chat) Reset() {
	c.history = []openai.ChatCompletionMessage{}
}

// 提取链接的正则表达式
var linkReg = regexp.MustCompile(`\[.*?\]\((.*?)\)`)

func (c *Chat) Search(ctx context.Context, inputContent string) (*SearchCompletion, error) {
	// 执行搜索时忽略上下文
	request := openai.ChatCompletionRequest{Model: sharedInstance.m}
	request.Messages = []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: inputContent}}
	request.Stream = true
	stream, err := sharedInstance.searchClient.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "create chat completion")
	}
	defer stream.Close()
	completion := &SearchCompletion{}
	for {
		response, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return nil, errors.Wrap(err, "receive chat stream")
			}
		}
		completion.Content += response.Choices[0].Delta.Content
	}
	index := strings.Index(completion.Content, "---")
	if index > 0 {
		completion.Summarize = completion.Content[:index]
		// 提取链接
		links := linkReg.FindAllStringSubmatch(completion.Content[index:], -1)
		// 保存链接
		for _, link := range links {
			completion.URL = append(completion.URL, link[1])
		}
	} else {
		completion.Summarize = completion.Content
	}
	return completion, nil
}

func (c *Chat) Chat(ctx context.Context, content string) (*ChatCompletion, error) {
	// 执行补全时需要发送上下文
	request := openai.ChatCompletionRequest{
		Model:       sharedInstance.m,
		Messages:    []openai.ChatCompletionMessage{generateChatGPTSystemMessage()},
		Temperature: 0.5,
		TopP:        1,
	}
	request.Messages = append(request.Messages, c.history...)
	request.Messages = append(request.Messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: content})
	response, err := sharedInstance.chatClient.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "create chat completion")
	}
	return &ChatCompletion{Content: response.Choices[0].Message.Content}, nil
}

func (c *Chat) AddHistory(recorcd *Record) error {
	c.history = append(
		c.history,
		openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: recorcd.User},
		openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: recorcd.Assistant},
	)
	// 保存最近 8 个历史消息
	if len(c.history) > 8 {
		c.history = c.history[2:]
	}
	// TODO: 生成历史摘要
	return nil
}

func (c *Chat) FormatHistory() string {
	m := "```\nchat history is empty\n```"
	if len(c.history) == 0 {
		return m
	}
	if data, err := json.MarshalIndent(c.history, "", "    "); err == nil {
		m = "```json\n" + string(data) + "\n```"
	}
	return m
}

func generateChatGPTSystemMessage() openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("\nYou are ChatGPT, a large language model trained by OpenAI.\nKnowledge cutoff: 2021-09\nCurrent model: gpt-3.5-turbo\nCurrent time: %s\nLatex inline: $x^2$ \nLatex block: $$e=mc^2$$\n\n", time.Now().UTC().Format("1/2/2006, 3:04:05 PM")),
	}
}
