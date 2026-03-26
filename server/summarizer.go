package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/pkg/errors"
)

type threadAnchor struct {
	ContextID      string
	ChannelID      string
	AnchorPosts    map[string]*model.Post
	LatestAnchorAt int64
	IsThread       bool
}

type conversationContext struct {
	ContextID        string
	ChannelLabel     string
	TeamLabel        string
	LatestActivityAt int64
	Text             string
}

type lookupCaches struct {
	channels map[string]*model.Channel
	teams    map[string]*model.Team
	users    map[string]*model.User
}

type summaryProgressFunc func(string)

func (p *Plugin) ensureBot() error {
	botUserID, err := p.client.Bot.EnsureBot(&model.Bot{
		Username:    botUsername,
		DisplayName: botDisplayName,
		Description: botDescription,
	})
	if err != nil {
		return err
	}
	p.botUserID = botUserID
	return nil
}

func (p *Plugin) listEligibleUsers(cfg *configuration) ([]*model.User, error) {
	if targetUsernames := parseCommaSeparated(cfg.TargetUsernames); len(targetUsernames) > 0 {
		users, err := p.client.User.ListByUsernames(targetUsernames)
		if err != nil {
			return nil, err
		}
		filtered := make([]*model.User, 0, len(users))
		for _, user := range users {
			if user == nil || user.IsBot || user.DeleteAt != 0 {
				continue
			}
			filtered = append(filtered, user)
		}
		return filtered, nil
	}

	users := make([]*model.User, 0, 256)
	for page := 0; ; page++ {
		batch, err := p.client.User.List(&model.UserGetOptions{
			Active:  true,
			Page:    page,
			PerPage: 200,
		})
		if err != nil {
			return nil, err
		}
		for _, user := range batch {
			if user == nil || user.IsBot || user.DeleteAt != 0 {
				continue
			}
			users = append(users, user)
		}
		if len(batch) < 200 {
			break
		}
	}

	return users, nil
}

func parseCommaSeparated(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	values := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		values = append(values, part)
	}
	return values
}

func (p *Plugin) sendSummaryToUser(user *model.User, referenceTime time.Time, cfg *configuration) error {
	return p.sendSummaryToUserWithProgress(user, referenceTime, cfg, nil)
}

func (p *Plugin) sendSummaryToUserWithProgress(user *model.User, referenceTime time.Time, cfg *configuration, progress summaryProgressFunc) error {
	if p.botUserID == "" {
		if err := p.ensureBot(); err != nil {
			return errors.Wrap(err, "failed to ensure bot")
		}
	}

	message, err := p.generateSummaryMessageWithProgress(user, referenceTime, cfg, progress)
	if err != nil {
		return err
	}

	if progress != nil {
		progress("최종 요약 DM을 전송하고 있습니다...")
	}

	return p.client.Post.DM(p.botUserID, user.Id, &model.Post{Message: message})
}

func (p *Plugin) generateSummaryMessage(user *model.User, referenceTime time.Time, cfg *configuration) (string, error) {
	return p.generateSummaryMessageWithProgress(user, referenceTime, cfg, nil)
}

