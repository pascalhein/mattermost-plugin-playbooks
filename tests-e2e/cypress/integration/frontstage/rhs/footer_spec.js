// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

// ***************************************************************
// - [#] indicates a test step (e.g. # Go to a page)
// - [*] indicates an assertion (e.g. * Check the title)
// ***************************************************************

describe('playbook run rhs > footer', () => {
    const playbookName = 'Playbook (' + Date.now() + ')';
    let teamId;
    let userId;
    let playbookId;

    before(() => {
        // # Login as user-1
        cy.apiLogin('user-1');

        cy.apiGetTeamByName('ad-1').then((team) => {
            teamId = team.id;
            cy.apiGetCurrentUser().then((user) => {
                userId = user.id;

                // # Create a playbook
                cy.apiCreateTestPlaybook({
                    teamId: team.id,
                    title: playbookName,
                    userId: user.id,
                }).then((playbook) => {
                    playbookId = playbook.id;
                });
            });
        });

        beforeEach(() => {
            // # Size the viewport to show the RHS without covering posts.
            cy.viewport('macbook-13');

            // # Login as user-1
            cy.apiLogin('user-1');
        });

        describe('has navigation button', () => {
            beforeEach(() => {
                const now = Date.now();
                const playbookRunName = 'Playbook Run (' + now + ')';
                const playbookRunChannelName = 'playbook-run-' + now;

                // # Run the playbook
                cy.apiRunPlaybook({
                    teamId,
                    playbookId,
                    playbookRunName,
                    ownerUserId: userId,
                });

                // # Navigate directly to the application and the playbook run channel
                cy.visit('/ad-1/channels/' + playbookRunChannelName);

                // # Select the tasks tab
                cy.findByTestId('tasks').click();
            });

            it('should always say update status', () => {
                // * Verify that the button contains the expected text
                cy.get('#playbookRunRHSFooter button').contains('Update Incident');

                // # Click on the End Incident button
                cy.get('#playbookRunRHSFooter button').click();

                // * Verify that the interactive dialog is visible
                cy.get('#interactiveDialogModalLabel').contains('Confirm End Incident');
            });
        });
    });
});
