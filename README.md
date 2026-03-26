# Mattermost Echo Summary Plugin

Mattermost Echo Summary is a server + webapp plugin that finds the conversations each user participated in yesterday, expands thread and nearby context, sends that context to a vLLM OpenAI-compatible Chat Completions API, and delivers the resulting summary by DM on scheduled time slots.

Documentation:
- [Korean README](./README.ko.md)
- [Detailed configuration guide](./docs/configuration.ko.md)
- [Architecture guide](./docs/architecture.ko.md)
- [Operations and troubleshooting guide](./docs/operations.ko.md)

## What it does

- Collects the previous day's conversations from the user's authored posts
- Expands thread context, plus nearby channel context for standalone posts
- Optionally includes threads where the user was mentioned
- Chunks large conversation sets to avoid oversized model requests
- Sends the final summary to the user as a bot DM
- Lets each user choose personal delivery times from the Mattermost user settings page

## Admin settings

Configure the plugin from the Mattermost System Console.

- `VLLMBaseURL`: root URL or `/v1` URL of the vLLM OpenAI-compatible endpoint
- `VLLMAPIKey`: optional bearer token
- `VLLMModel`: model name passed to Chat Completions
- `DefaultPrompt`: optional system prompt override
- `NotificationTimezone`: schedule timezone, defaults to `Asia/Seoul`
- `DefaultTimeSlots`: fallback delivery times for users without a personal override
- `TargetUsernames`: optional comma-separated allowlist
- `IncludeMentionedThreads`: include yesterday's mentioned threads
- `MaxThreadsPerUser`: cap for large participation days
- `MaxContextCharacters`: per-request chunk size
- `ContextMessagesBefore` / `ContextMessagesAfter`: nearby context window
- `RequestTimeoutSeconds`: timeout for each model request

## User settings

Each user gets an `Echo Summary` section in User Settings.

- Enable or disable personal delivery
- Save one or more `HH:mm` time slots
- Reset back to the admin default schedule

The same controls are also available from slash commands:

- `/echosummary now`
- `/echosummary status`
- `/echosummary set-times 09:00,13:30`
- `/echosummary disable`
- `/echosummary clear-times`

Notes:
- `/echosummary now` acknowledges immediately, then performs the summary in the background.
- While a manual summary is running, the bot updates a DM progress message so the user can see collection and multi-step vLLM progress.
- Personal delivery times are stored per Mattermost user preference, so each user can keep a different schedule.

## Documentation

- [README.ko.md](./README.ko.md): Korean overview and quick start
- [docs/configuration.ko.md](./docs/configuration.ko.md): admin settings, personal schedule behavior, slash command examples
- [docs/architecture.ko.md](./docs/architecture.ko.md): collection rules, scheduling model, summarization pipeline, data model
- [docs/operations.ko.md](./docs/operations.ko.md): install, deploy, verify, build, package, release, troubleshooting

## Build

```bash
make
```

That generates the manifest files, builds the server and webapp, and creates the plugin package:

```text
dist/com.mattermost.echosummary-<version>.tar.gz
```

For Windows-specific manual build and packaging steps, see [docs/operations.ko.md](./docs/operations.ko.md).

## Development

Install webapp dependencies once:

```bash
cd webapp
npm ci
```

Run tests:

```bash
go test ./...
cd webapp && npm test
```

Build the webapp only after generating the manifest:

```bash
go run build/manifest/main.go apply
cd webapp && npm run build
```

## Delivery flow

1. The background job wakes up every minute.
2. It resolves the current time in the configured timezone.
3. It finds users whose personal or default slots are due.
4. For each user, it gathers yesterday's participation anchors.
5. It expands context and sends chunked prompts to vLLM Chat Completions.
6. It merges partial summaries if needed.
7. It sends the report as a DM from the Echo Summary bot.
