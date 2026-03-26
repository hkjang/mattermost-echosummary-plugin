import manifest from 'manifest';

import {getEchoSummaryText} from './i18n';

type UnknownRecord = Record<string, unknown>;

function isRecord(value: unknown): value is UnknownRecord {
    return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isPluginNode(record: UnknownRecord): boolean {
    return record.id === manifest.id ||
        record.plugin_id === manifest.id ||
        record.pluginId === manifest.id;
}

function looksLikeEchoSummarySettingsSchema(record: UnknownRecord): boolean {
    if (!Array.isArray(record.sections)) {
        return false;
    }

    const sectionKeys = record.sections.
        filter(isRecord).
        map((section) => section.key).
        filter((key): key is string => typeof key === 'string');

    return ['vllm', 'scope', 'collection'].every((key) => sectionKeys.includes(key));
}

function setStringFields(record: UnknownRecord, fieldNames: string[], value: string | undefined) {
    if (!value) {
        return;
    }

    for (const fieldName of fieldNames) {
        record[fieldName] = value;
    }
}

function applySettingsSchemaLocalization(schema: UnknownRecord, locale?: string) {
    const text = getEchoSummaryText(locale).adminConsole;
    setStringFields(schema, ['header'], text.header);

    if (!Array.isArray(schema.sections)) {
        return;
    }

    for (const section of schema.sections) {
        if (!isRecord(section) || typeof section.key !== 'string') {
            continue;
        }

        const sectionText = text.sections[section.key];
        if (!sectionText) {
            continue;
        }

        setStringFields(section, ['title'], sectionText.title);

        if (!Array.isArray(section.settings)) {
            continue;
        }

        for (const setting of section.settings) {
            if (!isRecord(setting) || typeof setting.key !== 'string') {
                continue;
            }

            const settingText = sectionText.settings[setting.key];
            if (!settingText) {
                continue;
            }

            setStringFields(setting, ['display_name', 'displayName'], settingText.displayName);
            setStringFields(setting, ['help_text', 'helpText'], settingText.helpText);
            setStringFields(setting, ['placeholder'], settingText.placeholder);
        }
    }
}

function walkAndLocalize(node: unknown, locale?: string) {
    if (Array.isArray(node)) {
        for (const item of node) {
            walkAndLocalize(item, locale);
        }
        return;
    }

    if (!isRecord(node)) {
        return;
    }

    let settingsSchema: UnknownRecord | null = null;
    if (isRecord(node.settings_schema)) {
        settingsSchema = node.settings_schema;
    } else if (isRecord(node.settingsSchema)) {
        settingsSchema = node.settingsSchema;
    }

    if (settingsSchema && isPluginNode(node)) {
        applySettingsSchemaLocalization(settingsSchema, locale);
    } else if (looksLikeEchoSummarySettingsSchema(node)) {
        applySettingsSchemaLocalization(node, locale);
    }

    for (const value of Object.values(node)) {
        walkAndLocalize(value, locale);
    }
}

export function localizeAdminConsoleConfig(config: object, locale?: string) {
    walkAndLocalize(config, locale);
}
