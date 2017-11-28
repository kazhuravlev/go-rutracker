package rutracker

import (
	"net/http"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"strconv"
	"html"
	"github.com/kazhuravlev/go-rutracker/parser"
)

var (
	ErrBadResponse = errors.New("bad response")
	ErrNotFound    = errors.New("object not found")
)

type Client struct {
	httpClient *http.Client
}

func New(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		httpClient: httpClient,
	}, nil
}

//go:generate stringer -type=ForumType
type ForumType int

const (
	ForumTypeCategory ForumType = iota
	ForumTypeForum
)

type Forum struct {
	ID    string
	Type  ForumType
	Title string
}

type respForumTree struct {
	Result struct {
		Categories map[string]string `json:"c"`
		Forums     map[string]string `json:"f"`
	}
}

type respTopicList struct {
	Result map[string][3]int `json:"result"`
}

type Topic struct {
	ID string
	//TorStatus int
	Seeders int
}

type FullTopic struct {
	ID       string
	Hash     string
	ForumID  string
	AuthorID string
	Size     float64
	Seeders  int
	Title    string
}

type respFullTopicList struct {
	Result map[string]*struct {
		InfoHash       string `json:"info_hash"`        //"info_hash": "658EDAB6AF0B424E62FEFEC0E39DBE2AC55B9AE3",
		ForumId        int    `json:"forum_id"`         //"forum_id": 9,
		AuthorID       int    `json:"poster_id"`        //"poster_id": 670,
		Size           int    `json:"size"`             //"size": 5020938240,
		RegTime        int    `json:"reg_time"`         //"reg_time": 1112928696,
		TorStatus      int    `json:"tor_status"`       //"tor_status": 2,
		Seeders        int    `json:"seeders"`          //"seeders": 1,
		TopicTitle     string `json:"topic_title"`      //"topic_title": "Гражданин начальник / Сезон: 1 / Серии: 1-15 из 15 (Николай Досталь) [2001, драма, криминал, TVRip]",
		SeederLastSeen int    `json:"seeder_last_seen"` //"seeder_last_seen": 1509589261
	} `json:"result"`
}

func (c *Client) GetForumTree() ([]Forum, error) {
	u := "http://api.rutracker.org/v1/static/cat_forum_tree"

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrBadResponse
	}

	var r respForumTree
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	res := make([]Forum, len(r.Result.Forums))
	//res := make([]Forum, len(r.Result.Categories)+len(r.Result.Forums))
	i := 0
	//for categoryID, categoryTitle := range r.Result.Categories {
	//	res[i] = Forum{
	//		ID:    categoryID,
	//		Type:  ForumTypeCategory,
	//		Title: categoryTitle,
	//	}
	//
	//	i += 1
	//}
	for forumID, forumTitle := range r.Result.Forums {
		res[i] = Forum{
			ID:    forumID,
			Type:  ForumTypeForum,
			Title: forumTitle,
		}

		i += 1
	}

	return res, nil
}

func (c *Client) GetTopicsByForumID(forumID string) ([]Topic, error) {
	u := "http://api.rutracker.org/v1/static/pvc/f/" + forumID

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrBadResponse
	}

	var r respTopicList
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	res := make([]Topic, len(r.Result))
	i := 0
	for topicID, stat := range r.Result {
		res[i] = Topic{
			ID:      topicID,
			Seeders: stat[1],
		}
		i += 1
	}

	return res, nil
}

func (c *Client) GetFullTopic(topicIDs []string) ([]FullTopic, error) {
	query := url.Values{}
	query.Set("by", "topic_id")
	query.Set("val", strings.Join(topicIDs, ","))
	u := "http://api.rutracker.org/v1/get_tor_topic_data?" + query.Encode()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrBadResponse
	}

	var r respFullTopicList
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	var res []FullTopic
	for topicID, info := range r.Result {
		if info == nil {
			continue
		}

		res = append(res, FullTopic{
			ID:       topicID,
			Seeders:  info.Seeders,
			Title:    html.UnescapeString(info.TopicTitle),
			Size:     info.Size,
			ForumID:  strconv.Itoa(info.ForumId),
			Hash:     info.InfoHash,
			AuthorID: strconv.Itoa(info.AuthorID),
		})
	}

	return res, nil
}

func (c *Client) GetTopicMeta(topicID string) (*parser.TopicMeta, error) {
	query := url.Values{}
	query.Set("t", topicID)
	u := "https://rutracker.org/forum/viewtopic.php?" + query.Encode()

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrBadResponse
	}

	p, err := parser.NewParser()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return p.ParseTopicPage(resp.Body)
}

//func (c *Client) GetTorrentFile(topicIDs []string) ([]FullTopic, error) {
//	query := url.Values{}
//	query.Set("by", "topic_id")
//	query.Set("val", strings.Join(topicIDs, ","))
//	u := "http://api.rutracker.org/v1/get_tor_topic_data?" + query.Encode()
//
//	req, err := http.NewRequest("GET", u, nil)
//	if err != nil {
//		return nil, err
//	}
//
//	resp, err := c.httpClient.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return nil, ErrBadResponse
//	}
//
//	var r respFullTopicList
//	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
//		return nil, err
//	}
//
//	var res []FullTopic
//	for topicID, info := range r.Result {
//		if info == nil {
//			continue
//		}
//
//		res = append(res, FullTopic{
//			ID:       topicID,
//			Seeders:  info.Seeders,
//			Title:    info.TopicTitle,
//			Size:     info.Size,
//			ForumID:  strconv.Itoa(info.ForumId),
//			Hash:     info.InfoHash,
//			AuthorID: strconv.Itoa(info.AuthorID),
//		})
//	}
//
//	return res, nil
//}
