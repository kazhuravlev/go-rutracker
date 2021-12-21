package rutracker

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kazhuravlev/go-rutracker/v2/parser"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

const (
	ForumTypeCategory ForumType = iota
	ForumTypeForum
)

func (c *Client) GetForumTree(ctx context.Context) ([]Forum, error) {
	u := "http://api.rutracker.org/v1/static/cat_forum_tree"

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
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
	var i int
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

func (c *Client) GetTopicsByForumID(ctx context.Context, forumID string) ([]Topic, error) {
	u := "http://api.rutracker.org/v1/static/pvc/f/" + forumID

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
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
	var i int
	for topicID, stat := range r.Result {
		res[i] = Topic{
			ID:      topicID,
			Seeders: stat[1],
		}
		i += 1
	}

	return res, nil
}

func (c *Client) GetFullTopic(ctx context.Context, topicIDs []string) ([]FullTopic, error) {
	query := url.Values{}
	query.Set("by", "topic_id")
	query.Set("val", strings.Join(topicIDs, ","))
	u := "http://api.rutracker.org/v1/get_tor_topic_data?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
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
			Size:     int(info.Size),
			ForumID:  strconv.Itoa(info.ForumId),
			Hash:     info.InfoHash,
			AuthorID: strconv.Itoa(info.AuthorID),
		})
	}

	return res, nil
}

func (c *Client) GetTopicMeta(ctx context.Context, topicID string) (*parser.TopicMeta, error) {
	query := url.Values{}
	query.Set("t", topicID)
	u := "https://rutracker.org/forum/viewtopic.php?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
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
