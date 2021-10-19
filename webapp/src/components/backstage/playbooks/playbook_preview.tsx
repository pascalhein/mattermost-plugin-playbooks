// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import styled, {css} from 'styled-components';
import React, {useState} from 'react';
import {useSelector} from 'react-redux';
import {useIntl} from 'react-intl';

import {Duration} from 'luxon';

import {GlobalState} from 'mattermost-redux/types/store';
import {getChannel} from 'mattermost-redux/selectors/entities/channels';

import {useDefaultMarkdownOptionsByTeamId} from 'src/hooks/general';
import {useAllowRetrospectiveAccess} from 'src/hooks';
import {PlaybookWithChecklist, Checklist, ChecklistItem} from 'src/types/playbook';
import {messageHtmlToComponent, formatText} from 'src/webapp_globals';
import {PillBox} from 'src/components/widgets/pill';
import {UserList} from 'src/components/rhs/rhs_participant';
import CollapsibleChecklist from 'src/components/rhs/collapsible_checklist';
import {ChecklistContainer} from 'src/components/rhs/rhs_checklists';
import {ChecklistItemDetails} from 'src/components/checklist_item';
import ProfileSelector from 'src/components/profile/profile_selector';
import {formatDuration} from 'src/components/formatted_duration';

interface Props {
    playbook: PlaybookWithChecklist;
}

const ChannelBadge = ({channelId} : { channelId: string }) => {
    const channel = useSelector((state: GlobalState) => getChannel(state, channelId));

    return (
        <Badge key={channelId}>
            <i className={'icon-globe icon-12'}/>
            {channel?.display_name}
        </Badge>
    );
};

const Badge = styled(PillBox)`
    font-size: 11px;
    height: 20px;
    line-height: 16px;
    display: flex;
    align-items: center;
    color: rgba(var(--center-channel-color-rgb), 0.72);

    padding-left: 1px;
    padding-right: 8px;

    :not(:last-child) {
        margin-right: 8px;
    }

    i {
        color: rgba(var(--center-channel-color-rgb), 0.72);
        margin-top: -1px;
        margin-right: 3px;
    }

    font-weight: 600;
    font-size: 11px;
`;

const TextBadge = styled(Badge)`
    text-transform: uppercase;

    padding: 0 6px;
`;

const ChannelBadges = styled.div`
    display: flex;
    flex-direction: row;
`;

