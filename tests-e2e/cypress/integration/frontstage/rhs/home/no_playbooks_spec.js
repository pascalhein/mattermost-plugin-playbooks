// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import * as TIMEOUTS from '../../../../fixtures/timeouts';

describe('rhs home', () => {
    let testTeam;
    let testUser;

    before(() => {
        cy.apiInitSetup().then(({team, user}) => {
            testTeam = team;
            testUser = user;
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

    it('header should show welcome screen', () => {
        cy.get('#rhsContainer').should('exist').within(() => {
            // * Verify welcome text
            cy.findByText('Welcome to Playbooks!').should('be.visible');
            cy.findByText('Create playbook').should('be.visible');

            // * Verify no user playbooks listed.
            cy.findByText('Your Playbooks').should('not.exist');
        });
    });
});
