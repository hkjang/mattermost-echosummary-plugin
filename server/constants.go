package main

const (
	pluginID   = "com.mattermost.echosummary"
	pluginName = "Echo Summary"

	botUsername    = "echo.summary.bot"
	botDisplayName = "Echo Summary"
	botDescription = "Summarizes yesterday's Mattermost conversations and delivers them on a schedule."

	commandTrigger = "echosummary"

	userPreferenceCategory      = "pp_" + pluginID
	userPreferenceDeliveryTimes = "delivery_times"
	userPreferenceDisabledValue = "off"

	backgroundJobKey = "EchoSummaryScheduler"
	sentStatePrefix  = "summary-sent:"

	defaultNotificationTimezone  = "Asia/Seoul"
	defaultUserTimeSlots         = "09:00"
	defaultContextMessagesBefore = 3
	defaultContextMessagesAfter  = 6
	defaultMaxContextCharacters  = 12000
	defaultMaxThreadsPerUser     = 40
	defaultRequestTimeoutSeconds = 60
)

const defaultSummaryPrompt = `당신은 Mattermost 업무 대화를 요약하는 비서다.
중복을 제거하고, 사용자의 관점에서 중요한 내용만 한국어로 정리하라.
반드시 아래 섹션을 포함하라.
1. 한눈에 보기
2. 핵심 결정
3. 할 일
4. 미해결 이슈
5. 채널별 요약

각 섹션에서 정보가 없으면 "없음"이라고 명시하라.
추측은 하지 말고, 대화에 나온 사실만 바탕으로 간결하게 요약하라.`

const finalMergePrompt = `당신은 Mattermost 부분 요약들을 하나의 최종 업무 리포트로 합치는 비서다.
부분 요약 사이의 중복을 제거하고, 중요한 실행 항목과 결정 사항을 우선순위 있게 정리하라.
출력은 한국어로 작성하고, 최종 형식은 다음 섹션을 유지하라.
1. 한눈에 보기
2. 핵심 결정
3. 할 일
4. 미해결 이슈
5. 채널별 요약`