const PlaybookPreview = (props: Props) => {
    const {formatMessage} = useIntl();
    const markdownOptions = useDefaultMarkdownOptionsByTeamId(props.playbook.team_id);
    const renderMarkdown = (msg: string) =>
        messageHtmlToComponent(
            formatText(msg, markdownOptions),
            true,
            {}
        );

    const initState: boolean[] = new Array(props.playbook.checklists.length).fill(false);
    const [checklistsCollapsed, setChecklistsCollapsed] = useState(initState);

    const retrospectiveAccess = useAllowRetrospectiveAccess();

    return (
        <Container>
            <Content>
                {/* <Section
                    title={formatMessage({defaultMessage: 'Checklists'})}
                    >
                    {props.playbook.checklists.map((checklist: Checklist, checklistIndex: number) => (
                    <CollapsibleChecklist
                    key={checklist.title}
                    title={checklist.title}
                    items={checklist.items}
                    index={checklistIndex}
                    collapsed={checklistsCollapsed[checklistIndex]}
                    setCollapsed={(newState) => {
                    const newArr = {...checklistsCollapsed};
                    newArr[checklistIndex] = newState;
                    setChecklistsCollapsed(newArr);
                    }}
                    disabled={true}
                    >
                    <ChecklistContainer className='checklist'>
                    {checklist.items.map((checklistItem: ChecklistItem, index: number) => {
                    return (
                    <ChecklistItemDetails
                    key={checklist.title + checklistItem.title}
                    checklistItem={checklistItem}
                    checklistNum={checklistIndex}
                    itemNum={index}
                    channelId={''}
                    playbookRunId={''}
                    dragging={false}
                    disabled={true}
                    />
                    );
                    })}
                    </ChecklistContainer>
                    </CollapsibleChecklist>
                    ))}
                    </Section> */}
                <Section title={formatMessage({defaultMessage: 'Actions'})}>
                    <Card>
                        <CardEntry
                            title={formatMessage({
                                defaultMessage:
                                    'Prompt to run this playbook when a user posts a message containing the keywords',
                            })}
                            iconName={'message-text-outline'}
                            extraInfo={(
                                <ChannelBadges>
                                    {props.playbook.signal_any_keywords.map((keyword) => (
                                        <TextBadge key={keyword}>{keyword}</TextBadge>
                                    ))}
                                </ChannelBadges>
                            )}
                            enabled={props.playbook.signal_any_keywords_enabled}
                        />
                        <CardEntry
                            title={formatMessage({
                                defaultMessage: 'When a run starts',
                            })}
                            iconName={'play'}
                        >
                            <CardSubEntry
                                title={formatMessage(
                                    {defaultMessage: 'Create a {isPublic, select, true {public} other {private}} channel'},
                                    {isPublic: props.playbook.create_public_playbook_run},
                                )}
                                enabled={true}
                            />
                            <CardSubEntry
                                title={formatMessage(
                                    {
                                        defaultMessage:
                                            'Invite {numInvitedUsers, plural, =0 {no members} =1 {one member} other {# members}} to the channel',
                                    },
                                    {numInvitedUsers: props.playbook.invited_user_ids.length}
                                )}
                                extraInfo={(
                                    <UserList
                                        userIds={props.playbook.invited_user_ids}
                                        sizeInPx={20}
                                    />
                                )}
                                enabled={props.playbook.invite_users_enabled}
                            />
                            <CardSubEntry
                                title={formatMessage({
                                    defaultMessage: 'Assign the owner role to',
                                })}
                                enabled={props.playbook.default_owner_enabled}
                                extraInfo={(
                                    <StyledProfileSelector
                                        selectedUserId={props.playbook.default_owner_id}
                                        placeholder={null}
                                        placeholderButtonClass={'NoAssignee-button'}
                                        profileButtonClass={'Assigned-button'}
                                        enableEdit={false}
                                        getUsers={() => Promise.resolve([])}
                                    />
                                )}
                            />
                            <CardSubEntry
                                title={formatMessage(
                                    {defaultMessage: 'Announce in the {oneChannel, plural, one {channel} other {channels}}'},
                                    {oneChannel: props.playbook.broadcast_channel_ids.length}
                                )}
                                enabled={props.playbook.broadcast_enabled}
                                extraInfo={(
                                    <ChannelBadges>
                                        {props.playbook.broadcast_channel_ids.map((id) => (
                                            <ChannelBadge
                                                key={id}
                                                channelId={id}
                                            />
                                        ))}
                                    </ChannelBadges>
                                )}
                            />
                            <CardSubEntry
                                title={formatMessage({
                                    defaultMessage: 'Update run summary',
                                })}
                                enabled={props.playbook.webhook_on_status_update_enabled}
                            >
                                {renderMarkdown(props.playbook.reminder_message_template)}
                            </CardSubEntry>
                            <CardSubEntry
                                title={formatMessage({
                                    defaultMessage: 'Send an outgoing webhook',
                                })}
                                enabled={props.playbook.webhook_on_creation_enabled}
                            >
                                {renderMarkdown(props.playbook.webhook_on_creation_urls.join('\n\n'))}
                            </CardSubEntry>
                        </CardEntry>
                        <CardEntry
                            title={formatMessage({
                                defaultMessage:
                                    'When a new member joins the channel',
                            })}
                            iconName={'account-outline'}
                        >
                            <CardSubEntry
                                title={formatMessage({
                                    defaultMessage: 'Send a welcome message',
                                })}
                                enabled={props.playbook.message_on_join_enabled}
                            >
                                {renderMarkdown(props.playbook.message_on_join)}
                            </CardSubEntry>
                            <CardSubEntry
                                title={formatMessage({
                                    defaultMessage:
                                        'Add the channel to the sidebar category',
                                })}
                                enabled={props.playbook.categorize_channel_enabled}
                                extraInfo={(
                                    <TextBadge>
                                        {props.playbook.category_name}
                                    </TextBadge>
                                )}
                            />
                        </CardEntry>
                        {props.playbook.export_channel_on_finished_enabled &&
                            <CardEntry
                                title={formatMessage({
                                    defaultMessage:
                                        'When a run is finished, export the channel',
                                })}
                                iconName={'flag-outline'}
                            />
                        }
                    </Card>
                </Section>
                <Section
                    title={formatMessage({defaultMessage: 'Status updates'})}
                >
                    <Card>
                        <CardEntry
                            title={formatMessage(
                                {defaultMessage: 'The owner will {reminderEnabled, select, true {be prompted to provide a status update every} other {not be prompted to provide a status update}}'},
                                {reminderEnabled: props.playbook.reminder_timer_default_seconds !== 0},
                            )}
                            iconName={'clock-outline'}
                            extraInfo={props.playbook.reminder_timer_default_seconds !== 0 && (
                                <TextBadge>
                                    {formatDuration(Duration.fromObject({seconds: props.playbook.reminder_timer_default_seconds}), 'long')}
                                </TextBadge>
                            )}
                        >
                            <CardSubEntry
                                title={formatMessage({
                                    defaultMessage: 'Update template',
                                })}
                                enabled={props.playbook.reminder_message_template !== ''}
                            >
                                {renderMarkdown(props.playbook.reminder_message_template)}
                            </CardSubEntry>
                        </CardEntry>
                        <CardEntry
                            title={formatMessage({
                                defaultMessage: 'When an update is posted',
                            })}
                            iconName={'message-check-outline'}
                        >
                            <CardSubEntry
                                title={formatMessage(
                                    {defaultMessage: 'Broadcast updates in the {oneChannel, plural, one {channel} other {channels}}'},
                                    {oneChannel: props.playbook.broadcast_channel_ids.length}
                                )}
                                enabled={props.playbook.broadcast_enabled}
                                extraInfo={(
                                    <ChannelBadges>
                                        {props.playbook.broadcast_channel_ids.map((id) => (
                                            <ChannelBadge
                                                key={id}
                                                channelId={id}
                                            />
                                        ))}
                                    </ChannelBadges>
                                )}
                            />
                            <CardSubEntry
                                title={formatMessage({
                                    defaultMessage: 'Send an outgoing webhook',
                                })}
                                enabled={props.playbook.webhook_on_status_update_enabled}
                            >
                                {renderMarkdown(props.playbook.webhook_on_status_update_urls.join('\n\n'))}
                            </CardSubEntry>
                        </CardEntry>
                    </Card>
                </Section>
                {retrospectiveAccess &&
                    <Section
                        title={formatMessage({defaultMessage: 'Retrospective'})}
                    >
                        <Card>
                            <CardEntry
                                title={formatMessage(
                                    {defaultMessage: 'The channel will be reminded to perform the retrospective {reminderEnabled, select, true {every} other {}}'},
                                    {reminderEnabled: props.playbook.retrospective_reminder_interval_seconds !== 0}
                                )}
                                iconName={'lightbulb-outline'}
                                extraInfo={(
                                    <TextBadge>
                                        {props.playbook.retrospective_reminder_interval_seconds === 0 ? 'ONCE' : formatDuration(Duration.fromObject({seconds: props.playbook.retrospective_reminder_interval_seconds}), 'long')}
                                    </TextBadge>
                                )}
                            >
                                <CardSubEntry
                                    title={formatMessage({
                                        defaultMessage:
                                            'Retrospective report template',
                                    })}
                                    enabled={props.playbook.retrospective_template !== ''}
                                >
                                    {renderMarkdown(props.playbook.retrospective_template)}
                                </CardSubEntry>
                            </CardEntry>
                        </Card>
                    </Section>
                }
            </Content>
            <Spacer/>
            <Navbar/>
        </Container>
    );
};

