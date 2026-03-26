package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

func (p *Plugin) registerCommand() error {
	return p.client.SlashCommand.Register(&model.Command{
		Trigger:          commandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "Summarize yesterday's conversations and manage delivery times",
		AutoCompleteHint: "[now|status|set-times|disable|clear-times]",
		AutocompleteData: buildCommandAutocomplete(),
	})
}

func buildCommandAutocomplete() *model.AutocompleteData {
	root := model.NewAutocompleteData(commandTrigger, "", "Summarize yesterday's conversations and manage delivery times")

	nowCommand := model.NewAutocompleteData("now", "", "Generate yesterday's summary now and send it by DM")
	statusCommand := model.NewAutocompleteData("status", "", "Show current Echo Summary delivery settings")
	setTimesCommand := model.NewAutocompleteData("set-times", "[HH:mm,HH:mm]", "Save one or more personal delivery times")
	setTimesCommand.AddTextArgument("Comma-separated delivery times", "09:00,13:30", `^([0-2]\d:[0-5]\d)(,[0-2]\d:[0-5]\d)*$`)
	disableCommand := model.NewAutocompleteData("disable", "", "Disable personal deliveries")
	clearCommand := model.NewAutocompleteData("clear-times", "", "Remove personal override and fall back to admin defaults")

	root.AddCommand(nowCommand)
	root.AddCommand(statusCommand)
	root.AddCommand(setTimesCommand)
	root.AddCommand(disableCommand)
	root.AddCommand(clearCommand)

	return root
}

func (p *Plugin) handleCommand(args *model.CommandArgs) (*model.CommandResponse, error) {
	fields := strings.Fields(args.Command)
	if len(fields) <= 1 {
		return p.ephemeralCommandResponse(commandHelpText()), nil
	}

	subcommand := strings.ToLower(fields[1])
	cfg := p.getConfiguration().normalized()

	switch subcommand {
	case "now":
		if !cfg.isConfigured() {
			return p.ephemeralCommandResponse("관리자 설정이 아직 완료되지 않았습니다. 시스템 콘솔에서 vLLM URL과 모델명을 먼저 설정해 주세요."), nil
		}

		p.runManualSummary(args.UserId, args.ChannelId, time.Now(), cfg)
		return p.ephemeralCommandResponse("요약 요청을 접수했습니다. DM에서 진행 상태와 결과를 알려드릴게요."), nil

	case "status":
		settings, err := p.getUserDeliverySettings(args.UserId, cfg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load delivery settings")
		}
		location, locErr := loadScheduleLocation(cfg.NotificationTimezone)
		if locErr != nil {
			location = time.FixedZone("KST", 9*60*60)
		}

		lines := []string{
			fmt.Sprintf("vLLM 설정: %s", enabledLabel(cfg.isConfigured())),
			fmt.Sprintf("타임존: %s", location.String()),
			fmt.Sprintf("개인 설정 출처: %s", settings.Source),
		}
		if settings.Disabled {
			lines = append(lines, "개인 알림: 비활성화")
		} else {
			lines = append(lines, fmt.Sprintf("발송 시간: %s", strings.Join(settings.Slots, ", ")))
			lines = append(lines, fmt.Sprintf("다음 발송 예정: %s", nextDeliveryTime(time.Now().In(location), settings.Slots)))
		}
		if strings.TrimSpace(cfg.TargetUsernames) == "" {
			lines = append(lines, "대상 사용자 범위: 모든 활성 사용자")
		} else {
			lines = append(lines, fmt.Sprintf("대상 사용자 범위: %s", cfg.TargetUsernames))
		}
		return p.ephemeralCommandResponse(strings.Join(lines, "\n")), nil

	case "set-times":
		parts := strings.SplitN(args.Command, " ", 3)
		if len(parts) < 3 {
			return p.ephemeralCommandResponse("예시: `/echosummary set-times 09:00,13:30`"), nil
		}
		if err := p.saveUserDeliveryPreference(args.UserId, parts[2]); err != nil {
			return p.ephemeralCommandResponse(fmt.Sprintf("시간 저장에 실패했습니다: %s", err.Error())), nil
		}
		return p.ephemeralCommandResponse("개인 발송 시간이 저장되었습니다."), nil

	case "disable":
		if err := p.disableUserDeliveryPreference(args.UserId); err != nil {
			return nil, errors.Wrap(err, "failed to disable deliveries")
		}
		return p.ephemeralCommandResponse("개인 발송을 비활성화했습니다."), nil

	case "clear-times":
		if err := p.clearUserDeliveryPreference(args.UserId); err != nil {
			return nil, errors.Wrap(err, "failed to clear delivery override")
		}
		return p.ephemeralCommandResponse("개인 설정을 지우고 관리자 기본 시간대로 되돌렸습니다."), nil
	default:
		return p.ephemeralCommandResponse(commandHelpText()), nil
	}
}

func (p *Plugin) ephemeralCommandResponse(text string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         text,
	}
}

func enabledLabel(enabled bool) string {
	if enabled {
		return "완료"
	}
	return "미완료"
}

func commandHelpText() string {
	return strings.Join([]string{
		"`/echosummary now` 현재 시점 기준으로 어제 대화를 즉시 요약해 DM으로 받습니다.",
		"`/echosummary status` 현재 개인 발송 시간과 설정 상태를 확인합니다.",
		"`/echosummary set-times 09:00,13:30` 개인 발송 시간을 저장합니다.",
		"`/echosummary disable` 개인 발송을 끕니다.",
		"`/echosummary clear-times` 개인 설정을 지우고 관리자 기본값으로 돌아갑니다.",
	}, "\n")
}
