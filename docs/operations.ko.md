# Echo Summary 운영 및 배포 가이드

## 설치 방법

가장 쉬운 방법은 GitHub Releases의 배포 아티팩트를 사용하는 것입니다.

1. [GitHub Releases](https://github.com/hkjang/mattermost-echosummary-plugin/releases)에서 최신 `com.mattermost.echosummary-*.tar.gz` 파일을 받습니다.
2. Mattermost System Console 또는 Plugin 관리 화면에서 업로드합니다.
3. 플러그인을 활성화합니다.
4. 관리자 설정을 채운 뒤 테스트합니다.

## 초기 운영 점검

설치 후 다음 순서로 점검하는 것을 권장합니다.

1. `VLLMBaseURL`과 `VLLMModel` 저장
2. 기본 타임존과 기본 발송 시간 저장
3. 관리자 본인 계정에서 `/echosummary status` 실행
4. `/echosummary set-times 09:00` 실행
5. `/echosummary now` 실행
6. DM에서 진행 메시지와 최종 요약 수신 확인

## 정상 동작 확인 포인트

### `/echosummary status`

다음 정보를 확인합니다.

- vLLM 설정 완료 여부
- 적용 중인 타임존
- 사용자 개인 설정인지 기본값인지
- 현재 발송 시간
- 다음 발송 예정 시각

### `/echosummary set-times 09:00`

정상이라면:

- 슬래시 커맨드가 바로 응답
- `개인 발송 시간이 저장되었습니다.` 표시
- 이후 `/echosummary status`에서 발송 시간이 `09:00`으로 보임

### `/echosummary now`

정상이라면:

- 슬래시 커맨드가 즉시 응답
- DM에 진행 상태 메시지가 생성됨
- 최종 요약이 DM으로 도착함

## 사용자별 시간 운영 모델

이 플러그인은 사용자별 시간을 지원합니다.

예시 운영:

- 관리자 기본 시간: `09:00`
- 개발팀 A: 각자 `08:30`, `09:00`, `10:00`
- 운영팀 B: `09:00`, `17:30`
- 임원/리더: 개인 설정 없이 기본값 사용

개인 설정이 있으면 관리자 기본값보다 우선합니다.

## Linux/macOS 빌드

```bash
make
```

생성 산출물:

```text
dist/com.mattermost.echosummary-<version>.tar.gz
```

## Windows 수동 빌드

일부 Windows 환경에서는 `make`가 내부 `sh.exe` 제약 때문에 안정적으로 동작하지 않을 수 있습니다. 그런 경우에는 PowerShell에서 단계를 나눠 실행합니다.

핵심 순서:

1. manifest 생성
2. 서버 바이너리 다중 타깃 빌드
3. webapp 빌드
4. `build/package`로 tar.gz 생성

필수 도구:

- Go
- Node.js / npm
- PowerShell

## 패키징 검증

릴리즈 전에 다음을 확인합니다.

- `plugin.json` 버전이 올바른지
- tar.gz 안의 `server/dist/plugin-linux-amd64` 등이 실행 비트를 가지는지
- webapp `main.js`가 포함됐는지
- `plugin.json`이 아카이브 루트 아래 포함됐는지

확인 예시:

```powershell
tar -tvf dist\com.mattermost.echosummary-<version>.tar.gz
```

## 릴리즈 절차

권장 순서:

1. 문서/코드 수정
2. `go test ./...`
3. `cd webapp && npm run lint`
4. `cd webapp && npm run check-types`
5. 패키지 생성
6. Git commit / push
7. GitHub release 생성

## 트러블슈팅

### 슬래시 커맨드가 generic error처럼 보일 때

확인 사항:

- 관리자 설정이 완료됐는지
- user preference 저장 API에서 오류가 나는지
- 서버 로그에 command failure가 남는지

최신 버전에서는 커맨드 핸들러 오류를 generic failure 대신 사용자 메시지로 반환하도록 보완되어 있습니다.

### `/echosummary now`를 눌렀는데 반응이 약할 때

최신 버전에서는:

- 커맨드 ACK를 즉시 반환
- DM 진행 메시지를 생성
- 긴 vLLM 호출 중에도 진행 상태를 갱신

따라서 사용자는 엔터가 눌렸는지 바로 확인할 수 있어야 합니다.

### `set-times`가 저장되지 않는 것처럼 보일 때

확인 순서:

1. `/echosummary set-times 09:00`
2. `/echosummary status`
3. `발송 시간: 09:00` 표시 여부 확인

최신 버전에서는 typed-nil 문제를 수정해 정상 저장 시 오류처럼 보이지 않도록 처리했습니다.

### 자동 발송이 안 올 때

확인 사항:

- `NotificationTimezone`
- `DefaultTimeSlots` 또는 사용자 개인 시간
- 대상 사용자 제한 설정 여부
- 봇 계정 생성 여부
- 전날 실제 참여 대화가 있었는지

### 요약이 너무 늦을 때

조정 포인트:

- `MaxThreadsPerUser` 축소
- `MaxContextCharacters` 축소
- `ContextMessagesBefore/After` 축소
- vLLM 모델/서버 성능 점검

## 운영 권장 사항

- 초기에는 소수 사용자만 대상으로 시작
- 대화량이 많은 조직은 제한값을 보수적으로 시작
- 릴리즈 아티팩트를 우선 사용
- 긴 요약은 `/echosummary now` 수동 실행으로 먼저 검증