const StyledProfileSelector = styled(ProfileSelector)`
    margin-top: 0;
    height: 20px;

    .Assigned-button {
        border-radius: 16px;
        max-width: 100%;
        height: 20px;
        padding: 2px;
        padding-right: 6px;
        margin-top: 0;
        background: var(--center-channel-color-08);

        :hover {
            background: rgba(var(--center-channel-color-rgb), 0.16);
        }

        .image {
            width: 16px;
            height: 16px;
        }

font-weight: 600;
font-size: 11px;
line-height: 16px;

display: flex;
align-items: center;

    }
`;

interface SectionProps {
    title: string;
    children?: React.ReactNode;
}

const Section = ({title, children}: SectionProps) => {
    return (
        <SectionWrapper>
            <SectionTitle>{title}</SectionTitle>
            {children}
        </SectionWrapper>
    );
};

const Container = styled.main`
    display: flex;
    flex-direction: row;

    max-width: ${780 + 114 + 172}px;
    margin: auto;
    padding-top: 40px;
`;

const Content = styled.div`
    display: flex;
    flex-direction: column;
    max-width: 780px;

    flex-grow: 1;
`;

const SectionWrapper = styled.div`
    margin-bottom: 40px;
`;

const SectionTitle = styled.div`
    font-family: Metropolis, sans-serif;
    font-size: 20px;
    font-weight: 600;
    line-height: 28px;

    margin-bottom: 16px;
`;

