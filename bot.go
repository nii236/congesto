package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jmoiron/sqlx"
)

type Bot struct {
	bot           *tbot.BotAPI
	ch            tbot.UpdatesChannel
	conn          *sqlx.DB
	checkInterval time.Duration
}

func NewBot(conn *sqlx.DB, botToken string, checkInterval time.Duration) (*Bot, error) {
	logger.Info()
	bot, err := tbot.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	// bot.Debug = true

	u := tbot.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	return &Bot{bot, updates, conn, checkInterval}, nil
}

// RunChecker runs the checker regularly to determine when and who to notify
func (b *Bot) RunChecker() (*Server, error) {
	logger.Info("Starting bot checker")
	ticker := time.NewTicker(b.checkInterval)
	for {
		select {
		case <-ticker.C:
			err := b.Tick()
			if err != nil {
				logger.Error(err)
			}
		}
	}
}

// Check server status
func (b *Bot) Check(serverName string) (*Server, error) {
	regions, err := scrape(WorldStatusURL)
	if err != nil {
		return nil, err
	}

	var result *Server
	for _, region := range regions {
		for _, dc := range region.DataCentres {
			for _, server := range dc.Servers {
				if strings.ToLower(server.Name) == strings.ToLower(serverName) {
					result = server
				}
			}
		}
	}

	if result == nil {
		err = errors.New("World not found: " + serverName)
		return nil, err
	}

	return result, nil
}

// Notify the users when there has been a change
func (b *Bot) Notify(chatID int64, serverName string, category Category, creationAvailable bool) error {
	msgText := fmt.Sprintf(`
ALERT - CHANGE IN SERVER STATUS

Name: %s
Category: %s
Character creation available: %v`,
		serverName,
		category,
		creationAvailable,
	)
	msg := tbot.NewMessage(chatID, msgText)
	_, err := b.bot.Send(msg)
	if err != nil {
		logger.Error(err)
	}

	return nil
}

