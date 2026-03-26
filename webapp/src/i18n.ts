type UserSettingsText = {
    clearToDefault: string;
    description: string;
    enableDelivery: string;
    exampleLabel: string;
    exampleValue: string;
    heading: string;
    inputPlaceholder: string;
    resetError: string;
    resetSuccess: string;
    saveError: string;
    saveIdle: string;
    saveSaving: string;
    saveSuccessDisabled: string;
    saveSuccessEnabled: string;
    sectionTitle: string;
    validateEmpty: string;
};

type AdminSettingText = {
    displayName: string;
    helpText?: string;
    placeholder?: string;
};

type AdminSectionText = {
    title: string;
    settings: Record<string, AdminSettingText>;
};

type AdminConsoleText = {
    header: string;
    sections: Record<string, AdminSectionText>;
};

type EchoSummaryText = {
    userSettings: UserSettingsText;
    adminConsole: AdminConsoleText;
};

const en: EchoSummaryText = {
    userSettings: {
        clearToDefault: 'Reset to workspace default',
        description: 'Receive a DM summary of the conversations you joined yesterday. You can override the admin default times here or disable personal delivery.',
        enableDelivery: 'Enable personal delivery schedule',
        exampleLabel: 'Enter one or more times separated by commas or spaces. Example:',
        exampleValue: '09:00, 13:30',
        heading: 'Previous-day summary DM',
        inputPlaceholder: '09:00,13:30',
        resetError: 'Failed to clear your personal settings.',
        resetSuccess: 'Your personal override was removed and the workspace default will be used.',
        saveError: 'Failed to save the settings.',
        saveIdle: 'Save',
        saveSaving: 'Saving...',
        saveSuccessDisabled: 'Personal delivery has been disabled.',
        saveSuccessEnabled: 'Personal delivery times have been saved.',
        sectionTitle: 'Delivery schedule',
        validateEmpty: 'Enter at least one HH:mm time.',
    },
    adminConsole: {
        header: 'Echo Summary collects each user\'s previous-day Mattermost conversations, summarizes them through a configurable vLLM OpenAI-compatible Chat Completions API, and delivers the report by DM on scheduled times.',
        sections: {
            vllm: {
                title: 'vLLM summary settings',
                settings: {
                    VLLMBaseURL: {
                        displayName: 'vLLM Base URL',
                        helpText: 'Use the root URL or include /v1. The plugin appends /chat/completions automatically.',
                        placeholder: 'https://vllm.example.com/v1',
                    },
                    VLLMAPIKey: {
                        displayName: 'vLLM API key',
                        helpText: 'When provided, the value is sent as an Authorization Bearer token.',
                        placeholder: 'sk-...',
                    },
                    VLLMModel: {
                        displayName: 'Summary model name',
                        helpText: 'Sent as the model field in the Chat Completions request.',
                        placeholder: 'qwen2.5-14b-instruct',
                    },
                    DefaultPrompt: {
                        displayName: 'Default summary prompt',
                        helpText: 'Used as the system prompt. Leave blank to use Echo Summary\'s built-in default prompt.',
                    },
                    RequestTimeoutSeconds: {
                        displayName: 'Request timeout (seconds)',
                        helpText: 'Maximum time to wait for a vLLM response.',
                    },
                },
            },
            scope: {
                title: 'Target users and schedule',
                settings: {
                    NotificationTimezone: {
                        displayName: 'Default timezone',
                        helpText: 'Used when calculating the previous-day window and each user\'s delivery schedule.',
                        placeholder: 'Asia/Seoul',
                    },
                    DefaultTimeSlots: {
                        displayName: 'Default delivery times',
                        helpText: 'Fallback times used when a user has not set a personal override. You can enter multiple times separated by commas.',
                        placeholder: '09:00,13:30',
                    },
                    TargetUsernames: {
                        displayName: 'Target usernames',
                        helpText: 'Leave blank to include all active users. If set, only the listed usernames are summarized.',
                        placeholder: 'alice,bob',
                    },
                    IncludeMentionedThreads: {
                        displayName: 'Include mentioned threads',
                        helpText: 'Also include threads where the user was mentioned yesterday, even if they did not post in them.',
                    },
                },
            },
            collection: {
                title: 'Collection scope',
                settings: {
                    MaxThreadsPerUser: {
                        displayName: 'Max threads per user',
                        helpText: 'When a user joined many conversations, only the most recent threads up to this limit are summarized.',
                    },
                    MaxContextCharacters: {
                        displayName: 'Max characters per request',
                        helpText: 'Character budget used when batching collected context into Chat Completions requests.',
                    },
                    ContextMessagesBefore: {
                        displayName: 'Messages before anchor',
                        helpText: 'Maximum number of surrounding messages to include before each user-authored anchor post.',
                    },
                    ContextMessagesAfter: {
                        displayName: 'Messages after anchor',
                        helpText: 'Maximum number of surrounding messages to include after each user-authored anchor post.',
                    },
                },
            },
        },
    },
};

