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
	ctx := context.Background()

	c, _ := rutracker.New(nil)
	tree, err := c.GetForumTree(ctx)
	assert.Nil(t, err)
	require.NotNil(t, tree)

	assert.True(t, len(tree) >= 1)
}

func TestClient_GetTopicsByForumID(t *testing.T) {
	ctx := context.Background()

	c, _ := rutracker.New(nil)

	tree, _ := c.GetForumTree(ctx)

	topics, err := c.GetTopicsByForumID(ctx, tree[0].ID)
	fmt.Println(tree[0].ID)
	assert.Nil(t, err)
	require.NotNil(t, topics)

	assert.True(t, len(topics) >= 1)
}

func TestClient_GetFullTopic(t *testing.T) {
	ctx := context.Background()

	c, _ := rutracker.New(nil)

	forums, _ := c.GetForumTree(ctx)

	var topics []rutracker.Topic
	for _, forum := range forums {
		topics, _ = c.GetTopicsByForumID(ctx, forum.ID)
		if len(topics) >= 3 {
			break
		}
	}
	require.True(t, len(topics) >= 3)

	fullTopics, err := c.GetFullTopic(ctx, []string{topics[0].ID, topics[1].ID, topics[2].ID})
	fmt.Println(err)
	assert.Nil(t, err)
	require.NotNil(t, fullTopics)

	fmt.Println(fullTopics)
}
