package main

import (
	"testing"
	"time"

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
