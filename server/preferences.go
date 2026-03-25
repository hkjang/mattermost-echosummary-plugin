package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

type deliverySettings struct {
	Slots    []string
	Disabled bool
	Source   string
}

func loadScheduleLocation(name string) (*time.Location, error) {
	locationName := strings.TrimSpace(name)
	if locationName == "" {
		locationName = defaultNotificationTimezone
	}
	return time.LoadLocation(locationName)
}

func parseTimeSlots(raw string) ([]string, error) {
	tokens := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})

	slots := make([]string, 0, len(tokens))
	seen := map[string]struct{}{}
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		parsed, err := time.Parse("15:04", token)
		if err != nil {
			return nil, fmt.Errorf("invalid time slot %q; use HH:mm", token)
		}
		slot := parsed.Format("15:04")
		if _, ok := seen[slot]; ok {
			continue
		}
		seen[slot] = struct{}{}
		slots = append(slots, slot)
	}

	sort.Strings(slots)
	return slots, nil
}

func normalizeDeliveryPreference(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	switch strings.ToLower(trimmed) {
	case userPreferenceDisabledValue, "disable", "disabled":
		return userPreferenceDisabledValue, nil
	}

	slots, err := parseTimeSlots(trimmed)
	if err != nil {
		return "", err
	}
	if len(slots) == 0 {
		return "", nil
	}
	return strings.Join(slots, ","), nil
}

func dueDeliverySlots(now time.Time, slots []string, grace time.Duration) []string {
	due := make([]string, 0, len(slots))
	for _, slot := range slots {
		parsed, err := time.Parse("15:04", slot)
		if err != nil {
			continue
		}
		scheduled := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
		if now.Before(scheduled) {
			continue
		}
		if now.Sub(scheduled) <= grace {
			due = append(due, slot)
		}
	}
	return due
}

func nextDeliveryTime(now time.Time, slots []string) string {
	if len(slots) == 0 {
		return "설정 없음"
	}

	for dayOffset := 0; dayOffset <= 1; dayOffset++ {
		day := now.AddDate(0, 0, dayOffset)
		for _, slot := range slots {
			parsed, err := time.Parse("15:04", slot)
			if err != nil {
				continue
			}
			scheduled := time.Date(day.Year(), day.Month(), day.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
			if scheduled.After(now) {
				return scheduled.Format("2006-01-02 15:04 MST")
			}
		}
	}

	return "설정 없음"
}

func (p *Plugin) getUserDeliverySettings(userID string, cfg *configuration) (deliverySettings, error) {
	defaults := deliverySettings{Source: "default"}

	defaultSlots, err := parseTimeSlots(cfg.DefaultTimeSlots)
	if err != nil {
		return defaults, errors.Wrap(err, "failed to parse default delivery times")
	}
	defaults.Slots = defaultSlots

	preference, appErr := p.API.GetPreferenceForUser(userID, userPreferenceCategory, userPreferenceDeliveryTimes)
	if appErr != nil {
		if appErr.StatusCode == 404 {
			return defaults, nil
		}
		return defaults, errors.Wrap(appErr, "failed to get user preference")
	}

	value := strings.TrimSpace(preference.Value)
	if value == "" {
		return defaults, nil
	}
	if strings.EqualFold(value, userPreferenceDisabledValue) {
		return deliverySettings{
			Disabled: true,
			Source:   "user",
		}, nil
	}

	slots, err := parseTimeSlots(value)
	if err != nil {
		return defaults, errors.Wrap(err, "failed to parse user delivery times")
	}

	return deliverySettings{
		Slots:  slots,
		Source: "user",
	}, nil
}

func (p *Plugin) saveUserDeliveryPreference(userID, raw string) error {
	value, err := normalizeDeliveryPreference(raw)
	if err != nil {
		return err
	}
	if value == "" {
		return errors.New("at least one HH:mm time slot is required")
	}

	return p.API.UpdatePreferencesForUser(userID, []model.Preference{{
		UserId:   userID,
		Category: userPreferenceCategory,
		Name:     userPreferenceDeliveryTimes,
		Value:    value,
	}})
}

func (p *Plugin) disableUserDeliveryPreference(userID string) error {
	return p.API.UpdatePreferencesForUser(userID, []model.Preference{{
		UserId:   userID,
		Category: userPreferenceCategory,
		Name:     userPreferenceDeliveryTimes,
		Value:    userPreferenceDisabledValue,
	}})
}

func (p *Plugin) clearUserDeliveryPreference(userID string) error {
	return p.API.DeletePreferencesForUser(userID, []model.Preference{{
		UserId:   userID,
		Category: userPreferenceCategory,
		Name:     userPreferenceDeliveryTimes,
	}})
}
