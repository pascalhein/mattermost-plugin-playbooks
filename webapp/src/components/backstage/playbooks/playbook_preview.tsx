// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import styled from 'styled-components';
import React from 'react';

import {PlaybookWithChecklist} from 'src/types/playbook';

interface Props {
    playbook: PlaybookWithChecklist;
}

const PlaybookPreview = (props: Props) => {
    return (
        <Container>
            <Content>
                <Section title={'Description'}/>
                <Section title={'Checklists'}/>
                <Section title={'Actions'}/>
                <Section title={'Status updates'}/>
                <Section title={'Retrospective'}/>
                <Section title={'Other customizations'}/>
            </Content>
            <Spacer/>
            <Navbar/>
        </Container>
    );
};

interface SectionProps {
    title: string
    children?: React.ReactNode
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
`;

const Content = styled.div`
    display: flex;
    flex-direction: column;

    border: 1px solid red;

    flex-grow: 1;
`;

const SectionWrapper = styled.div`
    :not(:last-child) {
        margin-bottom: 40px;
    }
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

    border: 1px solid blue;
`;

export default PlaybookPreview;
