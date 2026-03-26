# Echo Summary 설정 가이드

## 개요

이 문서는 Echo Summary 플러그인의 관리자 설정, 사용자 설정, 슬래시 커맨드, 시간 슬롯 규칙을 자세히 설명합니다.

## 관리자 설정

Mattermost System Console에서 다음 항목을 설정합니다.

| 설정 키 | 필수 | 예시 | 설명 |
| --- | --- | --- | --- |
| `VLLMBaseURL` | 예 | `https://vllm.example.com/v1` | vLLM OpenAI 호환 엔드포인트의 루트 URL 또는 `/v1` 경로 |
| `VLLMAPIKey` | 아니오 | `sk-...` | Bearer 토큰이 필요한 경우 사용 |
| `VLLMModel` | 예 | `qwen2.5-14b-instruct` | Chat Completions 호출 시 전달할 모델명 |
| `DefaultPrompt` | 아니오 | 사용자 정의 프롬프트 | 비워두면 기본 프롬프트 사용 |
| `NotificationTimezone` | 아니오 | `Asia/Seoul` | 전날 기준일과 스케줄 계산에 사용되는 타임존 |
| `DefaultTimeSlots` | 아니오 | `09:00,13:30` | 사용자 개인 설정이 없을 때 사용할 기본 시간 |
| `TargetUsernames` | 아니오 | `alice,bob` | 비워두면 모든 활성 사용자, 값을 넣으면 username allowlist만 대상 |
| `IncludeMentionedThreads` | 아니오 | `true` | 전날 사용자가 멘션된 thread를 추가 수집할지 여부 |
| `MaxThreadsPerUser` | 아니오 | `40` | 하루에 너무 많은 대화가 있을 때 최신순으로 제한할 최대 스레드 수 |
| `MaxContextCharacters` | 아니오 | `12000` | 한 번의 vLLM 요청에 넣을 최대 문맥 길이 |
| `ContextMessagesBefore` | 아니오 | `3` | anchor 메시지 이전에 붙일 문맥 메시지 수 |
| `ContextMessagesAfter` | 아니오 | `6` | anchor 메시지 이후에 붙일 문맥 메시지 수 |
| `RequestTimeoutSeconds` | 아니오 | `60` | vLLM 요청 타임아웃 |

## 관리자 설정 팁

- `VLLMBaseURL`은 `/v1`까지 넣어도 되고, 루트 URL만 넣어도 됩니다.
- `NotificationTimezone`이 잘못되면 스케줄 계산이 어긋날 수 있습니다.
- `MaxContextCharacters`를 너무 크게 잡으면 모델 응답 지연이 길어질 수 있습니다.
- `TargetUsernames`는 초기 운영이나 제한 배포 시 유용합니다.

## 사용자 설정

사용자는 `User Settings > Echo Summary`에서 다음 동작을 할 수 있습니다.

- 개인 발송 활성화/비활성화
- 하나 이상의 `HH:mm` 시간 저장
- 관리자 기본값으로 되돌리기

### 사용자별 시간 지정 가능 여부

가능합니다. 각 사용자 설정은 Mattermost preference에 사용자별로 저장되므로 서로 독립적입니다.

예시:

- 사용자 A: `09:00`
- 사용자 B: `09:30,18:00`
- 사용자 C: 개인 설정 없음 -> 관리자 기본값 사용

즉, 한 조직 안에서도 사람마다 다른 시간으로 요약을 받을 수 있습니다.

## 시간 슬롯 규칙

- 형식은 반드시 `HH:mm`입니다.
- 예: `09:00`, `13:30`
- 쉼표, 공백, 세미콜론으로 여러 시간을 입력할 수 있습니다.
- 중복 값은 자동 제거됩니다.
- 내부 저장 시 정렬되어 저장됩니다.

예시 입력과 저장 결과:

- `09:00` -> `09:00`
- `13:30,09:00` -> `09:00,13:30`
- `09:00 13:30` -> `09:00,13:30`

## 슬래시 커맨드

### `/echosummary now`

- 즉시 응답을 반환합니다.
- 실제 요약은 백그라운드에서 수행됩니다.
- 봇 DM에서 진행 상태가 갱신됩니다.
- 최종 요약도 DM으로 도착합니다.

### `/echosummary status`

- 현재 vLLM 설정 여부
- 적용 중인 타임존
- 개인 설정인지 관리자 기본값인지
- 현재 발송 시간
- 다음 발송 예정 시각

를 확인할 수 있습니다.

### `/echosummary set-times 09:00`

- 현재 사용자에게 `09:00` 발송 시간을 저장합니다.
- 성공 시 `개인 발송 시간이 저장되었습니다.`를 반환합니다.

### `/echosummary set-times 09:00,13:30`

- 현재 사용자에게 두 개 이상의 발송 시간을 저장합니다.

### `/echosummary disable`

- 현재 사용자에 대해 개인 발송을 비활성화합니다.

### `/echosummary clear-times`

- 현재 사용자의 개인 설정을 지우고 관리자 기본 발송 시간으로 되돌립니다.

## 설정 우선순위

우선순위는 다음과 같습니다.

1. 사용자가 `disable` 상태로 저장한 경우
2. 사용자가 직접 저장한 개인 시간
3. 관리자 기본 시간

## 권장 초기 설정

작게 시작하려면 다음을 권장합니다.

- `NotificationTimezone = Asia/Seoul`
- `DefaultTimeSlots = 09:00`
- `MaxThreadsPerUser = 20`
- `MaxContextCharacters = 8000`
- `TargetUsernames =` 일부 사용자만 지정

운영 안정화 후에 대상 사용자와 문맥 길이를 점진적으로 늘리면 됩니다.
