package main

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/stretchr/testify/assert"
)

func TestParseTimeSlots(t *testing.T) {
	t.Run("normalizes and sorts unique values", func(t *testing.T) {
		slots, err := parseTimeSlots("13:30, 09:00,13:30")
		assert.NoError(t, err)
		assert.Equal(t, []string{"09:00", "13:30"}, slots)
	})

	t.Run("fails on invalid time", func(t *testing.T) {
		_, err := parseTimeSlots("25:00")
		assert.Error(t, err)
	})
}

func TestDueDeliverySlots(t *testing.T) {
	location := time.FixedZone("KST", 9*60*60)
	now := time.Date(2026, 3, 25, 9, 1, 0, 0, location)

	due := dueDeliverySlots(now, []string{"09:00", "13:00"}, 2*time.Minute)
	assert.Equal(t, []string{"09:00"}, due)
}

func TestGetUserDeliverySettings(t *testing.T) {
	t.Run("loads user-specific schedules independently", func(t *testing.T) {
		plugin, api := newTestPlugin()
		cfg := (&configuration{DefaultTimeSlots: "09:00"}).normalized()

		api.On("GetPreferenceForUser", "user-a", userPreferenceCategory, userPreferenceDeliveryTimes).Return(model.Preference{
			UserId:   "user-a",
			Category: userPreferenceCategory,
			Name:     userPreferenceDeliveryTimes,
			Value:    "08:30",
		}, nil).Once()
		api.On("GetPreferenceForUser", "user-b", userPreferenceCategory, userPreferenceDeliveryTimes).Return(model.Preference{
			UserId:   "user-b",
			Category: userPreferenceCategory,
			Name:     userPreferenceDeliveryTimes,
			Value:    "13:00,18:00",
		}, nil).Once()

		settingsA, err := plugin.getUserDeliverySettings("user-a", cfg)
		assert.NoError(t, err)
		assert.Equal(t, "user", settingsA.Source)
		assert.Equal(t, []string{"08:30"}, settingsA.Slots)

		settingsB, err := plugin.getUserDeliverySettings("user-b", cfg)
		assert.NoError(t, err)
		assert.Equal(t, "user", settingsB.Source)
		assert.Equal(t, []string{"13:00", "18:00"}, settingsB.Slots)

		api.AssertExpectations(t)
	})
}
