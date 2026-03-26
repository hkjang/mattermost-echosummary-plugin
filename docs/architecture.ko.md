# Echo Summary 아키텍처 가이드

## 목적

Echo Summary는 사용자가 전날 실제로 참여한 대화만 추려서, 그 사람 관점의 업무 요약을 생성하고 DM으로 전달하는 것을 목표로 합니다.

## 구성 요소

### 서버 플러그인

역할:

- 대화 수집
- 스케줄 실행
- vLLM 호출
- 요약 결과 전송
- 슬래시 커맨드 처리

핵심 파일:

- [server/plugin.go](../server/plugin.go)
- [server/job.go](../server/job.go)
- [server/summarizer.go](../server/summarizer.go)
- [server/command_handler.go](../server/command_handler.go)
- [server/manual_summary.go](../server/manual_summary.go)
- [server/preferences.go](../server/preferences.go)
- [server/vllm_client.go](../server/vllm_client.go)

### 웹앱

역할:

- User Settings에 `Echo Summary` 섹션 표시
- 개인 시간 설정 저장/초기화

핵심 파일:

- [webapp/src/index.tsx](../webapp/src/index.tsx)
- [webapp/src/user_settings.tsx](../webapp/src/user_settings.tsx)

## 대화 수집 기준

플러그인은 전날 사용자가 작성한 post를 anchor로 삼습니다.

기본 규칙:

- 사용자가 전날 직접 작성한 post가 있으면 포함
- thread에 참여했으면 해당 thread 문맥을 함께 포함
- thread가 아닌 일반 post는 앞뒤 주변 문맥을 확장
- 필요 시 전날 멘션된 thread도 후보에 포함
- 단순 읽음만으로는 포함하지 않음
- 이모지 반응만 있는 경우는 포함하지 않음

## 수집 흐름

1. 대상 사용자를 결정합니다.
2. 해당 사용자의 팀 목록을 가져옵니다.
3. 팀 단위로 전날 작성한 post를 검색합니다.
4. 설정된 경우 멘션된 thread도 추가 검색합니다.
5. DM/GM 채널에서도 전날 작성한 post를 별도로 찾습니다.
6. root post 기준으로 thread 또는 주변 문맥을 확장합니다.
7. 정렬과 제한을 적용해 최종 문맥 묶음을 만듭니다.

## 요약 파이프라인

1. 수집된 문맥을 채널/팀/시간 정보와 함께 텍스트화합니다.
2. 문맥 길이가 너무 길면 여러 chunk로 나눕니다.
3. chunk마다 vLLM Chat Completions를 호출합니다.
4. 부분 요약이 여러 개라면 다시 merge 프롬프트로 합칩니다.
5. 최종 결과를 DM 메시지로 전송합니다.

## 수동 요약 UX

`/echosummary now` 흐름:

1. 슬래시 커맨드는 즉시 ephemeral 응답을 반환합니다.
2. 백그라운드 goroutine이 실행됩니다.
3. 봇 DM에 진행 상태 메시지를 생성합니다.
4. 수집 단계, 부분 요약 단계, 최종 정리 단계마다 같은 DM 메시지를 갱신합니다.
5. 완료 후 최종 요약 DM을 별도로 전송합니다.

이 구조 덕분에 사용자는 엔터가 먹혔는지 헷갈리지 않고, 긴 vLLM 호출 동안 진행 상태도 볼 수 있습니다.

## 자동 스케줄링

배경 작업은 1분마다 실행됩니다.

동작:

- 설정된 타임존으로 현재 시각 계산
- 사용자별 적용 시간 계산
- 2분 grace window 안에 들어온 슬롯만 실행
- 이미 보낸 슬롯은 KV 상태로 중복 방지

## 사용자별 시간 저장 방식

개인 시간은 Mattermost preference에 저장됩니다.

키 구조:

- `category = pp_com.mattermost.echosummary`
- `name = delivery_times`

값 예시:

- `09:00`
- `09:00,13:30`
- `off`

따라서 사용자별로 완전히 독립적인 시간표를 유지할 수 있습니다.

## 오류 처리 원칙

- 설정 미완료는 사용자 친화적인 ephemeral 메시지로 안내
- 커맨드 핸들러 오류는 generic slash command failure 대신 명시적 오류 메시지 반환
- 수동 요약 실패 시 DM 또는 ephemeral 안내 제공
- 스케줄 실행 실패는 서버 로그에 남기고 다음 사용자 처리 계속 진행

## 패키징

Windows 환경에서도 Mattermost 서버 바이너리 실행 권한이 유지되도록 [build/package/main.go](../build/package/main.go)에서 tar.gz를 직접 생성합니다.

아카이브 안의 서버 바이너리 모드:

- `rwxr-xr-x`

이렇게 해야 Linux 기반 Mattermost 서버에서 플러그인 바이너리 실행 권한 문제를 줄일 수 있습니다.