const ko: EchoSummaryText = {
    userSettings: {
        clearToDefault: '관리자 기본값으로 되돌리기',
        description: '전날 참여한 대화 요약을 지정한 시간에 DM으로 받습니다. 관리자 기본 시간 대신 개인 시간을 따로 지정하거나, 개인 알림을 끌 수 있습니다.',
        enableDelivery: '개인 발송 사용',
        exampleLabel: '쉼표 또는 공백으로 여러 시간을 입력할 수 있습니다. 예:',
        exampleValue: '09:00, 13:30',
        heading: '전날 대화 요약 DM',
        inputPlaceholder: '09:00,13:30',
        resetError: '개인 설정을 지우지 못했습니다.',
        resetSuccess: '개인 설정을 지우고 관리자 기본값으로 되돌렸습니다.',
        saveError: '설정을 저장하지 못했습니다.',
        saveIdle: '저장',
        saveSaving: '저장 중...',
        saveSuccessDisabled: '개인 발송이 비활성화되었습니다.',
        saveSuccessEnabled: '개인 발송 시간이 저장되었습니다.',
        sectionTitle: '개인 발송 시간',
        validateEmpty: '최소 1개 이상의 HH:mm 시간을 입력해 주세요.',
    },
    adminConsole: {
        header: 'Echo Summary는 사용자가 전날 참여한 Mattermost 대화를 수집해, 설정한 vLLM OpenAI 호환 Chat Completions API로 요약하고, 지정한 시간에 DM으로 전달합니다.',
        sections: {
            vllm: {
                title: 'vLLM 요약 설정',
                settings: {
                    VLLMBaseURL: {
                        displayName: 'vLLM Base URL',
                        helpText: '루트 URL 또는 /v1 경로까지 넣으면 됩니다. 플러그인이 /chat/completions 경로를 자동으로 붙입니다.',
                        placeholder: 'https://vllm.example.com/v1',
                    },
                    VLLMAPIKey: {
                        displayName: 'vLLM API 키',
                        helpText: '필요한 경우 Authorization Bearer 토큰으로 전달됩니다.',
                        placeholder: 'sk-...',
                    },
                    VLLMModel: {
                        displayName: '요약 모델명',
                        helpText: 'Chat Completions 요청의 model 값으로 그대로 전달됩니다.',
                        placeholder: 'qwen2.5-14b-instruct',
                    },
                    DefaultPrompt: {
                        displayName: '기본 요약 프롬프트',
                        helpText: 'system 프롬프트로 사용됩니다. 비워두면 Echo Summary 기본 프롬프트를 사용합니다.',
                    },
                    RequestTimeoutSeconds: {
                        displayName: '요약 요청 타임아웃(초)',
                        helpText: 'vLLM 응답을 기다리는 최대 시간입니다.',
                    },
                },
            },
            scope: {
                title: '대상 사용자 및 스케줄',
                settings: {
                    NotificationTimezone: {
                        displayName: '기본 타임존',
                        helpText: '전날 기준일과 사용자별 발송 시간을 계산할 때 사용합니다.',
                        placeholder: 'Asia/Seoul',
                    },
                    DefaultTimeSlots: {
                        displayName: '기본 발송 시간',
                        helpText: '사용자 개인 설정이 없을 때 적용되는 기본 시간입니다. 쉼표로 여러 시간을 넣을 수 있습니다.',
                        placeholder: '09:00,13:30',
                    },
                    TargetUsernames: {
                        displayName: '대상 사용자 범위',
                        helpText: '비워두면 모든 활성 사용자를 대상으로 하고, 값을 넣으면 지정한 username만 대상으로 삼습니다.',
                        placeholder: 'alice,bob',
                    },
                    IncludeMentionedThreads: {
                        displayName: '멘션된 스레드 포함',
                        helpText: '사용자가 직접 글을 쓰지 않았더라도 전날 멘션된 스레드를 요약 후보에 포함합니다.',
                    },
                },
            },
            collection: {
                title: '수집 범위',
                settings: {
                    MaxThreadsPerUser: {
                        displayName: '사용자당 최대 스레드 수',
                        helpText: '하루에 참여한 대화가 많을 때 최신 순으로 이 개수까지만 요약합니다.',
                    },
                    MaxContextCharacters: {
                        displayName: '요청당 최대 문자 수',
                        helpText: '수집한 문맥을 Chat Completions 요청으로 나눌 때 사용하는 문자 수 기준입니다.',
                    },
                    ContextMessagesBefore: {
                        displayName: 'anchor 이전 문맥 수',
                        helpText: '사용자가 작성한 anchor 메시지 앞쪽에서 함께 포함할 최대 메시지 수입니다.',
                    },
                    ContextMessagesAfter: {
                        displayName: 'anchor 이후 문맥 수',
                        helpText: '사용자가 작성한 anchor 메시지 뒤쪽에서 함께 포함할 최대 메시지 수입니다.',
                    },
                },
            },
        },
    },
};

export function isKoreanLocale(locale?: string): boolean {
    return locale?.toLowerCase().startsWith('ko') ?? false;
}

export function getEchoSummaryText(locale?: string): EchoSummaryText {
    return isKoreanLocale(locale) ? ko : en;
}

export function formatInvalidTimeFormat(locale: string | undefined, token: string): string {
    if (isKoreanLocale(locale)) {
        return `잘못된 시간 형식입니다: ${token}`;
    }
    return `Invalid time format: ${token}`;
}

export function formatInvalidTimeValue(locale: string | undefined, token: string): string {
    if (isKoreanLocale(locale)) {
        return `잘못된 시간 값입니다: ${token}`;
    }
    return `Invalid time value: ${token}`;
}
