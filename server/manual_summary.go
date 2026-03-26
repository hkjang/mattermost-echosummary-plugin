package main

import (
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

type summaryProgressReporter struct {
	channelID   string
	lastMessage string
	plugin      *Plugin
	postID      string
	userID      string
}

func (p *Plugin) runManualSummary(userID, channelID string, referenceTime time.Time, cfg *configuration) {
	go func() {
		reporter, err := p.newSummaryProgressReporter(userID)
		if err != nil {
			p.API.LogWarn("Failed to create manual summary progress reporter", "user_id", userID, "err", err)
		}

		user, err := p.client.User.Get(userID)
		if err != nil {
			p.finishManualSummaryWithError(userID, channelID, reporter, errors.Wrap(err, "failed to load current user"))
			return
		}

		if err := p.sendSummaryToUserWithProgress(user, referenceTime, cfg, func(message string) {
			if reporter != nil {
				reporter.Update(message)
			}
		}); err != nil {
			p.finishManualSummaryWithError(userID, channelID, reporter, errors.Wrap(err, "failed to send manual summary"))
			return
		}

		if reporter != nil {
			reporter.Complete()
			return
		}

		p.sendEphemeralNotice(userID, channelID, "요약이 완료되었습니다. DM을 확인해 주세요.")
	}()
}

func (p *Plugin) finishManualSummaryWithError(userID, channelID string, reporter *summaryProgressReporter, err error) {
	p.API.LogError("Failed to run manual echo summary", "user_id", userID, "err", err)

	message := "Echo Summary 요청을 처리하지 못했습니다. 잠시 후 다시 시도해 주세요."
	if err != nil {
		message += "\n\n```text\n" + truncateForLog(err.Error(), 400) + "\n```"
	}

	if reporter != nil {
		reporter.Fail(message)
		return
	}

	if dmErr := p.sendBotDM(userID, message); dmErr != nil {
		p.API.LogWarn("Failed to send manual summary failure DM", "user_id", userID, "err", dmErr)
	}

	p.sendEphemeralNotice(userID, channelID, message)
}

func (p *Plugin) sendBotDM(userID, message string) error {
	if strings.TrimSpace(message) == "" {
		return nil
	}

	if p.botUserID == "" {
		if err := p.ensureBot(); err != nil {
			return errors.Wrap(err, "failed to ensure bot")
		}
	}

	return p.client.Post.DM(p.botUserID, userID, &model.Post{Message: message})
}

func (p *Plugin) sendEphemeralNotice(userID, channelID, message string) {
	if strings.TrimSpace(channelID) == "" || strings.TrimSpace(message) == "" {
		return
	}

	if p.botUserID == "" {
		if err := p.ensureBot(); err != nil {
			p.API.LogWarn("Failed to ensure bot for ephemeral notice", "user_id", userID, "err", err)
			return
		}
	}

	post := &model.Post{
		ChannelId: channelID,
		UserId:    p.botUserID,
		Message:   message,
	}
	p.client.Post.SendEphemeralPost(userID, post)
}

func (p *Plugin) newSummaryProgressReporter(userID string) (*summaryProgressReporter, error) {
	if p.botUserID == "" {
		if err := p.ensureBot(); err != nil {
			return nil, errors.Wrap(err, "failed to ensure bot")
		}
	}

	channel, appErr := p.API.GetDirectChannel(p.botUserID, userID)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get direct channel")
	}

	post := &model.Post{
		ChannelId: channel.Id,
		UserId:    p.botUserID,
		Message:   "요약 요청을 받았습니다. 준비를 시작합니다...",
	}
	if err := p.client.Post.CreatePost(post); err != nil {
		return nil, errors.Wrap(err, "failed to create progress post")
	}

	return &summaryProgressReporter{
		channelID:   channel.Id,
		lastMessage: post.Message,
		plugin:      p,
		postID:      post.Id,
		userID:      userID,
	}, nil
}

func (r *summaryProgressReporter) Update(message string) {
	if r == nil {
		return
	}

	trimmed := strings.TrimSpace(message)
	if trimmed == "" || trimmed == r.lastMessage {
		return
	}

	post := &model.Post{
		Id:        r.postID,
		ChannelId: r.channelID,
		UserId:    r.plugin.botUserID,
		Message:   trimmed,
	}
	if err := r.plugin.client.Post.UpdatePost(post); err != nil {
		r.plugin.API.LogWarn("Failed to update manual summary progress", "user_id", r.userID, "post_id", r.postID, "err", err)
		return
	}

	r.lastMessage = trimmed
}

func (r *summaryProgressReporter) Complete() {
	r.Update("요약이 완료되었습니다. 바로 아래 메시지를 확인해 주세요.")
}

func (r *summaryProgressReporter) Fail(message string) {
	r.Update(message)
}
