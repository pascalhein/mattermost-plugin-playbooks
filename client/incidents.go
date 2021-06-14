// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package client

import (
	"context"
	"fmt"
	"net/http"
)

// PlaybookRunService handles communication with the incident related
// methods of the Incident Collaboration API.
type PlaybookRunService struct {
	client *Client
}

// Get an incident.
func (s *PlaybookRunService) Get(ctx context.Context, playbookRunID string) (*PlaybookRun, error) {
	playbookRunURL := fmt.Sprintf("incidents/%s", playbookRunID)
	req, err := s.client.newRequest(http.MethodGet, playbookRunURL, nil)
	if err != nil {
		return nil, err
	}

	playbookRun := new(PlaybookRun)
	resp, err := s.client.do(ctx, req, playbookRun)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return playbookRun, nil
}

// GetByChannelID gets an incident by ChannelID.
func (s *PlaybookRunService) GetByChannelID(ctx context.Context, channelID string) (*PlaybookRun, error) {
	channelURL := fmt.Sprintf("incidents/channel/%s", channelID)
	req, err := s.client.newRequest(http.MethodGet, channelURL, nil)
	if err != nil {
		return nil, err
	}

	playbookRun := new(PlaybookRun)
	resp, err := s.client.do(ctx, req, playbookRun)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return playbookRun, nil
}

// Get an incident's metadata.
func (s *PlaybookRunService) GetMetadata(ctx context.Context, playbookRunID string) (*PlaybookRunMetadata, error) {
	playbookRunURL := fmt.Sprintf("incidents/%s/metadata", playbookRunID)
	req, err := s.client.newRequest(http.MethodGet, playbookRunURL, nil)
	if err != nil {
		return nil, err
	}

	playbookRun := new(PlaybookRunMetadata)
	resp, err := s.client.do(ctx, req, playbookRun)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return playbookRun, nil
}

// List the incidents.
func (s *PlaybookRunService) List(ctx context.Context, page, perPage int, opts PlaybookRunListOptions) (*GetPlaybookRunsResults, error) {
	playbookRunURL := "incidents"
	playbookRunURL, err := addOptions(playbookRunURL, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build options: %w", err)
	}
	playbookRunURL, err = addPaginationOptions(playbookRunURL, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to build pagination options: %w", err)
	}

	req, err := s.client.newRequest(http.MethodGet, playbookRunURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	result := &GetPlaybookRunsResults{}
	resp, err := s.client.do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	resp.Body.Close()

	return result, nil
}

// Create an incident.
func (s *PlaybookRunService) Create(ctx context.Context, opts PlaybookRunCreateOptions) (*PlaybookRun, error) {
	playbookRunURL := "incidents"
	req, err := s.client.newRequest(http.MethodPost, playbookRunURL, opts)
	if err != nil {
		return nil, err
	}

	playbookRun := new(PlaybookRun)
	resp, err := s.client.do(ctx, req, playbookRun)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("expected status code %d", http.StatusCreated)
	}

	return playbookRun, nil
}

func (s *PlaybookRunService) UpdateStatus(ctx context.Context, playbookRunID string, status Status, description, message string, reminderInSeconds int64) error {
	updateURL := fmt.Sprintf("incidents/%s/status", playbookRunID)
	opts := StatusUpdateOptions{
		Status:            status,
		Description:       description,
		Message:           message,
		ReminderInSeconds: reminderInSeconds,
	}
	req, err := s.client.newRequest(http.MethodPost, updateURL, opts)
	if err != nil {
		return err
	}

	_, err = s.client.do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
