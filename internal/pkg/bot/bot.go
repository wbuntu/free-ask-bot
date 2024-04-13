package bot

import (
	"context"
	"fmt"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/config"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/llm"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/log"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

func Setup(ctx context.Context, c *config.Config) error {
	bot, err := tgbotapi.NewBotAPI(c.Bot.Token)
	if err != nil {
		return errors.Wrap(err, "setup bot api")
	}
	bot.Debug = false
	sharedInstance = &Bot{
		bot:    bot,
		logger: log.WithField("module", "bot"),
	}
	sharedInstance.logger.Infof("Authorized on account %s", bot.Self.UserName)
	sharedInstance.ctx, sharedInstance.cancel = context.WithCancel(ctx)
	// 同步 bot 命令
	if err := sharedInstance.SyncCommands(); err != nil {
		return errors.Wrap(err, "sync commands")
	}
	return nil
}

func Serve() error {
	if sharedInstance != nil {
		return sharedInstance.Serve()
	}
	return nil
}

func Shutdown() error {
	if sharedInstance != nil {
		sharedInstance.Shutdown()
	}
	return nil
}

var sharedInstance *Bot

type Bot struct {
	ctx    context.Context
	cancel context.CancelFunc
	bot    *tgbotapi.BotAPI
	logger log.Logger
}

func (b *Bot) SyncCommands() error {
	commands := generateBotCommands()
	msg := tgbotapi.NewSetMyCommands(commands...)
	if _, err := b.bot.Request(msg); err != nil {
		return errors.Wrap(err, "send message")
	}
	b.logger.Infof("Sync bot commands success")
	return nil
}

func (b *Bot) Serve() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.bot.GetUpdatesChan(u)
	for {
		select {
		case <-b.ctx.Done():
			return nil
		case update := <-updates:
			if err := b.handleUpdate(&update); err != nil {
				b.logger.WithField("updateID", update.UpdateID).Errorf("handle update: %s", err)
			}
		}
	}
}

func (b *Bot) Shutdown() error {
	b.bot.StopReceivingUpdates()
	b.cancel()
	return nil
}

func (b *Bot) handleUpdate(update *tgbotapi.Update) error {
	if update.Message != nil {
		// TODO: 根据聊天 ID 入队列，并发处理请求
		input := update.Message
		logger := b.logger.WithFields(log.Fields{"chatID": input.Chat.ID, "messageID": input.MessageID, "userID": input.From.ID, "languageCode": input.From.LanguageCode})
		logger.Infof("message received")
		// logger.Debugf("inputText: %s", input.Text)
		output := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		chat, err := llm.GetChat(input.Chat.ID, input.From.ID, input.From.LanguageCode)
		if err != nil {
			return errors.Wrap(err, "get chat")
		}
		record := &llm.Record{}
		if input.IsCommand() {
			switch input.Command() {
			case BotCommandSearch.Command:
				if len(input.CommandArguments()) == 0 {
					output.Text = generateEmptySearchText(chat.LanguageCode)
				} else {
					// 搜索内容与搜索结果需要保存
					completion, err := chat.Search(b.ctx, input.CommandArguments())
					if err != nil {
						output.Text = fmt.Sprintf("Failed to fetch search message: %s", err)
					} else {
						output.Text = completion.Summarize
						if len(completion.URL) > 0 {
							output.Text += "参考链接:\n"
							for i := range completion.URL {
								output.Text += fmt.Sprintf("%d. %s\n", i+1, completion.URL[i])
							}
						}
						record.Save = true
						record.User = input.CommandArguments()
						record.Assistant = completion.Summarize
						// logger.Debugf("summarize: %s", completion.Summarize)
					}
				}
			case BotCommandLanguage.Command:
				// 无参数时提示设置语言，有参数时检查参数是否有效，若有效则切换语言
				if len(input.CommandArguments()) == 0 {
					output.Text = generateChooseLangeCodeText(chat.LanguageCode)
				} else {
					languageCode := utils.LanguageCode(input.CommandArguments())
					text, exist := generateSetLangeCodeText(languageCode)
					if exist {
						chat.LanguageCode = languageCode
					}
					output.Text = text
				}
			case BotCommandStart.Command:
				// 清除用户历史消息，重新开始对话
				chat.Reset()
				output.Text = generateResetText(chat)
			case BotCommandHelp.Command:
				// 打印帮助信息
				output.Text = generateHelpText(chat)
			case BotCommandSettings.Command:
				// 打印当前对话设置，使用 markdown 格式
				output.Text = generateSettingsText(chat)
				output.ParseMode = "MarkdownV2"
			default:
				output.Text = "Unsupported command"
			}
		} else {
			// 聊天中的提问与响应需要保存
			completion, err := chat.Chat(b.ctx, input.Text)
			if err != nil {
				output.Text = fmt.Sprintf("Failed to fetch chat message: %s", err)
			} else {
				output.Text = completion.Content
				record.Save = true
				record.User = input.Text
				record.Assistant = completion.Content
			}
		}
		// logger.Debugf("outputText: %s:outputEntities: %v", output.Text, output.Entities)
		outputMessage, err := b.bot.Send(output)
		if err != nil {
			return errors.Wrap(err, "send message")
		}
		logger.WithField("messageID", outputMessage.MessageID).Infof("message sent")
		// 记录消息上下文
		if record.Save {
			chat.AddHistory(record)
		}
	} else {
		b.logger.WithField("updateID", update.UpdateID).Errorf("message empty: print update: %v", update)
	}
	return nil
}