// Tick runs regularly to know when to notify subscribers
func (b *Bot) Tick() error {
	logger.Info("Running congesto ticker")
	regions, err := scrape(WorldStatusURL)
	if err != nil {
		return nil
	}

	subs, err := b.List()
	if err != nil {
		return nil
	}

	for _, sub := range subs {
		for _, region := range regions {
			for _, dc := range region.DataCentres {
				for _, server := range dc.Servers {
					if strings.ToLower(server.Name) == strings.ToLower(sub.ServerName) {
						logger.Infof("Checking %s for %s\n", server.Name, sub.UserName)
						if server.CreateCharacterAvailable != sub.CreationAvailable {
							logger.Info("Alerting user:", sub.UserName)
							err = b.Notify(int64(sub.ChatID), sub.ServerName, server.Category, server.CreateCharacterAvailable)
							if err != nil {
								logger.Error(err)
								continue
							}
							err = b.UpdateSubscription(int64(sub.ChatID), sub.ServerName, server.CreateCharacterAvailable)
							if err != nil {
								logger.Error(err)
								continue
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// Trigger a notification
func (b *Bot) Trigger(chatID int64, serverName string, category Category, available bool) error {
	return b.Tick()
}

// Subscribe a server
func (b *Bot) Subscribe(firstName, lastName, userName string, chatID int64, serverName string) error {
	q := "INSERT INTO subscribers (first_name, last_name, user_name, chat_id, server_name) VALUES (?, ?, ?, ?, ?);"
	_, err := b.conn.Exec(q, firstName, lastName, userName, chatID, serverName)
	if err != nil {
		return err
	}
	return nil
}

// Unsubscribe a server
func (b *Bot) Unsubscribe(chatID int64, serverName string) error {
	q := "DELETE FROM subscribers WHERE chat_id = ? AND server_name = ?"
	_, err := b.conn.Exec(q, chatID, serverName)
	if err != nil {
		return err
	}
	return nil
}

// UpdateSubscription of the chatID with the creation status
func (b *Bot) UpdateSubscription(chatID int64, serverName string, creationAvailable bool) error {
	q := "UPDATE subscribers SET creation_available = ? WHERE chat_id = ? AND server_name = ?"
	_, err := b.conn.Exec(q, creationAvailable, chatID, serverName)
	if err != nil {
		return err
	}
	return nil
}

type Subscription struct {
	FirstName         string `db:"first_name"`
	LastName          string `db:"last_name"`
	UserName          string `db:"user_name"`
	ChatID            int    `db:"chat_id"`
	ServerName        string `db:"server_name"`
	CreationAvailable bool   `db:"creation_available"`
}

// List subscribed servers
func (b *Bot) List() ([]*Subscription, error) {
	result := []*Subscription{}
	q := "SELECT first_name, last_name, user_name, chat_id, server_name, creation_available FROM subscribers"
	err := b.conn.Select(&result, q)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Run blocking updates for bot
func (b *Bot) Run() {
	logger.Info("Starting bot")
	for update := range b.ch {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		if len(update.Message.Text) < 1 {
			msgText := "Request was empty"
			msg := tbot.NewMessage(update.Message.Chat.ID, msgText)
			b.bot.Send(msg)
			continue
		}
		if update.Message.Text == "!trigger" {
			b.Trigger(update.Message.Chat.ID, "test", CategoryStandard, true)
			continue
		}
		if update.Message.Text == "!list" {
			subs, err := b.List()
			if err != nil {
				logger.Error(err)
			}
			result := []string{}
			for _, sub := range subs {
				if int64(sub.ChatID) == update.Message.Chat.ID {
					result = append(result, sub.ServerName)
				}
			}

			if len(result) == 0 {
				result = []string{"none"}
			}

			msgText := fmt.Sprintf("Subscriptions: %s", strings.Join(result, ", "))
			msg := tbot.NewMessage(update.Message.Chat.ID, msgText)
			_, err = b.bot.Send(msg)
			if err != nil {
				logger.Error(err)
				continue
			}
			continue
		}
		if strings.HasPrefix(update.Message.Text, "!subscribe ") {
			serverName := strings.TrimPrefix(update.Message.Text, "!subscribe ")
			err := b.Subscribe(
				update.Message.From.FirstName,
				update.Message.From.LastName,
				update.Message.From.UserName,
				update.Message.Chat.ID,
				serverName,
			)
			if err != nil {
				logger.Error(err)
				continue
			}
			msgText := fmt.Sprintf("Subscription successful: %s", serverName)
			msg := tbot.NewMessage(update.Message.Chat.ID, msgText)
			_, err = b.bot.Send(msg)
			if err != nil {
				logger.Error(err)
				continue
			}
			continue
		}

		if strings.HasPrefix(update.Message.Text, "!unsubscribe ") {
			serverName := strings.TrimPrefix(update.Message.Text, "!unsubscribe ")
			err := b.Unsubscribe(update.Message.Chat.ID, serverName)
			if err != nil {
				logger.Error(err)
				continue
			}
			msgText := fmt.Sprintf("Unsubscription successful: %s", serverName)
			msg := tbot.NewMessage(update.Message.Chat.ID, msgText)
			_, err = b.bot.Send(msg)
			if err != nil {
				logger.Error(err)
				continue
			}
			continue
		}

		if strings.HasPrefix(update.Message.Text, "!check") {
			serverName := strings.TrimPrefix(update.Message.Text, "!check ")
			result, err := b.Check(serverName)
			if err != nil {
				logger.Error(err)
				continue
			}
			msgText := fmt.Sprintf(`
Name: %s
Category: %s
Character creation available: %v`,
				result.Name,
				result.Category,
				result.CreateCharacterAvailable,
			)
			msg := tbot.NewMessage(update.Message.Chat.ID, msgText)
			_, err = b.bot.Send(msg)
			if err != nil {
				logger.Error(err)
				continue
			}
			continue
		}

		msgText := `
Commands: 
!list - List subscriptions
!unsubscribe [server_name] - Unsubscribe from notifications for server_name
!subscribe [server_name] - Subscribe from notifications for server_name
!check [server_name] - Check status of server_name immediately
`

		msg := tbot.NewMessage(update.Message.Chat.ID, msgText)
		_, err := b.bot.Send(msg)
		if err != nil {
			logger.Error(err)
			continue
		}
		continue
	}
}
