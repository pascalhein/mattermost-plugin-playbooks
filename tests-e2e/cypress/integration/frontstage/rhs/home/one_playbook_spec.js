// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import * as TIMEOUTS from '../../../../fixtures/timeouts';

describe('rhs home', () => {
    let testTeam;
    let testUser;
    let testPlaybook;

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

    it('header should indicate no running playbook', () => {
        cy.get('#rhsContainer').should('exist').within(() => {
            // * Verify welcome text
            cy.findByText('This channel is not running any playbook.').should('be.visible');

            // * Verify user playbooks listed.
            cy.findByText('Your Playbooks').should('be.visible');
            cy.findByText(testPlaybook.title);
        });
    });

    it('run should launch interactive dialog with playbook pre-selected', () => {
        cy.get('#rhsContainer').should('exist').within(() => {
            cy.findByText('Run').eq(0).click();
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
});
