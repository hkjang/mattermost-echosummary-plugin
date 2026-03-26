import {formatInvalidTimeFormat, formatInvalidTimeValue, getEchoSummaryText, isKoreanLocale} from './i18n';

test('returns Korean text for Korean locales', () => {
    expect(isKoreanLocale('ko')).toBe(true);
    expect(isKoreanLocale('ko-KR')).toBe(true);
    expect(getEchoSummaryText('ko-KR').userSettings.saveIdle).toBe('저장');
});

test('returns English text for non-Korean locales', () => {
    expect(isKoreanLocale('en')).toBe(false);
    expect(getEchoSummaryText('en').userSettings.saveIdle).toBe('Save');
});

test('formats localized time validation errors', () => {
    expect(formatInvalidTimeFormat('ko', '9:00')).toBe('잘못된 시간 형식입니다: 9:00');
    expect(formatInvalidTimeValue('en', '24:00')).toBe('Invalid time value: 24:00');
});
