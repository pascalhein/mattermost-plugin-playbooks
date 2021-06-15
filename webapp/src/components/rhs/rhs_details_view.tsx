// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {useSelector} from 'react-redux';

import {GlobalState} from 'mattermost-redux/types/store';

import {RHSContainer, RHSContent} from 'src/components/rhs/rhs_shared';
import RHSTabView from 'src/components/rhs/rhs_tab_view';
import {RHSTabState} from 'src/types/rhs';
import {currentPlaybookRun, currentRHSTabState} from 'src/selectors';
import RHSAbout from 'src/components/rhs/rhs_about';
import RHSPlaybookRunTasks from 'src/components/rhs/rhs_incident_tasks';
import RHSFooter from 'src/components/rhs/rhs_footer';
import {PlaybookRun} from 'src/types/incident';
import RHSTimeline from 'src/components/rhs/rhs_timeline';
import {useAllowTimelineViewInCurrentTeam} from 'src/hooks';

const RHSDetailsView = () => {
    const incident = useSelector<GlobalState, PlaybookRun | undefined>(currentPlaybookRun);
    const currentTabState = useSelector<GlobalState, RHSTabState>(currentRHSTabState);
    const allowTimelineView = useAllowTimelineViewInCurrentTeam();
    const showFooter = currentTabState !== RHSTabState.ViewingTimeline || allowTimelineView;

    if (!incident) {
        return null;
    }

    let currentView;
    switch (currentTabState) {
    case RHSTabState.ViewingAbout:
        currentView = <RHSAbout incident={incident}/>;
        break;
    case RHSTabState.ViewingTasks:
        currentView = <RHSPlaybookRunTasks incident={incident}/>;
        break;
    case RHSTabState.ViewingTimeline:
        currentView = <RHSTimeline incident={incident}/>;
        break;
    }

    return (
        <RHSContainer>
            <RHSContent>
                <RHSTabView/>
                {currentView}
                {showFooter && <RHSFooter incident={incident}/>}
            </RHSContent>
        </RHSContainer>
    );
};

export default RHSDetailsView;
