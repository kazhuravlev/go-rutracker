package parser_test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
	"github.com/kazhuravlev/go-rutracker/parser"
)

func TestParser_ParseCatalog(t *testing.T) {
	data, err := ioutil.ReadFile("./testdata/catalog.html")
	fmt.Println(err)

	p, _ := parser.NewParser()

	fmt.Println(p.ParseCatalog(bytes.NewBuffer(data)))
}

func TestParser_ParseTopicList(t *testing.T) {
	data, err := ioutil.ReadFile("./testdata/forum_list.html")
	fmt.Println(err)

	p, _ := parser.NewParser()

	fmt.Println(p.ParseTopicList(bytes.NewBuffer(data)))
}

func TestParser_ParseTopicPage(t *testing.T) {
	data, err := ioutil.ReadFile("./testdata/topic.html")
	fmt.Println(err)

	p, _ := parser.NewParser()

	topic, err := p.ParseTopicPage(bytes.NewBuffer(data))
	assert.Nil(t, err)
	require.NotNil(t, topic)

	assert.Equal(t, "http://i3.imageban.ru/out/2016/09/23/2e2f4cc98059082418fda972cbb16562.jpg", topic.PosterURL)
	assert.Equal(t, "magnet:THIS_IS_TEST_LINK", topic.MagnetLink)
	assert.Equal(t, "843231", topic.KinopoiskID)
	assert.Equal(t, "tt4176370", topic.IMDbID)

	res, err := ioutil.ReadAll(topic.Body)
	assert.Nil(t, err)
	require.NotNil(t, res)

	exp := []byte(`<tr><td class="poster_info td1 hide-for-print"><a id="73528050">`)
	assert.Equal(t, exp, res[:len(exp)])
}
