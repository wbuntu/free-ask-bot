package bot

import (
	"fmt"

	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/llm"
	"gitbub.com/wbuntu/free-ask-bot/internal/pkg/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command struct {
	// 1～32, lowercase English letters, digits and underscores.
	Command string
	// 3-256 characters, en|zh-Hans|zh-Hant
	Descriptions []string
}

var (
	BotCommandSearch = &Command{
		Command:      "search",
		Descriptions: []string{"Search internet and summarize results", "搜索互联网并总结查询结果"},
	}
	BotCommandLanguage = &Command{
		Command:      "language",
		Descriptions: []string{"Set language for search internet", "设置用于搜索引擎的语言"},
	}
	BotCommandStart = &Command{
		Command:      "start",
		Descriptions: []string{"Clear context and start new chat", "清除上下文并启动新的对话"},
	}
	BotCommandHelp = &Command{
		Command:      "help",
		Descriptions: []string{"Display help message", "显示当前使用文档"},
	}
	BotCommandSettings = &Command{
		Command:      "settings",
		Descriptions: []string{"Show bot settings", "显示 bot 配置"},
	}
)

var availableCommands = []*Command{BotCommandSearch, BotCommandStart, BotCommandHelp, BotCommandSettings}

func generateBotCommands() []tgbotapi.BotCommand {
	commands := []tgbotapi.BotCommand{}
	for _, v := range availableCommands {
		commands = append(commands, tgbotapi.BotCommand{
			Command:     v.Command,
			Description: v.Descriptions[1],
		})
	}
	return commands
}

func generateEmptySearchText(languageCode utils.LanguageCode) string {
	return "内容为空，请按照: /search content 的格式输入"
}

func generateChooseLangeCodeText(languageCode utils.LanguageCode) string {
	return "Choose your language:\nEnglish(en) 简体中文(zh-hans) 繁體中文(zh-hant)"
}

func generateSetLangeCodeText(languageCode utils.LanguageCode) (string, bool) {
	m := map[utils.LanguageCode]string{
		utils.LanguageCodeEn:     "Change language to English(en) success",
		utils.LanguageCodeZhHans: "设置简体中文(zh-hans)成功",
		utils.LanguageCodeZhHant: "設置繁體中文(zh-hant)成功",
	}
	if v, ok := m[languageCode]; ok {
		return v, true
	}
	return "Unsupported language", false
}

func generateResetText(chat *llm.Chat) string {
	m := "有什么可以帮助您的吗？"
	return m
}

func generateHelpText(chat *llm.Chat) string {
	m := "bot 命令使用文档\n"
	for _, v := range availableCommands {
		m += fmt.Sprintf("/%s - %s\n", v.Command, v.Descriptions[1])
	}
	return m
}

func generateSettingsText(chat *llm.Chat) string {
	m := fmt.Sprintf("UserID: %d\nLanguage: %s\nHistory:\n%s", chat.UserID, chat.LanguageCode.Description(), chat.FormatHistory())
	return m
}
