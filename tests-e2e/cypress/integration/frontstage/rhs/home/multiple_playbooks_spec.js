// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import * as TIMEOUTS from '../../../../fixtures/timeouts';

describe('rhs home', () => {
    let testTeam;
    let testUser;
    let testPlaybook;
    let testPlaybookWithChannelNameTemplate;

    before(() => {
        cy.apiInitSetup().then(({team, user}) => {
            testTeam = team;
            testUser = user;

            // # Login as testUser
            cy.apiLogin(testUser);

            // # Create a playbook
            cy.apiCreateTestPlaybook({
                teamId: testTeam.id,
                title: 'Playbook (' + Date.now() + ')',
                userId: testUser.id,
            }).then((playbook) => {
                testPlaybook = playbook;
            });

            // # Create a playbook
            cy.apiCreateTestPlaybook({
                teamId: testTeam.id,
                title: 'Playbook with Default Name (' + Date.now() + ')',
                userId: testUser.id,
                channelNameTemplate: 'default name',
            }).then((playbook) => {
                testPlaybookWithChannelNameTemplate = playbook;
            });
        });
    });

    beforeEach(() => {
        // # Size the viewport to show the RHS without covering posts.
        cy.viewport('macbook-13');

        // # Login as testUser
        cy.apiLogin(testUser);

        // # Start in town square
        cy.wait(TIMEOUTS.ONE_SEC).visit(`/${testTeam.name}/channels/town-square`);

        // # Open RHS
        cy.get('#channel-header').within(() => {
            cy.get('#incidentIcon').should('exist').click();
        });
    });

    it('run should launch interactive dialog with playbook pre-selected and no channel name template configured', () => {
        cy.get('#rhsContainer').should('exist').within(() => {
            cy.findAllByText('Run').eq(0).click();
        });

        cy.get('#interactiveDialogModal').within(() => {
            // * Shows current user as owner.
            cy.findByText(`${testUser.first_name} ${testUser.last_name}`).should('be.visible');

            // * Verify playbook dropdown prompt
            cy.get('input').eq(0).should('have.value', testPlaybook.title);

            // * Verify run name prompt
            cy.get('input').eq(1).should('be.empty');
        });
    });

    it('run should launch interactive dialog with run name filled in when channel name template configured', () => {
        cy.get('#rhsContainer').should('exist').within(() => {
            cy.findAllByText('Run').eq(1).click();
        });

        cy.get('#interactiveDialogModal').within(() => {
            // * Shows current user as owner.
            cy.findByText(`${testUser.first_name} ${testUser.last_name}`).should('be.visible');

            // * Verify playbook dropdown prompt
            cy.get('input').eq(0).should('have.value', testPlaybookWithChannelNameTemplate.title);

            // * Verify run name prompt
            cy.get('input').eq(1).should('have.value', 'default name');
        });
    });
});