func (p *Plugin) generateSummaryMessageWithProgress(user *model.User, referenceTime time.Time, cfg *configuration, progress summaryProgressFunc) (string, error) {
	location, err := loadScheduleLocation(cfg.NotificationTimezone)
	if err != nil {
		return "", err
	}

	summaryDate := referenceTime.In(location).AddDate(0, 0, -1)
	if progress != nil {
		progress("어제 참여한 대화를 수집하고 있습니다...")
	}

	contexts, channelCount, truncated, err := p.collectConversationContexts(user, summaryDate, cfg)
	if err != nil {
		return "", err
	}

	header := []string{
		fmt.Sprintf("## 전날 대화 요약 (%s)", summaryDate.Format("2006-01-02")),
		"",
		fmt.Sprintf("- 참여 스레드: %d개", len(contexts)),
		fmt.Sprintf("- 채널: %d개", channelCount),
	}
	if truncated {
		header = append(header, "- 오래된 참여 대화 일부는 최대 스레드 수 제한으로 생략되었습니다.")
	}

	if len(contexts) == 0 {
		if progress != nil {
			progress("어제 참여한 대화를 찾지 못했습니다. 결과 안내를 DM으로 전송합니다...")
		}
		return strings.Join(append(header,
			"",
			"어제 참여한 대화를 찾지 못해 요약을 건너뛰었습니다.",
			"",
			fmt.Sprintf("_생성 시각: %s_", referenceTime.In(location).Format("2006-01-02 15:04 MST")),
		), "\n"), nil
	}

	chunks := chunkContexts(contexts, cfg.MaxContextCharacters)
	if progress != nil {
		if len(chunks) == 1 {
			progress(fmt.Sprintf("%d개의 대화 문맥을 요약 모델에 요청하고 있습니다...", len(contexts)))
		} else {
			progress(fmt.Sprintf("%d개의 대화 문맥을 %d개의 요청으로 나눠 순차 요약하고 있습니다...", len(contexts), len(chunks)))
		}
	}

	partials := make([]string, 0, len(chunks))
	for index, chunk := range chunks {
		if progress != nil && len(chunks) > 1 {
			progress(fmt.Sprintf("부분 요약 %d/%d 진행 중입니다...", index+1, len(chunks)))
		}

		summary, err := p.createChatCompletion(cfg, []chatMessage{
			{Role: "system", Content: cfg.DefaultPrompt},
			{Role: "user", Content: buildSummaryUserPrompt(user, summaryDate, location, chunk)},
		})
		if err != nil {
			return "", err
		}
		partials = append(partials, summary)
	}

	finalSummary := partials[0]
	if len(partials) > 1 {
		if progress != nil {
			progress("부분 요약을 합쳐 최종 결과를 정리하고 있습니다...")
		}

		finalSummary, err = p.createChatCompletion(cfg, []chatMessage{
			{Role: "system", Content: finalMergePrompt},
			{Role: "user", Content: buildMergePrompt(user, summaryDate, partials)},
		})
		if err != nil {
			return "", err
		}
	}

	if progress != nil {
		progress("최종 요약 메시지를 정리하고 있습니다...")
	}

	messageLines := append(header, "", strings.TrimSpace(finalSummary), "", fmt.Sprintf("_생성 시각: %s_", referenceTime.In(location).Format("2006-01-02 15:04 MST")))
	return strings.Join(messageLines, "\n"), nil
}

func buildSummaryUserPrompt(user *model.User, summaryDate time.Time, location *time.Location, contexts []conversationContext) string {
	parts := []string{
		fmt.Sprintf("대상 사용자: @%s", user.Username),
		fmt.Sprintf("기준일: %s (%s)", summaryDate.Format("2006-01-02"), location.String()),
		"",
		"아래는 사용자가 어제 참여했던 Mattermost 대화 문맥이다.",
		"사용자 관점에서 중요한 내용만 요약하고, 반복을 제거하라.",
	}

	for _, context := range contexts {
		parts = append(parts, "", context.Text)
	}

	return strings.Join(parts, "\n")
}

func buildMergePrompt(user *model.User, summaryDate time.Time, partials []string) string {
	parts := []string{
		fmt.Sprintf("대상 사용자: @%s", user.Username),
		fmt.Sprintf("기준일: %s", summaryDate.Format("2006-01-02")),
		"",
		"아래 부분 요약들을 하나의 최종 리포트로 합쳐라.",
	}

	for index, partial := range partials {
		parts = append(parts, "", fmt.Sprintf("### 부분 요약 %d", index+1), partial)
	}

	return strings.Join(parts, "\n")
}

