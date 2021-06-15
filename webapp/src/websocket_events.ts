// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {Dispatch} from 'redux';

import {GetStateFunc} from 'mattermost-redux/types/actions';
import {Post} from 'mattermost-redux/types/posts';
import {WebSocketMessage} from 'mattermost-redux/types/websocket';
import {getCurrentTeam, getCurrentTeamId} from 'mattermost-redux/selectors/entities/teams';
import {getCurrentUserId} from 'mattermost-redux/selectors/entities/users';

import {navigateToUrl} from 'src/browser_routing';
import {
    incidentCreated, incidentUpdated,
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
import {PlaybookRun, isPlaybookRun, StatusPost} from 'src/types/incident';

export const websocketSubscribersToPlaybookRunUpdate = new Set<(incident: PlaybookRun) => void>();

export function handleReconnect(getState: GetStateFunc, dispatch: Dispatch) {
    return async (): Promise<void> => {
        const currentTeam = getCurrentTeam(getState());
        const currentUserId = getCurrentUserId(getState());

        if (!currentTeam || !currentUserId) {
            return;
        }

        const fetched = await fetchPlaybookRuns({
            team_id: currentTeam.id,
            member_id: currentUserId,
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
                console.error('received a websocket data payload that was not an incident in handleWebsocketPlaybookRunUpdate:', data);
            }
        }
        const incident = data as PlaybookRun;

        dispatch(incidentUpdated(incident));

        websocketSubscribersToPlaybookRunUpdate.forEach((fn) => fn(incident));
    };
}

export function handleWebsocketPlaybookRunCreated(getState: GetStateFunc, dispatch: Dispatch) {
    return (msg: WebSocketMessage<{ payload: string }>): void => {
        if (!msg.data.payload) {
            return;
        }
        const payload = JSON.parse(msg.data.payload);
        const data = payload.incident;

        // eslint-disable-next-line no-process-env
        if (process.env.NODE_ENV !== 'production') {
            if (!isPlaybookRun(data)) {
                // eslint-disable-next-line no-console
                console.error('received a websocket data payload that was not an incident in handleWebsocketPlaybookRunCreate:', data);
            }
        }
        const incident = data as PlaybookRun;

        dispatch(incidentCreated(incident));

        if (payload.client_id !== clientId(getState())) {
            return;
        }

        const currentTeam = getCurrentTeam(getState());

        // Navigate to the newly created channel
        const url = `/${currentTeam.name}/channels/${incident.channel_id}`;
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
                const incident = await fetchPlaybookRunByChannel(msg.broadcast.channel_id);
                dispatch(receivedTeamPlaybookRuns([incident]));
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
    let incident: PlaybookRun;
    try {
        incident = await fetchPlaybookRunByChannel(post.channel_id);
    } catch (err) {
        return null;
    }

    if (incident.status_posts.find((value: StatusPost) => post.id === value.id)) {
        return incident;
    }

    return null;
}

export const handleWebsocketPostEditedOrDeleted = (getState: GetStateFunc, dispatch: Dispatch) => {
    return async (msg: WebSocketMessage<{ post: string }>) => {
        const incidentsMap = myPlaybookRunsMap(getState());
        if (incidentsMap[msg.broadcast.channel_id]) {
            const incident = await getPlaybookRunFromStatusUpdate(JSON.parse(msg.data.post));
            if (incident) {
                dispatch(incidentUpdated(incident));
                websocketSubscribersToPlaybookRunUpdate.forEach((fn) => fn(incident));
            }
        }
    };
};

export const handleWebsocketChannelUpdated = (getState: GetStateFunc, dispatch: Dispatch) => {
    return async (msg: WebSocketMessage<{ channel: string }>) => {
        const channel = JSON.parse(msg.data.channel);

        // Ignore updates to non-incident channels.
        const incidentsMap = myPlaybookRunsMap(getState());
        if (!incidentsMap[channel.id]) {
            return;
        }

        // Fetch the updated incident, since some metadata (like incident name) comes directly
        // from the channel, and the plugin cannot detect channel update events for itself.
        const incident = await fetchPlaybookRunByChannel(channel.id);
        if (incident) {
            dispatch(incidentUpdated(incident));
        }
    };
};

export const handleWebsocketChannelViewed = (getState: GetStateFunc, dispatch: Dispatch) => {
    return async (msg: WebSocketMessage<{ channel_id: string }>) => {
        const channelId = msg.data.channel_id;

        // If this isn't an incident channel, stop
        const incident = myPlaybookRunsMap(getState())[channelId];
        if (!incident) {
            return;
        }

        if (!hasViewedByChannelID(getState())[channelId]) {
            const hasViewed = await fetchCheckAndSendMessageOnJoin(incident.id, channelId);
            if (hasViewed) {
                dispatch(setHasViewedChannel(channelId));
            }
        }
    };
};
