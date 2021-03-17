import * as React from 'react';

import {Channel} from 'mattermost-redux/types/channels';
import {Post} from 'mattermost-redux/types/posts';

import Icon from './components/icon';
import PostTypeSkyroom from './components/post_type_skyroom';
import I18nProvider from './components/i18n_provider';
import reducer from './reducers';
import {startMeeting} from './actions';

class PluginClass {
    initialize(registry: any, store: any) {
        registry.registerReducer(reducer);
        registry.registerChannelHeaderButtonAction(
            <Icon/>,
            (channel: Channel) => {
                store.dispatch(startMeeting(channel.id));
            },
            'ّInvite to Skyroom'
        );
        registry.registerPostTypeComponent('custom_skyroom', (props: {post: Post}) => (<I18nProvider><PostTypeSkyroom post={props.post}/></I18nProvider>));
        // registry.registerWebSocketEventHandler('custom_skyroom_config_update', () => store.dispatch(loadConfig()));
        // store.dispatch(loadConfig());
    }

    uninitialize() {
    }
}

(global as any).window.registerPlugin('online.skyroom.skyroom', new PluginClass());
