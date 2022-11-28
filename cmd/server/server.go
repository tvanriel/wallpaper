package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/spf13/viper"
	"github.com/tvanriel/wallpaper"
	"go.uber.org/fx"
)

type Configuration struct {
	Directory     string
	ListenAddress string

	BotToken string
	Channel  string
}

func NewConfiguration() (*Configuration, error) {

	viper.SetDefault("Listen", "0.0.0.0:8080")
	viper.SetDefault("Directory", "/opt/wallpapers/assets")
	viper.SetDefault("Bot_Token", "")
	viper.SetDefault("Channel", "00000000000000")
	viper.AddConfigPath("/opt/wallpapers/config/")

	viper.BindEnv("Listen", "LISTEN")
	viper.BindEnv("Channel", "CHANNEL")
	viper.BindEnv("Bot_Token", "BOT_TOKEN")
	viper.BindEnv("Directory", "DIRECTORY")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()

	if err != nil {
		viper.SafeWriteConfig()
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {

			return nil, err
		}
	}

	return &Configuration{
		Directory:     viper.GetString("Directory"),
		ListenAddress: viper.GetString("Listen"),

		BotToken: viper.GetString("Bot_Token"),
		Channel:  viper.GetString("Channel"),
	}, nil

}

func NewWallpaperHandlerMux(wph *wallpaper.WallpaperHandler) http.Handler {
	return wph.Handler()
}

func NewWallpaperHandler(config *Configuration) *wallpaper.WallpaperHandler {
	return wallpaper.NewWallpaperHandler(config.Directory)
}

func NewHttpServe(lc fx.Lifecycle, config *Configuration, mux http.Handler) *http.Server {
	srv := &http.Server{
		Addr:    config.ListenAddress,
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			fmt.Println("Starting HTTP server at", srv.Addr)
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}

func NewDiscordSaver(lc fx.Lifecycle, config *Configuration) (*wallpaper.DiscordWallpaperSaverBot, error) {
	bot, err := wallpaper.NewDiscordSaver(config.BotToken, config.Directory)
	if err != nil {
		return nil, err
	}
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go bot.ListenForMessages()
				return nil
			},
		},
	)
	return bot, nil
}

func main() {
	app := fx.New(
		fx.Provide(
			NewConfiguration,
			NewHttpServe,
			NewWallpaperHandlerMux,
			NewWallpaperHandler,
			NewDiscordSaver,
		),
		fx.Invoke(func(*http.Server, *wallpaper.DiscordWallpaperSaverBot) {}),
	)

	app.Run()

}
