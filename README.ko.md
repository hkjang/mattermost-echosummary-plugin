# Mattermost Echo Summary Plugin

Mattermost Echo Summary는 사용자가 전날 참여한 Mattermost 대화를 수집하고, thread 및 주변 문맥을 확장한 뒤, vLLM OpenAI 호환 Chat Completions API로 요약하여 지정된 시간에 DM으로 전달하는 Mattermost 플러그인입니다.

## 한눈에 보기

- 전날 사용자가 직접 참여한 대화를 기준으로 요약 대상을 찾습니다.
- thread 문맥과 주변 메시지를 함께 모아 단순 채널 덤프보다 정확한 요약을 만듭니다.
- 문맥이 길면 여러 번의 vLLM 호출로 나눠 요약한 뒤 최종 결과를 다시 합칩니다.
- 요약 결과는 봇 계정이 사용자 DM으로 전달합니다.
- 각 사용자는 User Settings 또는 슬래시 커맨드로 자신만의 발송 시간을 따로 저장할 수 있습니다.
- `/echosummary now`는 즉시 응답한 뒤 백그라운드에서 처리되며, DM에서 진행 상태를 업데이트합니다.

## 주요 기능

- 전날 참여 대화 자동 수집
- 멘션된 thread 선택 수집
- 사용자별 발송 시간 설정
- 관리자 기본 시간 슬롯 지원
- vLLM URL, 모델, 프롬프트, 타임아웃 설정
- 스케줄 기반 자동 발송
- 수동 요약 실행 및 진행 상태 표시

## 빠른 사용법

1. Mattermost에 플러그인을 설치하고 활성화합니다.
2. System Console에서 `VLLMBaseURL`, `VLLMModel`, 기본 시간대, 기본 발송 시간을 설정합니다.
3. 사용자는 `User Settings > Echo Summary`에서 개인 시간을 저장하거나, 아래 슬래시 커맨드를 사용합니다.

슬래시 커맨드:

- `/echosummary now`
- `/echosummary status`
- `/echosummary set-times 09:00`
- `/echosummary set-times 09:00,13:30`
- `/echosummary disable`
- `/echosummary clear-times`

## 사용자별 시간 지정

가능합니다. 이 플러그인은 각 Mattermost 사용자 preference에 발송 시간을 별도로 저장합니다.

- 사용자 A: `09:00`
- 사용자 B: `10:30,17:00`
- 사용자 C: 개인 설정 없음 -> 관리자 기본 시간 사용

즉, 한 서버 안에서도 사용자마다 서로 다른 시간표를 가질 수 있습니다.

## 문서 안내

- [docs/configuration.ko.md](./docs/configuration.ko.md)
설정값, 시간 슬롯 규칙, 슬래시 커맨드 동작, 사용자별 저장 방식 설명

- [docs/architecture.ko.md](./docs/architecture.ko.md)
수집 기준, 요약 파이프라인, 스케줄링, DM 진행 메시지, 내부 데이터 모델 설명

- [docs/operations.ko.md](./docs/operations.ko.md)
설치, 배포, 운영 점검, Windows 수동 빌드, 패키징, 릴리즈, 트러블슈팅 설명

## 빌드

일반적인 환경에서는 다음으로 빌드할 수 있습니다.

```bash
make
```

생성 산출물:

```text
dist/com.mattermost.echosummary-<version>.tar.gz
```

Windows에서 `make` 대신 PowerShell로 수동 빌드가 필요한 경우는 [docs/operations.ko.md](./docs/operations.ko.md)를 참고하세요.

## 운영 팁

- `/echosummary now`는 즉시 완료형 명령이 아니라 비동기 실행입니다.
- 수동 요약 중에는 봇 DM 한 개가 진행 메시지로 계속 갱신됩니다.
- 자동 스케줄은 설정된 타임존 기준으로 계산됩니다.
- 대화가 많으면 여러 번의 vLLM 요청이 발생할 수 있습니다.

## 릴리즈

최신 배포 아티팩트는 GitHub Releases에서 받을 수 있습니다.

- [Releases](https://github.com/hkjang/mattermost-echosummary-plugin/releases)
