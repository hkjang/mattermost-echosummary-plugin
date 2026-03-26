import {localizeAdminConsoleConfig} from './admin_console_localization';

test('localizes echo summary admin schema for Korean locale', () => {
    const config = {
        plugins: [{
            id: 'com.mattermost.echosummary',
            settings_schema: {
                header: 'old',
                sections: [{
                    key: 'vllm',
                    title: 'old title',
                    settings: [{
                        key: 'VLLMModel',
                        display_name: 'old name',
                        help_text: 'old help',
                    }],
                }],
            },
        }],
    };

    localizeAdminConsoleConfig(config, 'ko');

    const plugin = config.plugins[0];
    const schema = plugin.settings_schema;
    expect(schema.header).toContain('전날 참여한 Mattermost 대화');
    expect(schema.sections[0].title).toBe('vLLM 요약 설정');
    expect(schema.sections[0].settings[0].display_name).toBe('요약 모델명');
});

test('falls back to English for non-Korean locale', () => {
    const config = {
        id: 'com.mattermost.echosummary',
        settings_schema: {
            header: 'old',
            sections: [{
                key: 'scope',
                title: 'old title',
                settings: [{
                    key: 'DefaultTimeSlots',
                    display_name: 'old name',
                    help_text: 'old help',
                }],
            }],
        },
    };

    localizeAdminConsoleConfig(config, 'en-US');

    expect(config.settings_schema.header).toContain('previous-day Mattermost conversations');
    expect(config.settings_schema.sections[0].title).toBe('Target users and schedule');
    expect(config.settings_schema.sections[0].settings[0].display_name).toBe('Default delivery times');
});