const Spacer = styled.div`
    width: 114px;
`;

const Navbar = styled.nav`
    width: 172px;
    height: 340px;
`;

const Card = styled.div`
    background: var(--center-channel-bg);
    width: 100%;

    border: 1px solid rgba(61, 60, 64, 0.04);
    box-sizing: border-box;
    box-shadow: 0px 2px 3px rgba(0, 0, 0, 0.08);
    border-radius: 4px;

    padding: 16px;
    padding-left: 11px;
    padding-right: 20px;

    display: flex;
    flex-direction: column;
`;

interface CardEntryProps {
    iconName: string;
    title: string;
    extraInfo?: React.ReactNode;
    children?: React.ReactNode;
    className?: string;
    onClick?: () => void;
    enabled?: boolean;
}

const CardEntry = (props: CardEntryProps) => {
    if (!props.enabled) {
        return null;
    }

    return (
        <CardEntryWrapper
            className={props.className}
            onClick={props.onClick}
        >
            <CardEntryHeader>
                <i className={`icon-${props.iconName} icon-16`}/>
                <CardEntryTitle>{props.title}</CardEntryTitle>
                <ExtraInfo>{props.extraInfo}</ExtraInfo>
            </CardEntryHeader>
            {props.children && (
                <CardSubentries>{props.children}</CardSubentries>
            )}
        </CardEntryWrapper>
    );
};

const CardSubentries = styled.div`
    display: flex;
    flex-direction: column;
    margin-left: 22px;
    margin-top: 8px;

    font-size: 14px;
    color: rgba(var(--center-channel-color), 0.72);
`;

interface CardSubEntryProps {
    enabled: boolean;
    title: string;
    extraInfo?: React.ReactNode;
    children?: React.ReactNode;
}

const CardSubEntry = (props: CardSubEntryProps) => {
    const [open, setOpen] = useState(false);

    if (!props.enabled) {
        return null;
    }

    const icon = props.children ? <ChevronIcon open={open}/> : <MinusIcon/>;

    const toggleOpen = () => setOpen(!open);

    return (
        <CardSubEntryWrapper
            onClick={toggleOpen}
            withChildren={Boolean(props.children)}
        >
            <CardEntryHeader>
                {icon}
                <CardEntryTitle>{props.title}</CardEntryTitle>
                <ExtraInfo>{props.extraInfo}</ExtraInfo>
            </CardEntryHeader>
            {open && props.children && (
                <CardSubEntryContent>{props.children}</CardSubEntryContent>
            )}
        </CardSubEntryWrapper>
    );
};

const ChevronIcon = ({open}: {open: boolean}) => {
    return (
        <i className={`icon-${open ? 'chevron-down' : 'chevron-right'} icon-16`}/>
    );
};

const MinusIcon = () => {
    return (
        <SubtleIcon className={'icon-minus icon-16'}/>
    );
};

const SubtleIcon = styled.i`
    opacity: 0.48;
`;

const CardSubEntryContent = styled.div`
    margin-left: 30px;
    font-size: 13px;
    line-height: 16px;

    padding-right: 16px;
    padding-bottom: 11px;

    p {
        margin-bottom: 0;

        :not(:last-child) {
            margin-bottom: 6px;
        }
    }
`;

const CardEntryWrapper = styled.div`
    display: flex;
    flex-direction: column;
    border-radius: 4px;

    :not(:last-child) {
        margin-bottom: 20px;
    }
`;

const CardSubEntryWrapper = styled(CardEntryWrapper)<{withChildren: boolean}>`
    color: rgba(var(--center-channel-color-rgb), 0.72);

    ${({withChildren}) => withChildren && css`
        cursor: pointer;
        transition: background-color 0.2s linear 0s;
        :hover {
            background-color: rgba(var(--center-channel-color-rgb), 0.08);
        }

    `}

    && {
        :not(:last-child) {
            margin-bottom: 8px;
        }
    }

    font-size: 14px;
    line-height: 20px;
`;

const CardEntryHeader = styled.div`
    display: flex;
    flex-direction: row;
    align-items: center;
    height: 28px;

    i {
        color: rgba(var(--center-channel-color-rgb), 0.48);
    }

`;

const CardEntryTitle = styled.div`
    margin-left: 5px;
    margin-right: 8px;
`;

const ExtraInfo = styled.div``;

export default PlaybookPreview;
