// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Dispatch} from 'redux';

import {GetStateFunc} from 'mattermost-redux/types/actions';
import {Post} from 'mattermost-redux/types/posts';
import {WebSocketMessage} from 'mattermost-redux/types/websocket';
import {getCurrentTeam, getCurrentTeamId} from 'mattermost-redux/selectors/entities/teams';
import {getCurrentUserId} from 'mattermost-redux/selectors/entities/users';

import {PlaybookRun, isPlaybookRun, StatusPost} from 'src/types/playbook_run';

import {navigateToUrl} from 'src/browser_routing';
import {
    playbookRunCreated, playbookRunUpdated,
    removedFromPlaybookRunChannel,
    receivedTeamPlaybookRuns,
    playbookCreated,
    playbookDeleted, setHasViewedChannel,
} from 'src/actions';
import {
    fetchCheckAndSendMessageOnJoin,
    fetchPlaybookRunByChannel,
    fetchPlaybookRuns,
} from 'src/client';
import {clientId, hasViewedByChannelID, myPlaybookRunsMap} from 'src/selectors';

export const websocketSubscribersToPlaybookRunUpdate = new Set<(playbookRun: PlaybookRun) => void>();

export function handleReconnect(getState: GetStateFunc, dispatch: Dispatch) {
    return async (): Promise<void> => {
        const currentTeam = getCurrentTeam(getState());
        const currentUserId = getCurrentUserId(getState());

        if (!currentTeam || !currentUserId) {
            return;
        }

        const fetched = await fetchPlaybookRuns({
            page: 0,
            per_page: 0,
            team_id: currentTeam.id,
            participant_id: currentUserId,
        });

        dispatch(receivedTeamPlaybookRuns(fetched.items));
    };
}

export function handleWebsocketPlaybookRunUpdated(getState: GetStateFunc, dispatch: Dispatch) {
    return (msg: WebSocketMessage<{ payload: string }>): void => {
        if (!msg.data.payload) {
            return;
        }
        const data = JSON.parse(msg.data.payload);

        // eslint-disable-next-line no-process-env
        if (process.env.NODE_ENV !== 'production') {
            if (!isPlaybookRun(data)) {
                // eslint-disable-next-line no-console
                console.error('received a websocket data payload that was not a playbook run in handleWebsocketPlaybookRunUpdate:', data);
            }
        }
        const playbookRun = data as PlaybookRun;

        dispatch(playbookRunUpdated(playbookRun));

        websocketSubscribersToPlaybookRunUpdate.forEach((fn) => fn(playbookRun));
    };
}

export function handleWebsocketPlaybookRunCreated(getState: GetStateFunc, dispatch: Dispatch) {
    return (msg: WebSocketMessage<{ payload: string }>): void => {
        if (!msg.data.payload) {
            return;
        }
        const payload = JSON.parse(msg.data.payload);
        const data = payload.playbook_run;

        // eslint-disable-next-line no-process-env
        if (process.env.NODE_ENV !== 'production') {
            if (!isPlaybookRun(data)) {
                // eslint-disable-next-line no-console
                console.error('received a websocket data payload that was not a playbook run in handleWebsocketPlaybookRunCreate:', data);
            }
        }
        const playbookRun = data as PlaybookRun;

        dispatch(playbookRunCreated(playbookRun));

        if (payload.client_id !== clientId(getState())) {
            return;
        }

        const currentTeam = getCurrentTeam(getState());

        // Navigate to the newly created channel
        const url = `/${currentTeam.name}/channels/${playbookRun.channel_id}`;
        navigateToUrl(url);
    };
}

export function handleWebsocketPlaybookCreated(getState: GetStateFunc, dispatch: Dispatch) {
    return (msg: WebSocketMessage<{ payload: string }>): void => {
        if (!msg.data.payload) {
            return;
        }

        const payload = JSON.parse(msg.data.payload);

        dispatch(playbookCreated(payload.teamID));
    };
}

