import {GlobalState as ReduxGlobalState} from 'mattermost-redux/types/store';
// import {Post} from 'mattermost-redux/types/posts';

export type Config = {
    TeammateNameDisplay?: string
}

export type GlobalState = ReduxGlobalState & {
    'plugins-skyroom': {
    }
}
