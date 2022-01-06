package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"likelion-notice-bot/internal/scraper"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const baseUrl = "https://likelion.notion.site"
const noticePageId = "c1a67aaed0374963b86fcb3109c62644"

type Bot struct {
	session       *discordgo.Session
	noticeChannel string
	scraper       *scraper.Scraper
	cron          *cron.Cron
	cronEntry     *cron.EntryID
}

func readFileToString(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	s, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(s)), nil
}

func New() (*Bot, error) {
	token, err := readFileToString(".token")
	if err != nil {
		return nil, err
	}

	channel, err := readFileToString(".chan")
	if err != nil {
		return nil, err
	}

	session, err := discordgo.New("Bot " + string(token))
	if err != nil {
		return nil, err
	}

	s, err := scraper.NewScraper(baseUrl)
	if err != nil {
		return nil, err
	}

	return &Bot{
		session:       session,
		noticeChannel: channel,
		scraper:       s,
	}, nil
}

func (b *Bot) SendToChannel(message string) error {
	c, err := b.session.Channel(b.noticeChannel)
	if err != nil {
		return err
	}

	if c == nil {
		return fmt.Errorf("channel with id '%s' does not exists", b.noticeChannel)
	}

	if _, err := b.session.ChannelMessageSend(c.ID, message); err != nil {
		return errors.Wrap(err, "failed to send message to desired channel")
	}

	return nil
}

func (b *Bot) checkNotices() error {
	notices, err := b.scraper.GetNewNotices(noticePageId)
	if err != nil {
		return err
	}

	for _, notice := range notices {
		if err := b.SendToChannel(notice.String() + "\n" + baseUrl + notice.Uri); err != nil {
			return err
		}
	}

	return nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return err
	}

	if err := b.checkNotices(); err != nil {
		return err
	}

	b.cron = cron.New()
	entry, err := b.cron.AddFunc("@hourly", func() {
		if err := b.checkNotices(); err != nil {
			log.Println(err)
		}
	})
	if err != nil {
		return err
	}

	b.cronEntry = &entry
	b.cron.Start()

	log.Println("bot is now online")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Println("bot is now shutting down")

	b.scraper.Stop()
	return b.session.Close()
}
