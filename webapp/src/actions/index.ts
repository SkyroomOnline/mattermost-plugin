import {PostTypes} from 'mattermost-redux/action_types';
import {DispatchFunc, GetStateFunc, ActionFunc, ActionResult} from 'mattermost-redux/types/actions';
import {Post} from 'mattermost-redux/types/posts';
import Client from '../client';

export function startMeeting(channelId: string): ActionFunc {
    return async (dispatch: DispatchFunc, getState: GetStateFunc): Promise<ActionResult> => {
        try {
            const result = await Client.startMeeting(channelId);
            return {data: result};
        } catch (error) {
            const post: Post = {
                id: 'skyroomPlugin' + Date.now(),
                create_at: Date.now(),
                update_at: 0,
                edit_at: 0,
                delete_at: 0,
                is_pinned: false,
                user_id: getState().entities.users.currentUserId,
                channel_id: channelId,
                root_id: '',
                parent_id: '',
                original_id: '',
                reply_count: 0,
                message: 'We could not start a meeting at this time.',
                type: 'system_ephemeral',
                props: {},
                metadata: {
                    embeds: [],
                    emojis: [],
                    files: [],
                    images: {},
                    reactions: []
                },
                hashtags: '',
                pending_post_id: ''
            };

            dispatch({
                type: PostTypes.RECEIVED_POSTS,
                data: {
                    order: [],
                    posts: {
                        [post.id]: post
                    }
                },
                channelId
            });

            return {error};
        }
    };
}

export function joinMeeting(channelId: string, parameter: string): ActionFunc {
    return async (dispatch: DispatchFunc, getState: GetStateFunc): Promise<ActionResult> => {
        try {
            const result = await Client.joinMeeting(parameter);
            return {data: result};
        } catch (error) {
            const post: Post = {
                id: 'skyroomPlugin' + Date.now(),
                create_at: Date.now(),
                update_at: 0,
                edit_at: 0,
                delete_at: 0,
                is_pinned: false,
                user_id: getState().entities.users.currentUserId,
                channel_id: channelId,
                root_id: '',
                parent_id: '',
                original_id: '',
                reply_count: 0,
                message: 'Failed to join the meeting at this time.',
                type: 'system_ephemeral',
                props: {},
                metadata: {
                    embeds: [],
                    emojis: [],
                    files: [],
                    images: {},
                    reactions: []
                },
                hashtags: '',
                pending_post_id: ''
            };

            dispatch({
                type: PostTypes.RECEIVED_POSTS,
                data: {
                    order: [],
                    posts: {
                        [post.id]: post
                    }
                },
                channelId
            });

            return {error};
        }
    };
}