func chunkContexts(contexts []conversationContext, maxCharacters int) [][]conversationContext {
	if len(contexts) == 0 {
		return nil
	}

	chunks := make([][]conversationContext, 0)
	currentChunk := make([]conversationContext, 0)
	currentLength := 0

	for _, context := range contexts {
		contextLength := len(context.Text)
		if len(currentChunk) > 0 && currentLength+contextLength > maxCharacters {
			chunks = append(chunks, currentChunk)
			currentChunk = make([]conversationContext, 0)
			currentLength = 0
		}

		currentChunk = append(currentChunk, context)
		currentLength += contextLength
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

func (p *Plugin) collectConversationContexts(user *model.User, summaryDate time.Time, cfg *configuration) ([]conversationContext, int, bool, error) {
	location, err := loadScheduleLocation(cfg.NotificationTimezone)
	if err != nil {
		return nil, 0, false, err
	}

	anchors, err := p.collectAnchors(user, summaryDate, cfg, location)
	if err != nil {
		return nil, 0, false, err
	}

	truncated := false
	sort.Slice(anchors, func(i, j int) bool {
		return anchors[i].LatestAnchorAt > anchors[j].LatestAnchorAt
	})
	if len(anchors) > cfg.MaxThreadsPerUser {
		anchors = anchors[:cfg.MaxThreadsPerUser]
		truncated = true
	}

	caches := &lookupCaches{
		channels: map[string]*model.Channel{},
		teams:    map[string]*model.Team{},
		users:    map[string]*model.User{},
	}

	contexts := make([]conversationContext, 0, len(anchors))
	channelIDs := map[string]struct{}{}
	for _, anchor := range anchors {
		context, err := p.buildConversationContext(anchor, user.Id, cfg, caches)
		if err != nil {
			p.API.LogWarn("Failed to build conversation context", "context_id", anchor.ContextID, "err", err)
			continue
		}
		if context.Text == "" {
			continue
		}
		channelIDs[anchor.ChannelID] = struct{}{}
		contexts = append(contexts, context)
	}

	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].LatestActivityAt > contexts[j].LatestActivityAt
	})

	return contexts, len(channelIDs), truncated, nil
}

func (p *Plugin) collectAnchors(user *model.User, summaryDate time.Time, cfg *configuration, location *time.Location) ([]*threadAnchor, error) {
	anchors := map[string]*threadAnchor{}
	addAnchor := func(post *model.Post) {
		if post == nil || post.DeleteAt != 0 || post.IsSystemMessage() {
			return
		}

		contextID := post.Id
		isThread := false
		if post.RootId != "" {
			contextID = post.RootId
			isThread = true
		}

		existing, ok := anchors[contextID]
		if !ok {
			existing = &threadAnchor{
				ContextID:   contextID,
				ChannelID:   post.ChannelId,
				AnchorPosts: map[string]*model.Post{},
				IsThread:    isThread,
			}
			anchors[contextID] = existing
		}
		existing.AnchorPosts[post.Id] = post
		if post.CreateAt > existing.LatestAnchorAt {
			existing.LatestAnchorAt = post.CreateAt
		}
		if post.RootId != "" {
			existing.IsThread = true
		}
	}

	teams, err := p.client.Team.List(pluginapi.FilterTeamsByUser(user.Id))
	if err != nil {
		return nil, err
	}

	dateString := summaryDate.Format("2006-01-02")
	_, offset := summaryDate.In(location).Zone()
	for _, team := range teams {
		posts, err := p.client.Post.SearchPostsInTeam(team.Id, []*model.SearchParams{{
			FromUsers:      []string{user.Username},
			OnDate:         dateString,
			TimeZoneOffset: offset,
		}})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to search posts for team %s", team.Id)
		}
		for _, post := range posts {
			addAnchor(post)
		}

		if cfg.IncludeMentionedThreads {
			mentioned, err := p.client.Post.SearchPostsInTeam(team.Id, []*model.SearchParams{{
				Terms:          "@" + user.Username,
				OnDate:         dateString,
				TimeZoneOffset: offset,
			}})
			if err != nil {
				return nil, errors.Wrapf(err, "failed to search mentioned posts for team %s", team.Id)
			}
			for _, post := range mentioned {
				if post.UserId == user.Id {
					continue
				}
				addAnchor(post)
			}
		}
	}

	startOfDay := time.Date(summaryDate.Year(), summaryDate.Month(), summaryDate.Day(), 0, 0, 0, 0, location)
	endOfDay := startOfDay.Add(24 * time.Hour)
	startMillis := model.GetMillisForTime(startOfDay)
	endMillis := model.GetMillisForTime(endOfDay)

	seenChannels := map[string]struct{}{}
	for _, team := range teams {
		channels, err := p.client.Channel.ListForTeamForUser(team.Id, user.Id, false)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list channels for team %s", team.Id)
		}
		for _, channel := range channels {
			if channel == nil || channel.Type != model.ChannelTypeDirect && channel.Type != model.ChannelTypeGroup {
				continue
			}
			if _, ok := seenChannels[channel.Id]; ok {
				continue
			}
			seenChannels[channel.Id] = struct{}{}

			postList, err := p.client.Post.GetPostsSince(channel.Id, startMillis)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to load direct/group posts for channel %s", channel.Id)
			}
			for _, post := range postList.ToSlice() {
				if post.CreateAt >= endMillis || post.CreateAt < startMillis {
					continue
				}
				if post.UserId == user.Id {
					addAnchor(post)
				}
			}
		}
	}

	result := make([]*threadAnchor, 0, len(anchors))
	for _, anchor := range anchors {
		result = append(result, anchor)
	}
	return result, nil
}

