package main

import (
	"fmt"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMoveThreadCommand(t *testing.T) {
	originalChannel := &model.Channel{
		Id:   model.NewId(),
		Name: "original-channel",
	}

	targetTeam := &model.Team{
		Id:   model.NewId(),
		Name: "target-team",
	}
	targetChannel := &model.Channel{
		Id:   model.NewId(),
		Name: "target-channel",
	}

	api := &plugintest.API{}
	api.On("GetPostThread", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(mockGeneratePostList(3, originalChannel.Id, false), nil)
	api.On("GetChannelMember", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(mockGenerateChannelMember(), nil)
	api.On("GetChannel", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(targetChannel, nil)
	api.On("GetTeam", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(targetTeam, nil)
	api.On("CreatePost", mock.Anything, mock.Anything).Return(mockGeneratePost(), nil)
	api.On("DeletePost", mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)

	var plugin Plugin
	plugin.SetAPI(api)

	t.Run("no args", func(t *testing.T) {
		resp, isUserError, err := plugin.runMoveThreadCommand([]string{}, &model.CommandArgs{ChannelId: originalChannel.Id})
		require.NoError(t, err)
		assert.True(t, isUserError)
		assert.Contains(t, resp.Text, "Error: missing arguments")
	})

	t.Run("one arg", func(t *testing.T) {
		resp, isUserError, err := plugin.runMoveThreadCommand([]string{"id1"}, &model.CommandArgs{ChannelId: originalChannel.Id})
		require.NoError(t, err)
		assert.True(t, isUserError)
		assert.Contains(t, resp.Text, "Error: missing arguments")
	})

	t.Run("move thread successfully", func(t *testing.T) {
		resp, isUserError, err := plugin.runMoveThreadCommand([]string{"id1", "id2"}, &model.CommandArgs{ChannelId: originalChannel.Id})
		require.NoError(t, err)
		assert.False(t, isUserError)
		assert.Contains(t, resp.Text, fmt.Sprintf("A thread with %d posts has been moved [ team=%s, channel=%s ]", 3, targetTeam.Name, targetChannel.Name))
	})

	t.Run("not in thread channel", func(t *testing.T) {
		resp, isUserError, err := plugin.runMoveThreadCommand([]string{"id1", "id2"}, &model.CommandArgs{ChannelId: model.NewId()})
		require.NoError(t, err)
		assert.True(t, isUserError)
		assert.Contains(t, resp.Text, "Error: the move command must be run from the channel containing the post")
	})

	t.Run("thread is above configuration move-maximum", func(t *testing.T) {
		plugin.setConfiguration(&configuration{MaxThreadCountMoveSize: "1"})
		require.NoError(t, plugin.configuration.IsValid())
		resp, isUserError, err := plugin.runMoveThreadCommand([]string{"id1", "id2"}, &model.CommandArgs{ChannelId: model.NewId()})
		require.NoError(t, err)
		assert.True(t, isUserError)
		assert.Contains(t, resp.Text, "Error: the thread is 3 posts long, but the move thead command is configured to only move threads of up to 1 posts")
	})
}

func mockGeneratePostList(total int, channelID string, systemMessages bool) *model.PostList {
	postList := model.NewPostList()
	for i := 0; i < total; i++ {
		id := model.NewId()
		post := &model.Post{
			Id:        id,
			ChannelId: channelID,
			Message:   fmt.Sprintf("This is message %d", i),
		}
		if systemMessages {
			post.Type = model.POST_SYSTEM_MESSAGE_PREFIX
		}
		postList.AddPost(post)
		postList.AddOrder(id)
	}

	return postList
}

func mockGenerateChannelMember() *model.ChannelMember {
	return &model.ChannelMember{
		ChannelId: model.NewId(),
	}
}

func mockGeneratePost() *model.Post {
	return &model.Post{
		Id: model.NewId(),
	}
}
