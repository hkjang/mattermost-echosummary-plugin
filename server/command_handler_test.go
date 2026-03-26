package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/stretchr/testify/assert"
)

func newTestPlugin() (*Plugin, *plugintest.API) {
	api := &plugintest.API{}
	driver := &plugintest.Driver{}

	return &Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    api,
			Driver: driver,
		},
		client: pluginapi.NewClient(api, driver),
	}, api
}

func TestHandleCommandSetTimes(t *testing.T) {
	t.Run("saves a single slot for the calling user", func(t *testing.T) {
		plugin, api := newTestPlugin()
		api.On("UpdatePreferencesForUser", "user-1", []model.Preference{{
			UserId:   "user-1",
			Category: userPreferenceCategory,
			Name:     userPreferenceDeliveryTimes,
			Value:    "09:00",
		}}).Return(nil).Once()

		response, err := plugin.handleCommand(&model.CommandArgs{
			UserId:  "user-1",
			Command: "/echosummary set-times 09:00",
		})

		assert.NoError(t, err)
		assert.Equal(t, model.CommandResponseTypeEphemeral, response.ResponseType)
		assert.Equal(t, "개인 발송 시간이 저장되었습니다.", response.Text)
		api.AssertExpectations(t)
	})

	t.Run("normalizes multiple values before saving", func(t *testing.T) {
		plugin, api := newTestPlugin()
		api.On("UpdatePreferencesForUser", "user-1", []model.Preference{{
			UserId:   "user-1",
			Category: userPreferenceCategory,
			Name:     userPreferenceDeliveryTimes,
			Value:    "09:00,13:30",
		}}).Return(nil).Once()

		response, err := plugin.handleCommand(&model.CommandArgs{
			UserId:  "user-1",
			Command: "/echosummary set-times 13:30 09:00,13:30",
		})

		assert.NoError(t, err)
		assert.Equal(t, "개인 발송 시간이 저장되었습니다.", response.Text)
		api.AssertExpectations(t)
	})
}