func (p *Plugin) buildConversationContext(anchor *threadAnchor, userID string, cfg *configuration, caches *lookupCaches) (conversationContext, error) {
	channel, err := p.getChannel(anchor.ChannelID, caches)
	if err != nil {
		return conversationContext{}, err
	}

	posts, err := p.loadContextPosts(anchor, cfg)
	if err != nil {
		return conversationContext{}, err
	}
	if len(posts) == 0 {
		return conversationContext{}, nil
	}

	teamLabel := ""
	if channel.TeamId != "" {
		team, err := p.getTeam(channel.TeamId, caches)
		if err == nil && team != nil {
			teamLabel = team.DisplayName
			if teamLabel == "" {
				teamLabel = team.Name
			}
		}
	}

	lines := []string{
		"### 대화 문맥",
		fmt.Sprintf("- 채널: %s", describeChannel(channel)),
	}
	if teamLabel != "" {
		lines = append(lines, fmt.Sprintf("- 팀: %s", teamLabel))
	}

	messageLines := make([]string, 0, len(posts))
	latestActivity := int64(0)
	for _, post := range posts {
		if post.CreateAt > latestActivity {
			latestActivity = post.CreateAt
		}
		author, err := p.getUser(post.UserId, caches)
		if err != nil {
			return conversationContext{}, err
		}

		authorLabel := "unknown"
		if author != nil && author.Username != "" {
			authorLabel = "@" + author.Username
		}
		if post.UserId == userID {
			authorLabel += " (본인)"
		}

		body := sanitizePostMessage(post)
		if body == "" {
			continue
		}
		messageLines = append(messageLines, fmt.Sprintf("- [%s] %s: %s", model.GetTimeForMillis(post.CreateAt).Format("15:04"), authorLabel, body))
	}

	if len(messageLines) == 0 {
		return conversationContext{}, nil
	}

	lines = append(lines, "- 메시지:")
	lines = append(lines, messageLines...)

	return conversationContext{
		ContextID:        anchor.ContextID,
		ChannelLabel:     describeChannel(channel),
		TeamLabel:        teamLabel,
		LatestActivityAt: latestActivity,
		Text:             strings.Join(lines, "\n"),
	}, nil
}

func (p *Plugin) loadContextPosts(anchor *threadAnchor, cfg *configuration) ([]*model.Post, error) {
	threadPosts, err := p.client.Post.GetPostThread(anchor.ContextID)
	if err == nil && threadPosts != nil && len(threadPosts.Order) > 1 {
		return selectThreadWindow(threadPosts.ToSlice(), anchor, cfg), nil
	}

	post, err := p.client.Post.GetPost(anchor.ContextID)
	if err != nil {
		return nil, err
	}

	combined := model.NewPostList()
	before, err := p.client.Post.GetPostsBefore(post.ChannelId, post.Id, 0, cfg.ContextMessagesBefore)
	if err == nil && before != nil {
		combined.Extend(before)
	}
	combined.AddPost(post)
	combined.AddOrder(post.Id)
	after, err := p.client.Post.GetPostsAfter(post.ChannelId, post.Id, 0, cfg.ContextMessagesAfter)
	if err == nil && after != nil {
		combined.Extend(after)
	}

	return sortAndFilterPosts(combined.ToSlice()), nil
}

