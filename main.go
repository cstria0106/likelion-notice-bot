package main

import (
	"likelion-notice-bot/internal/bot"
	"log"
)

func main() {
	b, err := bot.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := b.Start(); err != nil {
		log.Fatal(err)
	}
}
