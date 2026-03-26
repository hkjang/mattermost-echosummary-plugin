// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import manifest from 'manifest';
import type {Store} from 'redux';

import type {GlobalState} from '@mattermost/types/store';

import {getCurrentUser} from 'mattermost-redux/selectors/entities/common';

import type {PluginRegistry} from 'types/mattermost-webapp';

import {localizeAdminConsoleConfig} from './admin_console_localization';
import {getEchoSummaryText} from './i18n';
import {DeliverySettingsSection} from './user_settings';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState>) {
        const locale = getCurrentUser(store.getState())?.locale || 'en';
        const text = getEchoSummaryText(locale);

        registry.registerAdminConsolePlugin((config: object) => {
            localizeAdminConsoleConfig(config, locale);
        });

        registry.registerUserSettings({
            id: manifest.id,
            uiName: manifest.name,
            sections: [{
                title: text.userSettings.sectionTitle,
                component: DeliverySettingsSection,
            }],
        });
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