func selectThreadWindow(posts []*model.Post, anchor *threadAnchor, cfg *configuration) []*model.Post {
	ordered := sortAndFilterPosts(posts)
	if len(ordered) == 0 {
		return nil
	}

	selected := map[string]struct{}{ordered[0].Id: {}}
	for index, post := range ordered {
		if _, ok := anchor.AnchorPosts[post.Id]; !ok {
			continue
		}

		start := index - cfg.ContextMessagesBefore
		if start < 0 {
			start = 0
		}
		end := index + cfg.ContextMessagesAfter
		if end >= len(ordered) {
			end = len(ordered) - 1
		}
		for i := start; i <= end; i++ {
			selected[ordered[i].Id] = struct{}{}
		}
	}

	window := make([]*model.Post, 0, len(selected))
	for _, post := range ordered {
		if _, ok := selected[post.Id]; ok {
			window = append(window, post)
		}
	}
	return window
}

func sortAndFilterPosts(posts []*model.Post) []*model.Post {
	filtered := make([]*model.Post, 0, len(posts))
	for _, post := range posts {
		if post == nil || post.DeleteAt != 0 || post.IsSystemMessage() {
			continue
		}
		filtered = append(filtered, post)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].CreateAt == filtered[j].CreateAt {
			return filtered[i].Id < filtered[j].Id
		}
		return filtered[i].CreateAt < filtered[j].CreateAt
	})

	return filtered
}

func sanitizePostMessage(post *model.Post) string {
	parts := make([]string, 0, 2)
	if message := strings.TrimSpace(post.Message); message != "" {
		message = strings.Join(strings.Fields(message), " ")
		if len(message) > 500 {
			message = message[:500] + "..."
		}
		parts = append(parts, message)
	}
	if len(post.FileIds) > 0 {
		parts = append(parts, fmt.Sprintf("[첨부 %d개]", len(post.FileIds)))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func describeChannel(channel *model.Channel) string {
	switch channel.Type {
	case model.ChannelTypeOpen, model.ChannelTypePrivate:
		if channel.Name != "" {
			return "~" + channel.Name
		}
		if channel.DisplayName != "" {
			return channel.DisplayName
		}
	case model.ChannelTypeDirect:
		if channel.DisplayName != "" {
			return "DM (" + channel.DisplayName + ")"
		}
		return "DM"
	case model.ChannelTypeGroup:
		if channel.DisplayName != "" {
			return "GM (" + channel.DisplayName + ")"
		}
		return "GM"
	}
	return channel.DisplayName
}

func (p *Plugin) getChannel(channelID string, caches *lookupCaches) (*model.Channel, error) {
	if channel, ok := caches.channels[channelID]; ok {
		return channel, nil
	}
	channel, err := p.client.Channel.Get(channelID)
	if err != nil {
		return nil, err
	}
	caches.channels[channelID] = channel
	return channel, nil
}

func (p *Plugin) getTeam(teamID string, caches *lookupCaches) (*model.Team, error) {
	if team, ok := caches.teams[teamID]; ok {
		return team, nil
	}
	team, err := p.client.Team.Get(teamID)
	if err != nil {
		return nil, err
	}
	caches.teams[teamID] = team
	return team, nil
}

func (p *Plugin) getUser(userID string, caches *lookupCaches) (*model.User, error) {
	if user, ok := caches.users[userID]; ok {
		return user, nil
	}
	user, err := p.client.User.Get(userID)
	if err != nil {
		return nil, err
	}
	caches.users[userID] = user
	return user, nil
}
