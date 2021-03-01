import * as React from 'react';
import {FormattedMessage} from 'react-intl';

import {Post} from 'mattermost-redux/types/posts';
import {Theme} from 'mattermost-redux/types/preferences';
// import {ActionResult} from 'mattermost-redux/types/actions';
import Client from '../../client';
import Svgs from '../../constants/svgs';

import {makeStyleFromTheme} from 'mattermost-redux/utils/theme_utils';

export type Props = {
    post?: Post,
    theme: Theme,
    creatorName: string
}

type State = {
    error: string
}

export class PostTypeSkyroom extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props);

        this.state = {
            error: ''
        };
    }

    componentDidMount() {
    }

    joinMeetingClicked = (e: React.MouseEvent) => {
        e.preventDefault();
        if (this.props.post) {
            const postProps = this.props.post.props;
            Client.joinMeeting(postProps.join_parameter).then((res: any) => {
                const skyroomLink = res['skyroom_link'] as string;
                if (skyroomLink && skyroomLink.startsWith('https://')) {
                    window.open(skyroomLink, '_blank');
                }
            }).catch(() => {
                this.setState({...this.state, error: 'skyroom.failed'});
            });
        }
    }

    renderUntilDate = (post: Post, style: any): React.ReactNode => {
        const postProps = post.props;

        if (postProps.valid_until) {
            const miliseconds = 1000;
            const date = new Date(postProps.valid_until * miliseconds);
            const dateStr = date.toString();
            //let dateStr = postProps.jwt_meeting_valid_until;
            // if (!isNaN(date.getTime())) {
            //     dateStr = date.toString();
            // }
            return (
                <div style={style.validUntil}>
                    <FormattedMessage
                        id='skyroom.link-valid-until'
                        defaultMessage=' Meeting link valid until: '
                    />
                    <b>{dateStr}</b>
                </div>
            );
        }
        return null;
    }

    render() {
        const style = getStyle(this.props.theme);
        const post = this.props.post;
        if (!post) {
            return null;
        }

        const postProps = post.props;

        const preText = (
            <FormattedMessage
                id='skyroom.creator-has-invited-to-meeting'
                defaultMessage='{creator} invited you to join a meeting'
                values={{creator: this.props.creatorName}}
            />
        );

        let title = (
            <FormattedMessage
                id='skyroom.default-title'
                defaultMessage='Skyroom Meeting'
            />
        );
        if (postProps.meeting_topic) {
            title = postProps.meeting_topic;
        }
        const error = this.state.error.length && (
            <FormattedMessage
                id='skyroom.failed'
                defaultMessage='Skyroom Failed to join the meeting'
            />
        );

        return (
            <div>
                {preText}
                <div style={style.attachment}>
                    <div style={style.content}>
                        <div style={style.container}>
                            <h1 style={style.title}>
                                {title}
                            </h1>
                            <span>
                                <a
                                    target='_blank'
                                    rel='noopener noreferrer'
                                    onClick={this.joinMeetingClicked}
                                >
                                    <FormattedMessage
                                        id='skyroom.join-meeting'
                                        defaultMessage='JOIN MEETING'
                                    />
                                </a>
                            </span>
                            <div>
                                <div style={style.body}>
                                    <div>
                                        <a
                                            className='btn btn-lg btn-primary'
                                            style={style.button}
                                            target='_blank'
                                            rel='noopener noreferrer'
                                            onClick={this.joinMeetingClicked}
                                        >
                                            <i
                                                style={style.buttonIcon}
                                                dangerouslySetInnerHTML={{__html: Svgs.VIDEO_CAMERA_3}}
                                            />
                                            <FormattedMessage
                                                id='skyroom.join-meeting'
                                                defaultMessage='JOIN MEETING'
                                            />
                                        </a>
                                    </div>
                                    {this.renderUntilDate(post, style)}
                                </div>
                                {error}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

const getStyle = makeStyleFromTheme((theme) => {
    return {
        attachment: {
            marginLeft: '-20px',
            position: 'relative'
        },
        content: {
            borderRadius: '4px',
            borderStyle: 'solid',
            borderWidth: '1px',
            borderColor: '#BDBDBF',
            margin: '5px 0 5px 20px',
            padding: '2px 5px'
        },
        container: {
            borderLeftStyle: 'solid',
            borderLeftWidth: '4px',
            padding: '10px',
            borderLeftColor: '#89AECB'
        },
        body: {
            overflowX: 'auto',
            overflowY: 'hidden',
            paddingRight: '5px',
            width: '100%'
        },
        title: {
            fontSize: '16px',
            fontWeight: '600',
            height: '22px',
            lineHeight: '18px',
            margin: '5px 0 1px 0',
            padding: '0'
        },
        button: {
            fontFamily: 'Open Sans',
            fontSize: '12px',
            fontWeight: 'bold',
            letterSpacing: '1px',
            lineHeight: '19px',
            marginTop: '12px',
            borderRadius: '4px',
            color: theme.buttonColor
        },
        buttonIcon: {
            paddingRight: '8px',
            fill: theme.buttonColor
        },
        validUntil: {
            marginTop: '10px'
        }
    };
});
