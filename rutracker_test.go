package rutracker_test

import (
	"context"
	"fmt"
	"github.com/kazhuravlev/go-rutracker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient_GetForumTree(t *testing.T) {
	c, _ := rutracker.New(nil)

	tree, err := c.GetForumTree(context.Background())
	assert.Nil(t, err)
	require.NotNil(t, tree)

	assert.True(t, len(tree) >= 1)
}

func TestClient_GetTopicsByForumID(t *testing.T) {
	c, _ := rutracker.New(nil)

	tree, _ := c.GetForumTree(context.Background())

	topics, err := c.GetTopicsByForumID(context.Background(), tree[0].ID)
	fmt.Println(tree[0].ID)
	assert.Nil(t, err)
	require.NotNil(t, topics)

	assert.True(t, len(topics) >= 1)
}

func TestClient_GetFullTopic(t *testing.T) {
	c, _ := rutracker.New(nil)

	forums, _ := c.GetForumTree(context.Background())

	var topics []rutracker.Topic
	for _, forum := range forums {
		topics, _ = c.GetTopicsByForumID(context.Background(), forum.ID)
		if len(topics) >= 3 {
			break
		}
	}
	require.True(t, len(topics) >= 3)

	fullTopics, err := c.GetFullTopic(context.Background(), []string{topics[0].ID, topics[1].ID, topics[2].ID})
	fmt.Println(err)
	assert.Nil(t, err)
	require.NotNil(t, fullTopics)

	fmt.Println(fullTopics)
}
