import {connect} from 'react-redux';
import {bindActionCreators, Dispatch, ActionCreatorsMapObject} from 'redux';

import {getTheme} from 'mattermost-redux/selectors/entities/preferences';
import {GenericAction, ActionFunc, ActionResult} from 'mattermost-redux/types/actions';

import {Post} from 'mattermost-redux/types/posts';

import {GlobalState} from '../../types';
import {displayUsernameForUser} from '../../utils/user_utils';
import {startMeeting, joinMeeting} from '../../actions';

import {PostTypeSkyroom} from './post_type_skyroom';

type OwnProps = {
    post: Post,
}

function mapStateToProps(state: GlobalState, ownProps: OwnProps) {
    const post = ownProps.post;
    const user = state.entities.users.profiles[post.user_id];

    return {
        ...ownProps,
        theme: getTheme(state),
        creatorName: displayUsernameForUser(user, state.entities.general.config)
    };
}

type Actions = {
    startMeeting: (channelId: string) => Promise<ActionResult>,
    joinMeeting: (channelId: string, parameter: string) => ActionResult,
}

function mapDispatchToProps(dispatch: Dispatch<GenericAction>) {
    return {
        actions: bindActionCreators<ActionCreatorsMapObject<ActionFunc>, Actions>({
            startMeeting,
            joinMeeting
        }, dispatch)
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(PostTypeSkyroom);
