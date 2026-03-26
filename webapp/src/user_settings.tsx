import manifest from 'manifest';
import React, {useEffect, useMemo, useState} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import type {PreferenceType} from '@mattermost/types/preferences';
import type {GlobalState} from '@mattermost/types/store';

import {deletePreferences, savePreferences} from 'mattermost-redux/actions/preferences';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/common';
import {get as getPreference} from 'mattermost-redux/selectors/entities/preferences';
import {getCurrentUserId} from 'mattermost-redux/selectors/entities/users';

import {formatInvalidTimeFormat, formatInvalidTimeValue, getEchoSummaryText} from './i18n';

const preferenceCategory = `pp_${manifest.id}`;
const deliveryTimesName = 'delivery_times';
const disabledValue = 'off';
const defaultDraft = '09:00';

type SaveState = 'idle' | 'saving' | 'saved' | 'error';

const sectionStyle: React.CSSProperties = {
    display: 'grid',
    gap: '12px',
    maxWidth: '720px',
    padding: '8px 0',
};

const helperStyle: React.CSSProperties = {
    color: 'rgba(63, 67, 80, 0.72)',
    fontSize: '12px',
    lineHeight: 1.5,
};

const buttonRowStyle: React.CSSProperties = {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
};

const primaryButtonStyle: React.CSSProperties = {
    background: '#1d4ed8',
    border: 'none',
    borderRadius: '4px',
    color: '#ffffff',
    cursor: 'pointer',
    fontSize: '13px',
    fontWeight: 600,
    padding: '8px 14px',
};

const secondaryButtonStyle: React.CSSProperties = {
    background: '#ffffff',
    border: '1px solid #cbd5e1',
    borderRadius: '4px',
    color: '#1f2937',
    cursor: 'pointer',
    fontSize: '13px',
    fontWeight: 600,
    padding: '8px 14px',
};

const inputStyle: React.CSSProperties = {
    border: '1px solid #cbd5e1',
    borderRadius: '4px',
    fontSize: '14px',
    padding: '10px 12px',
    width: '100%',
};

function normalizeTimes(raw: string): string {
    const tokens = raw.split(/[\s,;]+/).map((value) => value.trim()).filter(Boolean);
    const unique = Array.from(new Set(tokens));
    unique.sort();
    return unique.join(',');
}

function validateTimes(raw: string, locale: string | undefined, emptyMessage: string): string | null {
    const tokens = raw.split(/[\s,;]+/).map((value) => value.trim()).filter(Boolean);
    if (tokens.length === 0) {
        return emptyMessage;
    }

    for (const token of tokens) {
        if (!(/^\d{2}:\d{2}$/).test(token)) {
            return formatInvalidTimeFormat(locale, token);
        }

        const [hour, minute] = token.split(':').map(Number);
        if (hour > 23 || minute > 59) {
            return formatInvalidTimeValue(locale, token);
        }
    }

    return null;
}

export const DeliverySettingsSection = () => {
    const dispatch = useDispatch<any>();
    const currentUserId = useSelector(getCurrentUserId);
    const currentLocale = useSelector((state: GlobalState) => getCurrentUser(state)?.locale || 'en');
    const storedValue = useSelector((state: GlobalState) => getPreference(state, preferenceCategory, deliveryTimesName, ''));
    const text = getEchoSummaryText(currentLocale).userSettings;

    const [enabled, setEnabled] = useState(storedValue !== disabledValue);
    const [draft, setDraft] = useState(storedValue && storedValue !== disabledValue ? storedValue : defaultDraft);
    const [saveState, setSaveState] = useState<SaveState>('idle');
    const [message, setMessage] = useState('');

    useEffect(() => {
        setEnabled(storedValue !== disabledValue);
        setDraft(storedValue && storedValue !== disabledValue ? storedValue : defaultDraft);
    }, [storedValue]);

    const preferenceSkeleton = useMemo<PreferenceType>(() => ({
        user_id: currentUserId || '',
        category: preferenceCategory,
        name: deliveryTimesName,
        value: '',
    }), [currentUserId]);

    const handleSave = async () => {
        if (!currentUserId) {
            return;
        }

        setSaveState('saving');
        setMessage('');

        let value = disabledValue;
        if (enabled) {
            const validationMessage = validateTimes(draft, currentLocale, text.validateEmpty);
            if (validationMessage) {
                setSaveState('error');
                setMessage(validationMessage);
                return;
            }

            value = normalizeTimes(draft);
        }

        const result = await dispatch(savePreferences(currentUserId, [{
            ...preferenceSkeleton,
            user_id: currentUserId,
            value,
        }]));

        if (result.error) {
            setSaveState('error');
            setMessage(text.saveError);
            return;
        }

        setSaveState('saved');
        setMessage(enabled ? text.saveSuccessEnabled : text.saveSuccessDisabled);
    };

    const handleReset = async () => {
        if (!currentUserId) {
            return;
        }

        setSaveState('saving');
        setMessage('');

        const result = await dispatch(deletePreferences(currentUserId, [{
            ...preferenceSkeleton,
            user_id: currentUserId,
        }]));

        if (result.error) {
            setSaveState('error');
            setMessage(text.resetError);
            return;
        }

        setSaveState('saved');
        setMessage(text.resetSuccess);
    };

    return (
        <div style={sectionStyle}>
            <div>
                <strong>{text.heading}</strong>
                <div style={helperStyle}>
                    {text.description}
                </div>
            </div>

            <label style={{display: 'flex', gap: '8px', alignItems: 'center'}}>
                <input
                    checked={enabled}
                    onChange={(event) => setEnabled(event.target.checked)}
                    type='checkbox'
                />
                {text.enableDelivery}
            </label>

            <div>
                <input
                    disabled={!enabled}
                    onChange={(event) => setDraft(event.target.value)}
                    placeholder={text.inputPlaceholder}
                    style={inputStyle}
                    value={draft}
                />
                <div style={helperStyle}>
                    {text.exampleLabel} <code>{text.exampleValue}</code>
                </div>
            </div>

            <div style={buttonRowStyle}>
                <button
                    disabled={saveState === 'saving'}
                    onClick={handleSave}
                    style={primaryButtonStyle}
                    type='button'
                >
                    {saveState === 'saving' ? text.saveSaving : text.saveIdle}
                </button>
                <button
                    disabled={saveState === 'saving'}
                    onClick={handleReset}
                    style={secondaryButtonStyle}
                    type='button'
                >
                    {text.clearToDefault}
                </button>
            </div>

            {message && (
                <div
                    style={{
                        ...helperStyle,
                        color: saveState === 'error' ? '#b91c1c' : '#0f766e',
                    }}
                >
                    {message}
                </div>
            )}
        </div>
    );
};
