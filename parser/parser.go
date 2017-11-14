package parser

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/Sirupsen/logrus"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"
	"bytes"
)

var (
	urlViewForum = "https://rutracker.org/forum/viewforum.php"
)

type Parser struct {
	log *logrus.Entry
}

func NewParser() (*Parser, error) {
	return &Parser{
		log: logrus.New().WithField("module", "parser"),
	}, nil
}

func (p *Parser) ParseCatalog(r io.Reader) ([]*url.URL, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var res []*url.URL
	doc.Find("select#fs-main optgroup option").Each(func(i int, s *goquery.Selection) {
		forumID, exists := s.Attr("value")
		if !exists {
			return
		}

		u, err := url.Parse(fmt.Sprintf(urlViewForum+"?f=%s", forumID))
		if err != nil {
			return
		}

		res = append(res, u)
	})

	return res, nil
}

type TopicPreview struct {
	URL      string
	Title    string
	Seeders  int
	Leechers int
}

func (p *Parser) ParseTopicList(r io.Reader) ([]TopicPreview, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var res []TopicPreview
	doc.Find("tr.hl-tr").Each(func(i int, s *goquery.Selection) {
		var forum TopicPreview
		{
			leechersQ := s.Find("td.leechmed b").First()
			if leechersQ.Length() > 0 {
				var err error
				forum.Leechers, err = strconv.Atoi(leechersQ.Text())
				if err != nil {
					p.log.WithError(err).Warn("Cannot get leechers")
				}
			}
		}
		{
			seedersQ := s.Find("td b.seedmed").First()
			if seedersQ.Length() > 0 {
				var err error
				forum.Seeders, err = strconv.Atoi(seedersQ.Text())
				if err != nil {
					p.log.WithError(err).Warn("Cannot get seeders")
				}
			}
		}
		{
			titleQ := s.Find(".t-title a").First()
			if titleQ.Length() > 0 {
				forum.Title = titleQ.Text()
				forum.Title = strings.Replace(forum.Title, "\n", " ", -1)
				forum.Title = strings.Trim(forum.Title, " ")

				u, exists := titleQ.Attr("href")
				if exists {
					forum.URL = u
				}
			}
		}

		res = append(res, forum)
	})

	return res, nil
}

type RawPage struct {
	Body io.Reader
}

type TopicMeta struct {
	RawPage
	TopicPreview
	PosterURL   string
	MagnetLink  string
	KinopoiskID string
	IMDbID      string
}

func (p *Parser) ParseTopicPage(r io.Reader) (*TopicMeta, error) {
	document, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var res TopicMeta
	{
		metaTable := document.Find(".attach.bordered.med").First()
		magnetLinkQ := metaTable.Find(".magnet-link").First()
		if magnetLinkQ.Length() > 0 {
			magnetLink, exists := magnetLinkQ.Attr("href")
			if exists {
				res.MagnetLink = magnetLink
			}
		}
	}

	// изображение/постер. если есть
	{
		posterQ := document.Find("var.postImg.postImgAligned.img-right").First()
		if posterQ.Length() > 0 {
			posterURL, exists := posterQ.Attr("title")
			if exists {
				u, err := url.Parse(posterURL)
				if err == nil {
					res.PosterURL = u.String()
				}
			}
		}
	}

	// идентификатор кинопоиска через шильдик КП
	{
		kinopoiskIDQ := document.Find("var[title*=kinopoisk\\.ru\\/rating]").First()
		if kinopoiskIDQ.Length() > 0 {
			kinopoiskImgURL, exists := kinopoiskIDQ.Attr("title")
			if exists {
				u, err := url.Parse(kinopoiskImgURL)
				if err == nil {
					_, fileName := path.Split(u.Path)
					extension := ".gif"
					if strings.HasSuffix(fileName, extension) {
						res.KinopoiskID = fileName[:len(fileName)-len(extension)]
					}
				}
			}
		}
	}

	// идентификатор кинопоиска через ссылку на страницу КП
	if len(res.KinopoiskID) == 0 {
		kinopoiskIDQ := document.Find("a[href*=kinopoisk\\.ru\\/film]").First()
		if kinopoiskIDQ.Length() > 0 {
			kinopoiskFilmURL, exists := kinopoiskIDQ.Attr("href")
			if exists {
				u, err := url.Parse(kinopoiskFilmURL)
				if err == nil {
					_, fileName := path.Split(strings.Trim(u.Path, "/"))
					res.KinopoiskID = fileName
				}
			}
		}
	}

	// идентификатор IMDb через ссылку на страницу
	{
		IMDbQ := document.Find("a[href*=imdb\\.com\\/title]").First()
		if IMDbQ.Length() > 0 {
			IMDbURL, exists := IMDbQ.Attr("href")
			if exists {
				u, err := url.Parse(IMDbURL)
				if err == nil {
					_, fileName := path.Split(strings.Trim(u.Path, "/"))
					res.IMDbID = fileName
				}
			}
		}
	}

	topicBody := document.Find("#topic_main>tbody:nth-child(2)").First()
	if topicBody.Length() == 0 {
		panic("unknown error")
	}

	htmlText, err := topicBody.Html()
	if  err != nil {
		return nil, err
	}

	res.RawPage.Body = bytes.NewBufferString(htmlText)

	return &res, nil
}
