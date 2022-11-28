package wallpaper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
)

type DiscordWallpaperSaverBot struct {
	client    *disgord.Client
	directory string
}

func NewDiscordSaver(botToken string, directory string) (*DiscordWallpaperSaverBot, error) {
	if botToken == "" {
		return nil, errors.New("no bot token was provided where one was expected")
	}
	discordClient := disgord.New(disgord.Config{
		Intents:  disgord.AllIntents(),
		BotToken: botToken,
	})

	bot := &DiscordWallpaperSaverBot{
		client:    discordClient,
		directory: directory,
	}
	bot.addMessageListeners()
	return bot, nil

}

func (d *DiscordWallpaperSaverBot) ListenForMessages() error {
	return d.client.Gateway().StayConnectedUntilInterrupted()
}

func (d *DiscordWallpaperSaverBot) addMessageListeners() {
	content, _ := std.NewMsgFilter(context.Background(), d.client)
	d.client.Gateway().
		WithMiddleware(content.NotByBot,
			content.NotByWebhook,
			content.ContainsBotMention,
		).
		MessageCreate(func(s disgord.Session, evt *disgord.MessageCreate) {
			attachments := evt.Message.Attachments
			for attachment := range attachments {
				err := d.download(
					attachments[attachment].URL,
					attachments[attachment].Filename,
				)
				if err != nil {
					evt.Message.Reply(context.Background(), s, err.Error())
					continue
				}

				evt.Message.Reply(context.Background(), s,
					fmt.Sprintf(":white_check_mark: Downloaded `%v` successfully.",
						strings.ReplaceAll(attachments[attachment].Filename, "`", ""),
					),
				)
			}
		})
}

func (d *DiscordWallpaperSaverBot) download(url string, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	file, err := os.Create(
		filepath.Join(
			d.directory,
			strings.Join([]string{
				strconv.Itoa(int(time.Now().UnixMilli())),
				filename,
			}, "-"),
		),
	)
	if err != nil {
		return err
	}
	io.Copy(file, resp.Body)
	return nil
}
