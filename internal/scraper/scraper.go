package scraper

import (
	"context"
	"encoding/gob"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"log"
	"os"
	"strings"
)

type Scraper struct {
	baseUrl    string
	ctx        context.Context
	cancel     func()
	noticeUris []string
}

func NewScraper(baseUrl string) (*Scraper, error) {
	ctx, cancel := chromedp.NewContext(context.Background())

	file, err := os.Open("uris")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var uris []string
	if file != nil {
		if err := gob.NewDecoder(file).Decode(&uris); err != nil {
			return nil, errors.Wrap(err, "failed to decode uris file")
		}
	}

	return &Scraper{baseUrl, ctx, cancel, uris}, nil
}

func (s *Scraper) Stop() {
	s.cancel()

	file, err := os.OpenFile("uris", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	if file != nil {
		if err := gob.NewEncoder(file).Encode(s.noticeUris); err != nil {
			log.Println(err)
		}
	}
}

func (s *Scraper) hasNotice(notice Notice) bool {
	for _, uri := range s.noticeUris {
		if notice.Uri == uri {
			return true
		}
	}

	return false
}

func (s *Scraper) GetNewNotices(pageId string) ([]Notice, error) {
	var nodes []*cdp.Node

	if err := chromedp.Run(
		s.ctx,
		chromedp.Navigate(s.baseUrl+"/"+pageId),
		chromedp.Nodes(".notion-selectable.notion-page-block.notion-collection-item > a:nth-child(1)", &nodes, chromedp.NodeVisible),
	); err != nil {
		return nil, err
	}

	var newNotices []Notice
	for _, node := range nodes {
		var text string
		if err := chromedp.Run(
			s.ctx,
			chromedp.Text([]cdp.NodeID{node.NodeID}, &text, chromedp.ByNodeID),
		); err != nil {
			return nil, err
		}

		uri, _ := node.Attribute("href")
		split := strings.Split(text, "\n")
		notice := noticeFromArray(split, uri)

		if !s.hasNotice(notice) {
			newNotices = append(newNotices, notice)
		}
	}

	for _, notice := range newNotices {
		s.noticeUris = append(s.noticeUris, notice.Uri)
	}

	return newNotices, nil
}

func (s *Scraper) GetNoticeDetail(notice Notice) (NoticeDetail, error) {
	return NoticeDetail{}, nil
}
