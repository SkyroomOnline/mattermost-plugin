import {Client4} from 'mattermost-redux/client';
import {ClientError} from 'mattermost-redux/client/client4';

import {id} from '../manifest';

export default class Client {
    url: string;

    constructor() {
        this.url = '/plugins/' + id;
        if ((window as any).basename) {
            this.url = (window as any).basename + '/plugins/' + id;
        }
    }

    startMeeting = async (channelId: string) => {
        return this.doGet(`${this.url}/api/v1/start?channelId=${channelId}`);
    }

    joinMeeting = async (parameter: string) => {
        return this.doGet(`${this.url}/api/v1/join?p=${parameter}`);
    }

    doGet = async (url: string, headers: any = {}) => {
        const options = {
            method: 'get',
            headers
        };

        const response = await fetch(url, Client4.getOptions(options));

        if (response.ok) {
            return response.json();
        }

        const text = await response.text();

        throw new ClientError(Client4.url, {
            message: text || '',
            status_code: response.status,
            url
        });
    }
}
