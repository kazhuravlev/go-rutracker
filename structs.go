package rutracker

//go:generate stringer -type=ForumType
type ForumType int

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
	Size     int
	Seeders  int
	Title    string
}

type respFullTopicList struct {
	Result map[string]*struct {
		InfoHash       string  `json:"info_hash"`        //"info_hash": "658EDAB6AF0B424E62FEFEC0E39DBE2AC55B9AE3",
		ForumId        int     `json:"forum_id"`         //"forum_id": 9,
		AuthorID       int     `json:"poster_id"`        //"poster_id": 670,
		Size           float64 `json:"size"`             //"size": 5020938240,
		RegTime        int     `json:"reg_time"`         //"reg_time": 1112928696,
		TorStatus      int     `json:"tor_status"`       //"tor_status": 2,
		Seeders        int     `json:"seeders"`          //"seeders": 1,
		TopicTitle     string  `json:"topic_title"`      //"topic_title": "Гражданин начальник / Сезон: 1 / Серии: 1-15 из 15 (Николай Досталь) [2001, драма, криминал, TVRip]",
		SeederLastSeen int     `json:"seeder_last_seen"` //"seeder_last_seen": 1509589261
	} `json:"result"`
}