export function handleWebsocketPlaybookDeleted(getState: GetStateFunc, dispatch: Dispatch) {
    return (msg: WebSocketMessage<{ payload: string }>): void => {
        if (!msg.data.payload) {
            return;
        }

        const payload = JSON.parse(msg.data.payload);

        dispatch(playbookDeleted(payload.teamID));
    };
}

export function handleWebsocketUserAdded(getState: GetStateFunc, dispatch: Dispatch) {
    return async (msg: WebSocketMessage<{ team_id: string, user_id: string }>) => {
        const currentUserId = getCurrentUserId(getState());
        const currentTeamId = getCurrentTeamId(getState());
        if (currentUserId === msg.data.user_id && currentTeamId === msg.data.team_id) {
            try {
                const playbookRun = await fetchPlaybookRunByChannel(msg.broadcast.channel_id);
                dispatch(receivedTeamPlaybookRuns([playbookRun]));
            } catch (error) {
                if (error.status_code !== 404) {
                    throw error;
                }
            }
        }
    };
}

export function handleWebsocketUserRemoved(getState: GetStateFunc, dispatch: Dispatch) {
    return (msg: WebSocketMessage<{ channel_id: string, user_id: string }>) => {
        const currentUserId = getCurrentUserId(getState());
        if (currentUserId === msg.broadcast.user_id) {
            dispatch(removedFromPlaybookRunChannel(msg.data.channel_id));
        }
    };
}

async function getPlaybookRunFromStatusUpdate(post: Post): Promise<PlaybookRun | null> {
    let playbookRun: PlaybookRun;
    try {
        playbookRun = await fetchPlaybookRunByChannel(post.channel_id);
    } catch (err) {
        return null;
    }

    if (playbookRun.status_posts.find((value: StatusPost) => post.id === value.id)) {
        return playbookRun;
    }

    return null;
}

export const handleWebsocketPostEditedOrDeleted = (getState: GetStateFunc, dispatch: Dispatch) => {
    return async (msg: WebSocketMessage<{ post: string }>) => {
        const playbookRunsMap = myPlaybookRunsMap(getState());
        if (playbookRunsMap[msg.broadcast.channel_id]) {
            const playbookRun = await getPlaybookRunFromStatusUpdate(JSON.parse(msg.data.post));
            if (playbookRun) {
                dispatch(playbookRunUpdated(playbookRun));
                websocketSubscribersToPlaybookRunUpdate.forEach((fn) => fn(playbookRun));
            }
        }
    };
};

export const handleWebsocketChannelUpdated = (getState: GetStateFunc, dispatch: Dispatch) => {
    return async (msg: WebSocketMessage<{ channel: string }>) => {
        const channel = JSON.parse(msg.data.channel);

        // Ignore updates to non-playbook run channels.
        const playbookRunsMap = myPlaybookRunsMap(getState());
        if (!playbookRunsMap[channel.id]) {
            return;
        }

        // Fetch the updated playbook run, since some metadata (like playbook run name) comes directly
        // from the channel, and the plugin cannot detect channel update events for itself.
        const playbookRun = await fetchPlaybookRunByChannel(channel.id);
        if (playbookRun) {
            dispatch(playbookRunUpdated(playbookRun));
        }
    };
};

export const handleWebsocketChannelViewed = (getState: GetStateFunc, dispatch: Dispatch) => {
    return async (msg: WebSocketMessage<{ channel_id: string }>) => {
        const channelId = msg.data.channel_id;

        // If this isn't a playbook run channel, stop
        const playbookRun = myPlaybookRunsMap(getState())[channelId];
        if (!playbookRun) {
            return;
        }

        if (!hasViewedByChannelID(getState())[channelId]) {
            const hasViewed = await fetchCheckAndSendMessageOnJoin(playbookRun.id, channelId);
            if (hasViewed) {
                dispatch(setHasViewedChannel(channelId));
            }
        }
    };
};
