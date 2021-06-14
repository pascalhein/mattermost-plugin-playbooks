package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	icClient "github.com/mattermost/mattermost-plugin-incident-collaboration/client"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-incident-collaboration/server/app"
	mock_app "github.com/mattermost/mattermost-plugin-incident-collaboration/server/app/mocks"
	mock_poster "github.com/mattermost/mattermost-plugin-incident-collaboration/server/bot/mocks"
	"github.com/mattermost/mattermost-plugin-incident-collaboration/server/config"
	mock_config "github.com/mattermost/mattermost-plugin-incident-collaboration/server/config/mocks"
)

func TestPlaybookRuns(t *testing.T) {
	var mockCtrl *gomock.Controller
	var handler *Handler
	var poster *mock_poster.MockPoster
	var logger *mock_poster.MockLogger
	var configService *mock_config.MockService
	var playbookService *mock_app.MockPlaybookService
	var playbookRunService *mock_app.MockPlaybookRunService
	var pluginAPI *plugintest.API
	var client *pluginapi.Client

	// mattermostHandler simulates the Mattermost server routing HTTP requests to a plugin.
	mattermostHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/plugins/com.mattermost.plugin-incident-management")
		r.Header.Add("Mattermost-User-ID", "testUserID")

		handler.ServeHTTP(w, r)
	})

	server := httptest.NewServer(mattermostHandler)
	t.Cleanup(server.Close)

	c, err := icClient.New(&model.Client4{Url: server.URL})
	require.NoError(t, err)

	reset := func(t *testing.T) {
		t.Helper()

		mockCtrl = gomock.NewController(t)
		configService = mock_config.NewMockService(mockCtrl)
		pluginAPI = &plugintest.API{}
		client = pluginapi.NewClient(pluginAPI)
		poster = mock_poster.NewMockPoster(mockCtrl)
		logger = mock_poster.NewMockLogger(mockCtrl)
		handler = NewHandler(client, configService, logger)
		playbookService = mock_app.NewMockPlaybookService(mockCtrl)
		playbookRunService = mock_app.NewMockPlaybookRunService(mockCtrl)
		NewPlaybookRunHandler(handler.APIRouter, playbookRunService, playbookService, client, poster, logger, configService)
	}

	setDefaultExpectations := func(t *testing.T) {
		t.Helper()

		configService.EXPECT().
			IsAtLeastE10Licensed().
			Return(true)

		configService.EXPECT().
			GetConfiguration().
			Return(&config.Configuration{
				EnabledTeams: []string{},
			})
	}

	t.Run("create valid incident, but it's disabled on this team", func(t *testing.T) {
		reset(t)

		configService.EXPECT().
			GetConfiguration().
			Return(&config.Configuration{
				EnabledTeams: []string{"notthisteam"},
			})

		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		withid := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
		}

		testPlaybookRun := app.PlaybookRun{
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			PlaybookID:  withid.ID,
			Checklists:  withid.Checklists,
		}

		playbookRunJSON, err := json.Marshal(testPlaybookRun)
		require.NoError(t, err)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents", bytes.NewBuffer(playbookRunJSON))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("create valid incident from dialog", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		withid := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			Description:             "description",
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		dialogRequest := model.SubmitDialogRequest{
			TeamId: teamID,
			UserId: "testUserID",
			State:  "{}",
			Submission: map[string]interface{}{
				app.DialogFieldPlaybookIDKey: "playbookid1",
				app.DialogFieldNameKey:       "incidentName",
			},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(withid, nil).
			Times(1)

		i := app.PlaybookRun{
			OwnerUserID:     dialogRequest.UserId,
			TeamID:          dialogRequest.TeamId,
			Name:            "incidentName",
			PlaybookID:      "playbookid1",
			Description:     "description",
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
		}
		retI := i.Clone()
		retI.ChannelID = "channelID"
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		poster.EXPECT().PublishWebsocketEventToUser(gomock.Any(), gomock.Any(), gomock.Any())
		poster.EXPECT().EphemeralPost(gomock.Any(), gomock.Any(), gomock.Any())
		playbookRunService.EXPECT().CreatePlaybookRun(&i, &withid, "testUserID", true).Return(retI, nil)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents/dialog", bytes.NewBuffer(dialogRequest.ToJson()))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("create valid incident from dialog with description", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		withid := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			Description:             "description",
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		dialogRequest := model.SubmitDialogRequest{
			TeamId: teamID,
			UserId: "testUserID",
			State:  "{}",
			Submission: map[string]interface{}{
				app.DialogFieldPlaybookIDKey: "playbookid1",
				app.DialogFieldNameKey:       "incidentName",
			},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(withid, nil).
			Times(1)

		i := app.PlaybookRun{
			OwnerUserID:     dialogRequest.UserId,
			TeamID:          dialogRequest.TeamId,
			Name:            "incidentName",
			Description:     "description",
			PlaybookID:      withid.ID,
			Checklists:      withid.Checklists,
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
		}
		retI := i
		retI.ChannelID = "channelID"
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		poster.EXPECT().PublishWebsocketEventToUser(gomock.Any(), gomock.Any(), gomock.Any())
		poster.EXPECT().EphemeralPost(gomock.Any(), gomock.Any(), gomock.Any())
		playbookRunService.EXPECT().CreatePlaybookRun(&i, &withid, "testUserID", true).Return(&retI, nil)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents/dialog", bytes.NewBuffer(dialogRequest.ToJson()))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("create incident from dialog - no permissions for public channels", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		withid := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		dialogRequest := model.SubmitDialogRequest{
			TeamId: teamID,
			UserId: "testUserID",
			State:  "{}",
			Submission: map[string]interface{}{
				app.DialogFieldPlaybookIDKey: "playbookid1",
				app.DialogFieldNameKey:       "incidentName",
			},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(withid, nil).
			Times(1)

		i := app.PlaybookRun{
			OwnerUserID: dialogRequest.UserId,
			TeamID:      dialogRequest.TeamId,
			Name:        "incidentName",
			PlaybookID:  withid.ID,
			Checklists:  withid.Checklists,
		}
		retI := i
		retI.ChannelID = "channelID"
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(false)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents/dialog", bytes.NewBuffer(dialogRequest.ToJson()))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var dialogResp model.SubmitDialogResponse
		err = json.NewDecoder(resp.Body).Decode(&dialogResp)
		require.Nil(t, err)

		expectedDialogResp := model.SubmitDialogResponse{
			Errors: map[string]string{
				"incidentName": "You are not able to create a public channel: permissions error",
			},
		}

		require.Equal(t, expectedDialogResp, dialogResp)
	})

	t.Run("create incident from dialog - no permissions for public channels", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		withid := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			CreatePublicPlaybookRun: false,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		dialogRequest := model.SubmitDialogRequest{
			TeamId: teamID,
			UserId: "testUserID",
			State:  "{}",
			Submission: map[string]interface{}{
				app.DialogFieldPlaybookIDKey: "playbookid1",
				app.DialogFieldNameKey:       "incidentName",
			},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(withid, nil).
			Times(1)

		i := app.PlaybookRun{
			OwnerUserID: dialogRequest.UserId,
			TeamID:      dialogRequest.TeamId,
			Name:        "incidentName",
			PlaybookID:  withid.ID,
			Checklists:  withid.Checklists,
		}
		retI := i
		retI.ChannelID = "channelID"
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PRIVATE_CHANNEL).Return(false)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents/dialog", bytes.NewBuffer(dialogRequest.ToJson()))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var dialogResp model.SubmitDialogResponse
		err = json.NewDecoder(resp.Body).Decode(&dialogResp)
		require.Nil(t, err)

		expectedDialogResp := model.SubmitDialogResponse{
			Errors: map[string]string{
				"incidentName": "You are not able to create a private channel: permissions error",
			},
		}

		require.Equal(t, expectedDialogResp, dialogResp)
	})

	t.Run("create incident from dialog - dialog request userID doesn't match requester's id", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		withid := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		dialogRequest := model.SubmitDialogRequest{
			TeamId: teamID,
			UserId: "fakeUserID",
			State:  "{}",
			Submission: map[string]interface{}{
				app.DialogFieldPlaybookIDKey: "playbookid1",
				app.DialogFieldNameKey:       "incidentName",
			},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(withid, nil).
			Times(1)

		i := app.PlaybookRun{
			OwnerUserID: dialogRequest.UserId,
			TeamID:      dialogRequest.TeamId,
			Name:        "incidentName",
			PlaybookID:  withid.ID,
			Checklists:  withid.Checklists,
		}
		retI := i
		retI.ChannelID = "channelID"

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents/dialog", bytes.NewBuffer(dialogRequest.ToJson()))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var res struct{ Error string }
		err = json.NewDecoder(resp.Body).Decode(&res)
		assert.NoError(t, err)
		assert.Equal(t, "interactive dialog's userID must be the same as the requester's userID", res.Error)
	})

	t.Run("create valid incident with missing playbookID from dialog", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		dialogRequest := model.SubmitDialogRequest{
			TeamId: teamID,
			UserId: "testUserID",
			State:  "{}",
			Submission: map[string]interface{}{
				app.DialogFieldPlaybookIDKey: "playbookid1",
				app.DialogFieldNameKey:       "incidentName",
			},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(
				app.Playbook{},
				errors.Wrap(app.ErrNotFound, "playbook does not exist for id 'playbookid1'"),
			).
			Times(1)

		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents/dialog", bytes.NewBuffer(dialogRequest.ToJson()))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("create incident from dialog -- user does not have permission for the original postID's channel", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		withid := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			Description:             "description",
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		dialogRequest := model.SubmitDialogRequest{
			TeamId: teamID,
			UserId: "testUserID",
			State:  `{"post_id": "privatePostID"}`,
			Submission: map[string]interface{}{
				app.DialogFieldPlaybookIDKey: "playbookid1",
				app.DialogFieldNameKey:       "incidentName",
			},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(withid, nil).
			Times(1)

		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		pluginAPI.On("GetPost", "privatePostID").Return(&model.Post{ChannelId: "privateChannelId"}, nil)
		pluginAPI.On("HasPermissionToChannel", "testUserID", "privateChannelId", model.PERMISSION_READ_CHANNEL).Return(false)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents/dialog", bytes.NewBuffer(dialogRequest.ToJson()))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("create incident from dialog -- user is not a member of the playbook", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		withid := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"some_other_id"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		dialogRequest := model.SubmitDialogRequest{
			TeamId: teamID,
			UserId: "testUserID",
			State:  "{}",
			Submission: map[string]interface{}{
				app.DialogFieldPlaybookIDKey: "playbookid1",
				app.DialogFieldNameKey:       "incidentName",
			},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(withid, nil).
			Times(1)

		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		pluginAPI.On("GetPost", "privatePostID").Return(&model.Post{ChannelId: "privateChannelId"}, nil)
		pluginAPI.On("HasPermissionToChannel", "testUserID", "privateChannelId", model.PERMISSION_READ_CHANNEL).Return(false)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("POST", "/api/v0/incidents/dialog", bytes.NewBuffer(dialogRequest.ToJson()))
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("create valid incident", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybook := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			Description:             "description",
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		testPlaybookRun := app.PlaybookRun{
			OwnerUserID:     "testUserID",
			TeamID:          teamID,
			Name:            "incidentName",
			Description:     "description",
			PlaybookID:      testPlaybook.ID,
			Checklists:      testPlaybook.Checklists,
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(testPlaybook, nil).
			Times(1)

		retI := testPlaybookRun
		retI.ID = "incidentID"
		retI.ChannelID = "channelID"
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		playbookRunService.EXPECT().CreatePlaybookRun(&testPlaybookRun, &testPlaybook, "testUserID", true).Return(&retI, nil)

		// Verify that the websocket event is published
		poster.EXPECT().
			PublishWebsocketEventToUser(gomock.Any(), gomock.Any(), gomock.Any())

		resultPlaybookRun, err := c.PlaybookRuns.Create(context.TODO(), icClient.PlaybookRunCreateOptions{
			Name:        testPlaybookRun.Name,
			OwnerUserID: testPlaybookRun.OwnerUserID,
			TeamID:      testPlaybookRun.TeamID,
			Description: testPlaybookRun.Description,
			PlaybookID:  testPlaybookRun.PlaybookID,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resultPlaybookRun.ID)
	})

	t.Run("create valid incident, invite users enabled", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybook := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			Description:             "description",
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      true,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		testPlaybookRun := app.PlaybookRun{
			OwnerUserID:     "testUserID",
			TeamID:          teamID,
			Name:            "incidentName",
			Description:     "description",
			PlaybookID:      testPlaybook.ID,
			Checklists:      testPlaybook.Checklists,
			InvitedUserIDs:  []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs: []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(testPlaybook, nil).
			Times(1)

		retI := testPlaybookRun
		retI.ID = "incidentID"
		retI.ChannelID = "channelID"
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		playbookRunService.EXPECT().CreatePlaybookRun(&testPlaybookRun, &testPlaybook, "testUserID", true).Return(&retI, nil)

		// Verify that the websocket event is published
		poster.EXPECT().
			PublishWebsocketEventToUser(gomock.Any(), gomock.Any(), gomock.Any())

		resultPlaybookRun, err := c.PlaybookRuns.Create(context.TODO(), icClient.PlaybookRunCreateOptions{
			Name:        testPlaybookRun.Name,
			OwnerUserID: testPlaybookRun.OwnerUserID,
			TeamID:      testPlaybookRun.TeamID,
			Description: testPlaybookRun.Description,
			PlaybookID:  testPlaybookRun.PlaybookID,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resultPlaybookRun.ID)
	})

	t.Run("create valid incident without playbook", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
		}

		retI := testPlaybookRun
		retI.ID = "incidentID"
		retI.ChannelID = "channelID"
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		playbookRunService.EXPECT().CreatePlaybookRun(&testPlaybookRun, nil, "testUserID", true).Return(&retI, nil)

		// Verify that the websocket event is published
		poster.EXPECT().
			PublishWebsocketEventToUser(gomock.Any(), gomock.Any(), gomock.Any())

		resultPlaybookRun, err := c.PlaybookRuns.Create(context.TODO(), icClient.PlaybookRunCreateOptions{
			Name:        testPlaybookRun.Name,
			OwnerUserID: testPlaybookRun.OwnerUserID,
			TeamID:      testPlaybookRun.TeamID,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resultPlaybookRun.ID)
	})

	t.Run("create invalid incident - missing owner", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			TeamID: teamID,
			Name:   "incidentName",
		}

		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)

		resultPlaybookRun, err := c.PlaybookRuns.Create(context.TODO(), icClient.PlaybookRunCreateOptions{
			Name:   testPlaybookRun.Name,
			TeamID: testPlaybookRun.TeamID,
		})
		requireErrorWithStatusCode(t, err, http.StatusBadRequest)
		require.Nil(t, resultPlaybookRun)
	})

	t.Run("create invalid incident - missing team", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		testPlaybookRun := app.PlaybookRun{
			OwnerUserID: "testUserID",
			Name:        "incidentName",
		}

		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)

		resultPlaybookRun, err := c.PlaybookRuns.Create(context.TODO(), icClient.PlaybookRunCreateOptions{
			Name:        testPlaybookRun.Name,
			OwnerUserID: testPlaybookRun.OwnerUserID,
		})
		requireErrorWithStatusCode(t, err, http.StatusBadRequest)
		require.Nil(t, resultPlaybookRun)
	})

	t.Run("create invalid incident - missing name", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			OwnerUserID: "testUserID",
			TeamID:      teamID,
		}

		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)

		resultPlaybookRun, err := c.PlaybookRuns.Create(context.TODO(), icClient.PlaybookRunCreateOptions{
			Name:        "",
			TeamID:      testPlaybookRun.TeamID,
			OwnerUserID: testPlaybookRun.OwnerUserID,
		})
		requireErrorWithStatusCode(t, err, http.StatusBadRequest)
		require.Nil(t, resultPlaybookRun)
	})

	t.Run("create incident in unlicensed server with pricing plan differentiation enabled", func(*testing.T) {
		mockCtrl = gomock.NewController(t)
		configService = mock_config.NewMockService(mockCtrl)
		pluginAPI = &plugintest.API{}
		client = pluginapi.NewClient(pluginAPI)
		poster = mock_poster.NewMockPoster(mockCtrl)
		logger = mock_poster.NewMockLogger(mockCtrl)
		handler = NewHandler(client, configService, logger)
		playbookService = mock_app.NewMockPlaybookService(mockCtrl)
		playbookRunService = mock_app.NewMockPlaybookRunService(mockCtrl)
		NewPlaybookRunHandler(handler.APIRouter, playbookRunService, playbookService, client, poster, logger, configService)

		configService.EXPECT().
			IsAtLeastE10Licensed().
			Return(false)

		configService.EXPECT().
			GetConfiguration().
			Return(&config.Configuration{
				EnabledTeams: []string{},
			})

		teamID := model.NewId()
		testPlaybook := app.Playbook{
			ID:                      "playbookid1",
			Title:                   "My Playbook",
			TeamID:                  teamID,
			Description:             "description",
			CreatePublicPlaybookRun: true,
			MemberIDs:               []string{"testUserID"},
			InviteUsersEnabled:      false,
			InvitedUserIDs:          []string{"testInvitedUserID1", "testInvitedUserID2"},
			InvitedGroupIDs:         []string{"testInvitedGroupID1", "testInvitedGroupID2"},
		}

		testPlaybookRun := app.PlaybookRun{
			OwnerUserID:     "testUserID",
			TeamID:          teamID,
			Name:            "incidentName",
			Description:     "description",
			PlaybookID:      testPlaybook.ID,
			Checklists:      testPlaybook.Checklists,
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
		}

		playbookService.EXPECT().
			Get("playbookid1").
			Return(testPlaybook, nil).
			Times(1)

		retI := testPlaybookRun
		retI.ID = "incidentID"
		retI.ChannelID = "channelID"
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_CREATE_PUBLIC_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToTeam", "testUserID", teamID, model.PERMISSION_VIEW_TEAM).Return(true)
		playbookRunService.EXPECT().CreatePlaybookRun(&testPlaybookRun, &testPlaybook, "testUserID", true).Return(&retI, nil)

		// Verify that the websocket event is published
		poster.EXPECT().
			PublishWebsocketEventToUser(gomock.Any(), gomock.Any(), gomock.Any())

		resultPlaybookRun, err := c.PlaybookRuns.Create(context.TODO(), icClient.PlaybookRunCreateOptions{
			Name:        testPlaybookRun.Name,
			OwnerUserID: testPlaybookRun.OwnerUserID,
			TeamID:      testPlaybookRun.TeamID,
			Description: testPlaybookRun.Description,
			PlaybookID:  testPlaybookRun.PlaybookID,
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resultPlaybookRun.ID)
	})

	t.Run("get incident by channel id", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:              "incidentID",
			OwnerUserID:     "testUserID",
			TeamID:          teamID,
			Name:            "incidentName",
			ChannelID:       "channelID",
			Checklists:      []app.Checklist{},
			StatusPosts:     []app.StatusPost{},
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
			TimelineEvents:  []app.TimelineEvent{},
		}

		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel("channelID").Return("incidentID", nil)
		playbookRunService.EXPECT().GetPlaybookRun("incidentID").Return(&testPlaybookRun, nil)

		resultPlaybookRun, err := c.PlaybookRuns.GetByChannelID(context.TODO(), testPlaybookRun.ChannelID)
		require.NoError(t, err)
		assert.Equal(t, testPlaybookRun, toInternalPlaybookRun(*resultPlaybookRun))
	})

	t.Run("get incident by channel id - not found", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())
		userID := "testUserID"

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel("channelID").Return("", app.ErrNotFound)
		logger.EXPECT().Warnf("User %s does not have permissions to get incident for channel %s", userID, testPlaybookRun.ChannelID)

		resultPlaybookRun, err := c.PlaybookRuns.GetByChannelID(context.TODO(), testPlaybookRun.ChannelID)
		requireErrorWithStatusCode(t, err, http.StatusNotFound)
		require.Nil(t, resultPlaybookRun)
	})

	t.Run("get incident by channel id - not authorized", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(false)
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{}, nil)
		playbookRunService.EXPECT().GetPlaybookRunIDForChannel(testPlaybookRun.ChannelID).Return(testPlaybookRun.ID, nil)
		playbookRunService.EXPECT().GetPlaybookRun(testPlaybookRun.ID).Return(&testPlaybookRun, nil)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		resultPlaybookRun, err := c.PlaybookRuns.GetByChannelID(context.TODO(), testPlaybookRun.ChannelID)
		requireErrorWithStatusCode(t, err, http.StatusNotFound)
		require.Nil(t, resultPlaybookRun)
	})

	t.Run("get private incident - not part of channel", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
			PostID:      "",
			PlaybookID:  "",
			Checklists:  nil,
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_PRIVATE}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(false)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil)

		resultPlaybookRun, err := c.PlaybookRuns.Get(context.TODO(), testPlaybookRun.ID)
		requireErrorWithStatusCode(t, err, http.StatusForbidden)
		require.Nil(t, resultPlaybookRun)
	})

	t.Run("get private incident - part of channel", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:              "incidentID",
			OwnerUserID:     "testUserID",
			TeamID:          teamID,
			Name:            "incidentName",
			ChannelID:       "channelID",
			PostID:          "",
			PlaybookID:      "",
			Checklists:      []app.Checklist{},
			StatusPosts:     []app.StatusPost{},
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
			TimelineEvents:  []app.TimelineEvent{},
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_PRIVATE}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(true)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil).Times(2)

		resultPlaybookRun, err := c.PlaybookRuns.Get(context.TODO(), testPlaybookRun.ID)
		require.NoError(t, err)
		assert.Equal(t, testPlaybookRun, toInternalPlaybookRun(*resultPlaybookRun))
	})

	t.Run("get public incident - not part of channel or team", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:              "incidentID",
			OwnerUserID:     "testUserID",
			TeamID:          teamID,
			Name:            "incidentName",
			ChannelID:       "channelID",
			PostID:          "",
			PlaybookID:      "",
			Checklists:      nil,
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_OPEN, TeamId: testPlaybookRun.TeamID}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(false)
		pluginAPI.On("HasPermissionToTeam", "testUserID", testPlaybookRun.TeamID, model.PERMISSION_LIST_TEAM_CHANNELS).
			Return(false)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil)

		resultPlaybookRun, err := c.PlaybookRuns.Get(context.TODO(), testPlaybookRun.ID)
		requireErrorWithStatusCode(t, err, http.StatusForbidden)
		require.Nil(t, resultPlaybookRun)
	})

	t.Run("get public incident - not part of channel, but part of team", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:              "incidentID",
			OwnerUserID:     "testUserID",
			TeamID:          teamID,
			Name:            "incidentName",
			ChannelID:       "channelID",
			PostID:          "",
			PlaybookID:      "",
			Checklists:      []app.Checklist{},
			StatusPosts:     []app.StatusPost{},
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
			TimelineEvents:  []app.TimelineEvent{},
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_OPEN, TeamId: testPlaybookRun.TeamID}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(false)
		pluginAPI.On("HasPermissionToTeam", "testUserID", testPlaybookRun.TeamID, model.PERMISSION_LIST_TEAM_CHANNELS).
			Return(true)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil).Times(2)

		resultPlaybookRun, err := c.PlaybookRuns.Get(context.TODO(), testPlaybookRun.ID)
		require.NoError(t, err)
		assert.Equal(t, testPlaybookRun, toInternalPlaybookRun(*resultPlaybookRun))
	})

	t.Run("get public incident - part of channel", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:              "incidentID",
			OwnerUserID:     "testUserID",
			TeamID:          teamID,
			Name:            "incidentName",
			ChannelID:       "channelID",
			PostID:          "",
			PlaybookID:      "",
			Checklists:      []app.Checklist{},
			StatusPosts:     []app.StatusPost{},
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
			TimelineEvents:  []app.TimelineEvent{},
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_OPEN, TeamId: testPlaybookRun.TeamID}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(true)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil).Times(2)

		resultPlaybookRun, err := c.PlaybookRuns.Get(context.TODO(), testPlaybookRun.ID)
		require.NoError(t, err)
		assert.Equal(t, testPlaybookRun, toInternalPlaybookRun(*resultPlaybookRun))
	})

	t.Run("get private incident metadata - not part of channel", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
			PostID:      "",
			PlaybookID:  "",
			Checklists:  nil,
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_PRIVATE}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(false)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil)

		resultPlaybookRunMetadata, err := c.PlaybookRuns.GetMetadata(context.TODO(), testPlaybookRun.ID)
		requireErrorWithStatusCode(t, err, http.StatusForbidden)
		require.Nil(t, resultPlaybookRunMetadata)
	})

	t.Run("get private incident metadata - part of channel", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
			PostID:      "",
			PlaybookID:  "",
			Checklists:  []app.Checklist{},
		}

		testPlaybookRunMetadata := app.Metadata{
			ChannelName:        "theChannelName",
			ChannelDisplayName: "theChannelDisplayName",
			TeamName:           "ourAwesomeTeam",
			NumMembers:         11,
			TotalPosts:         42,
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_PRIVATE}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(true)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil)

		playbookRunService.EXPECT().
			GetPlaybookRunMetadata("incidentID").
			Return(&testPlaybookRunMetadata, nil)

		resultPlaybookRunMetadata, err := c.PlaybookRuns.GetMetadata(context.TODO(), testPlaybookRun.ID)
		require.NoError(t, err)
		assert.Equal(t, testPlaybookRunMetadata, toInternalPlaybookRunMetadata(*resultPlaybookRunMetadata))
	})

	t.Run("get public incident metadata - not part of channel or team", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
			PostID:      "",
			PlaybookID:  "",
			Checklists:  nil,
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_OPEN, TeamId: testPlaybookRun.TeamID}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(false)
		pluginAPI.On("HasPermissionToTeam", "testUserID", testPlaybookRun.TeamID, model.PERMISSION_LIST_TEAM_CHANNELS).
			Return(false)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil)

		resultPlaybookRunMetadata, err := c.PlaybookRuns.GetMetadata(context.TODO(), testPlaybookRun.ID)
		requireErrorWithStatusCode(t, err, http.StatusForbidden)
		require.Nil(t, resultPlaybookRunMetadata)
	})

	t.Run("get public incident metadata - not part of channel, but part of team", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
			PostID:      "",
			PlaybookID:  "",
			Checklists:  []app.Checklist{},
		}

		testPlaybookRunMetadata := app.Metadata{
			ChannelName:        "theChannelName",
			ChannelDisplayName: "theChannelDisplayName",
			TeamName:           "ourAwesomeTeam",
			NumMembers:         11,
			TotalPosts:         42,
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_OPEN, TeamId: testPlaybookRun.TeamID}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(false)
		pluginAPI.On("HasPermissionToTeam", "testUserID", testPlaybookRun.TeamID, model.PERMISSION_LIST_TEAM_CHANNELS).
			Return(true)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil)

		playbookRunService.EXPECT().
			GetPlaybookRunMetadata("incidentID").
			Return(&testPlaybookRunMetadata, nil)

		resultPlaybookRunMetadata, err := c.PlaybookRuns.GetMetadata(context.TODO(), testPlaybookRun.ID)
		require.NoError(t, err)
		assert.Equal(t, testPlaybookRunMetadata, toInternalPlaybookRunMetadata(*resultPlaybookRunMetadata))
	})

	t.Run("get public incident metadata - part of channel", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
			PostID:      "",
			PlaybookID:  "",
			Checklists:  []app.Checklist{},
		}

		testPlaybookRunMetadata := app.Metadata{
			ChannelName:        "theChannelName",
			ChannelDisplayName: "theChannelDisplayName",
			TeamName:           "ourAwesomeTeam",
			NumMembers:         11,
			TotalPosts:         42,
		}

		pluginAPI.On("GetChannel", testPlaybookRun.ChannelID).
			Return(&model.Channel{Type: model.CHANNEL_OPEN, TeamId: testPlaybookRun.TeamID}, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", "testUserID", testPlaybookRun.ChannelID, model.PERMISSION_READ_CHANNEL).
			Return(true)

		logger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		playbookRunService.EXPECT().
			GetPlaybookRun("incidentID").
			Return(&testPlaybookRun, nil)

		playbookRunService.EXPECT().
			GetPlaybookRunMetadata("incidentID").
			Return(&testPlaybookRunMetadata, nil)

		resultPlaybookRunMetadata, err := c.PlaybookRuns.GetMetadata(context.TODO(), testPlaybookRun.ID)
		require.NoError(t, err)
		assert.Equal(t, testPlaybookRunMetadata, toInternalPlaybookRunMetadata(*resultPlaybookRunMetadata))
	})

	t.Run("get incidents", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		playbookRun1 := app.PlaybookRun{
			ID:              "incidentID1",
			OwnerUserID:     "testUserID1",
			TeamID:          teamID,
			Name:            "incidentName1",
			ChannelID:       "channelID1",
			Checklists:      []app.Checklist{},
			StatusPosts:     []app.StatusPost{},
			InvitedUserIDs:  []string{},
			InvitedGroupIDs: []string{},
			TimelineEvents:  []app.TimelineEvent{},
		}

		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("GetUser", "testUserID").Return(&model.User{}, nil)
		pluginAPI.On("HasPermissionToTeam", mock.Anything, mock.Anything, model.PERMISSION_VIEW_TEAM).Return(true)
		result := &app.GetPlaybookRunsResults{
			TotalCount: 100,
			PageCount:  200,
			HasMore:    true,
			Items:      []app.PlaybookRun{playbookRun1},
		}
		playbookRunService.EXPECT().GetPlaybookRuns(gomock.Any(), gomock.Any()).Return(result, nil)

		actualList, err := c.PlaybookRuns.List(context.TODO(), 0, 200, icClient.PlaybookRunListOptions{
			TeamID: teamID,
		})
		require.NoError(t, err)

		expectedList := &icClient.GetPlaybookRunsResults{
			TotalCount: 100,
			PageCount:  200,
			HasMore:    true,
			Items:      []icClient.PlaybookRun{toAPIPlaybookRun(playbookRun1)},
		}
		assert.Equal(t, expectedList, actualList)
	})

	t.Run("get empty list of incidents", func(t *testing.T) {
		reset(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())
		setDefaultExpectations(t)

		teamID := model.NewId()
		pluginAPI.On("HasPermissionToTeam", mock.Anything, teamID, model.PERMISSION_VIEW_TEAM).Return(false)

		resultPlaybookRun, err := c.PlaybookRuns.List(context.TODO(), 0, 100, icClient.PlaybookRunListOptions{
			TeamID: teamID,
		})
		requireErrorWithStatusCode(t, err, http.StatusForbidden)
		require.Nil(t, resultPlaybookRun)
	})

	t.Run("get disabled list of incidents", func(t *testing.T) {
		reset(t)

		disabledTeamID := model.NewId()
		enabledTeamID := model.NewId()
		configService.EXPECT().
			GetConfiguration().
			Return(&config.Configuration{
				EnabledTeams: []string{enabledTeamID},
			})

		setDefaultExpectations(t)

		pluginAPI.On("HasPermissionToTeam", mock.Anything, disabledTeamID, model.PERMISSION_VIEW_TEAM).Return(true)

		actualList, err := c.PlaybookRuns.List(context.TODO(), 0, 100, icClient.PlaybookRunListOptions{
			TeamID: disabledTeamID,
		})
		require.NoError(t, err)

		expectedList := &icClient.GetPlaybookRunsResults{
			Disabled: true,
		}
		assert.Equal(t, expectedList, actualList)
	})

	t.Run("checklist autocomplete for a channel without permission to view", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel(testPlaybookRun.ChannelID).Return(testPlaybookRun.ID, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		playbookRunService.EXPECT().GetPlaybookRun(testPlaybookRun.ID).Return(&testPlaybookRun, nil)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(false)
		pluginAPI.On("GetChannel", mock.Anything).Return(&model.Channel{Type: model.CHANNEL_PRIVATE}, nil)

		testrecorder := httptest.NewRecorder()
		testreq, err := http.NewRequest("GET", "/api/v0/incidents/checklist-autocomplete?channel_id="+testPlaybookRun.ChannelID, nil)
		testreq.Header.Add("Mattermost-User-ID", "testUserID")
		require.NoError(t, err)
		handler.ServeHTTP(testrecorder, testreq)

		resp := testrecorder.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("update incident status", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel(testPlaybookRun.ChannelID).Return(testPlaybookRun.ID, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		playbookRunService.EXPECT().GetPlaybookRun(testPlaybookRun.ID).Return(&testPlaybookRun, nil).Times(2)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_CREATE_POST).Return(true)

		updateOptions := app.StatusUpdateOptions{
			Status:      "Active",
			Message:     "test message",
			Description: "test description",
			Reminder:    600 * time.Second,
		}
		playbookRunService.EXPECT().UpdateStatus("incidentID", "testUserID", updateOptions).Return(nil)

		err := c.PlaybookRuns.UpdateStatus(context.TODO(), "incidentID", icClient.StatusActive, "test description", "test message", 600)
		require.NoError(t, err)
	})

	t.Run("update incident status, bad status", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel(testPlaybookRun.ChannelID).Return(testPlaybookRun.ID, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		playbookRunService.EXPECT().GetPlaybookRun(testPlaybookRun.ID).Return(&testPlaybookRun, nil).Times(2)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_CREATE_POST).Return(true)

		err := c.PlaybookRuns.UpdateStatus(context.TODO(), "incidentID", "Arrrrrrrctive", "test description", "test message", 600)
		requireErrorWithStatusCode(t, err, http.StatusBadRequest)
	})

	t.Run("update incident status, no permission to post", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel(testPlaybookRun.ChannelID).Return(testPlaybookRun.ID, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		playbookRunService.EXPECT().GetPlaybookRun(testPlaybookRun.ID).Return(&testPlaybookRun, nil).Times(2)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_CREATE_POST).Return(false)

		err := c.PlaybookRuns.UpdateStatus(context.TODO(), "incidentID", icClient.StatusActive, "test description", "test message", 600)
		requireErrorWithStatusCode(t, err, http.StatusForbidden)
	})

	t.Run("update incident status, message empty", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel(testPlaybookRun.ChannelID).Return(testPlaybookRun.ID, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		playbookRunService.EXPECT().GetPlaybookRun(testPlaybookRun.ID).Return(&testPlaybookRun, nil).Times(2)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_CREATE_POST).Return(true)

		err := c.PlaybookRuns.UpdateStatus(context.TODO(), "incidentID", icClient.StatusActive, "test description", "  \t   \r   \t  \r\r  ", 600)
		requireErrorWithStatusCode(t, err, http.StatusBadRequest)
	})

	t.Run("update incident status, status empty", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel(testPlaybookRun.ChannelID).Return(testPlaybookRun.ID, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		playbookRunService.EXPECT().GetPlaybookRun(testPlaybookRun.ID).Return(&testPlaybookRun, nil).Times(2)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_CREATE_POST).Return(true)

		err := c.PlaybookRuns.UpdateStatus(context.TODO(), "incidentID", "\t   \r  ", "test description", "test message", 600)
		requireErrorWithStatusCode(t, err, http.StatusBadRequest)
	})

	t.Run("update incident status, description empty", func(t *testing.T) {
		reset(t)
		setDefaultExpectations(t)
		logger.EXPECT().Warnf(gomock.Any(), gomock.Any(), gomock.Any())

		teamID := model.NewId()
		testPlaybookRun := app.PlaybookRun{
			ID:          "incidentID",
			OwnerUserID: "testUserID",
			TeamID:      teamID,
			Name:        "incidentName",
			ChannelID:   "channelID",
		}

		playbookRunService.EXPECT().GetPlaybookRunIDForChannel(testPlaybookRun.ChannelID).Return(testPlaybookRun.ID, nil)
		pluginAPI.On("HasPermissionTo", mock.Anything, model.PERMISSION_MANAGE_SYSTEM).Return(false)
		playbookRunService.EXPECT().GetPlaybookRun(testPlaybookRun.ID).Return(&testPlaybookRun, nil).Times(2)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_READ_CHANNEL).Return(true)
		pluginAPI.On("HasPermissionToChannel", mock.Anything, mock.Anything, model.PERMISSION_CREATE_POST).Return(true)

		err := c.PlaybookRuns.UpdateStatus(context.TODO(), "incidentID", "Active", "  \r \n  ", "test message", 600)
		requireErrorWithStatusCode(t, err, http.StatusBadRequest)
	})
}
