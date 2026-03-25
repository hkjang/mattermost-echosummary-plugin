package main

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

func (p *Plugin) runJob() {
	cfg := p.getConfiguration().normalized()
	if !cfg.isConfigured() {
		p.API.LogDebug("Echo Summary skipped scheduled run because configuration is incomplete")
		return
	}

	users, err := p.listEligibleUsers(cfg)
	if err != nil {
		p.API.LogError("Failed to list eligible users", "err", err)
		return
	}

	location, err := loadScheduleLocation(cfg.NotificationTimezone)
	if err != nil {
		p.API.LogError("Failed to load configured timezone", "timezone", cfg.NotificationTimezone, "err", err)
		location = time.FixedZone("KST", 9*60*60)
	}

	now := time.Now().In(location)
	for _, user := range users {
		settings, err := p.getUserDeliverySettings(user.Id, cfg)
		if err != nil {
			p.API.LogError("Failed to load user delivery settings", "user_id", user.Id, "err", err)
			continue
		}
		if settings.Disabled || len(settings.Slots) == 0 {
			continue
		}

		for _, slot := range dueDeliverySlots(now, settings.Slots, 2*time.Minute) {
			summaryDate := now.AddDate(0, 0, -1).Format("2006-01-02")
			sentKey := buildSentStateKey(user.Id, summaryDate, slot)

			sent, err := p.wasDeliveryRecorded(sentKey)
			if err != nil {
				p.API.LogError("Failed to read delivery state", "user_id", user.Id, "slot", slot, "err", err)
				continue
			}
			if sent {
				continue
			}

			if err := p.sendSummaryToUser(user, now, cfg); err != nil {
				p.API.LogError("Failed to send scheduled summary", "user_id", user.Id, "slot", slot, "err", err)
				continue
			}

			if err := p.recordDelivery(sentKey, now); err != nil {
				p.API.LogError("Failed to record delivery state", "user_id", user.Id, "slot", slot, "err", err)
			}
			p.API.LogInfo("Sent scheduled echo summary", "user_id", user.Id, "slot", slot)
		}
	}
}

func buildSentStateKey(userID, summaryDate, slot string) string {
	return fmt.Sprintf("%s%s:%s:%s", sentStatePrefix, userID, summaryDate, slot)
}

func (p *Plugin) wasDeliveryRecorded(key string) (bool, error) {
	var payload map[string]int64
	if err := p.client.KV.Get(key, &payload); err != nil {
		return false, err
	}
	return len(payload) > 0, nil
}

func (p *Plugin) recordDelivery(key string, sentAt time.Time) error {
	_, err := p.client.KV.Set(key, map[string]int64{
		"sent_at": model.GetMillisForTime(sentAt),
	}, pluginapi.SetExpiry(7*24*time.Hour))
	return err
}
