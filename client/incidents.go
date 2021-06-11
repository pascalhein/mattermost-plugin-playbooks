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
func (s *PlaybookRunService) Get(ctx context.Context, incidentID string) (*PlaybookRun, error) {
	incidentURL := fmt.Sprintf("incidents/%s", incidentID)
	req, err := s.client.newRequest(http.MethodGet, incidentURL, nil)
	if err != nil {
		return nil, err
	}

	incident := new(PlaybookRun)
	resp, err := s.client.do(ctx, req, incident)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return incident, nil
}

// GetByChannelID gets an incident by ChannelID.
func (s *PlaybookRunService) GetByChannelID(ctx context.Context, channelID string) (*PlaybookRun, error) {
	channelURL := fmt.Sprintf("incidents/channel/%s", channelID)
	req, err := s.client.newRequest(http.MethodGet, channelURL, nil)
	if err != nil {
		return nil, err
	}

	incident := new(PlaybookRun)
	resp, err := s.client.do(ctx, req, incident)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return incident, nil
}

// Get an incident's metadata.
func (s *PlaybookRunService) GetMetadata(ctx context.Context, incidentID string) (*PlaybookRunMetadata, error) {
	incidentURL := fmt.Sprintf("incidents/%s/metadata", incidentID)
	req, err := s.client.newRequest(http.MethodGet, incidentURL, nil)
	if err != nil {
		return nil, err
	}

	incident := new(PlaybookRunMetadata)
	resp, err := s.client.do(ctx, req, incident)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return incident, nil
}

// List the incidents.
func (s *PlaybookRunService) List(ctx context.Context, page, perPage int, opts PlaybookRunListOptions) (*GetPlaybookRunsResults, error) {
	incidentURL := "incidents"
	incidentURL, err := addOptions(incidentURL, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build options: %w", err)
	}
	incidentURL, err = addPaginationOptions(incidentURL, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to build pagination options: %w", err)
	}

	req, err := s.client.newRequest(http.MethodGet, incidentURL, nil)
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
	incidentURL := "incidents"
	req, err := s.client.newRequest(http.MethodPost, incidentURL, opts)
	if err != nil {
		return nil, err
	}

	incident := new(PlaybookRun)
	resp, err := s.client.do(ctx, req, incident)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("expected status code %d", http.StatusCreated)
	}

	return incident, nil
}

func (s *PlaybookRunService) UpdateStatus(ctx context.Context, incidentID string, status Status, description, message string, reminderInSeconds int64) error {
	updateURL := fmt.Sprintf("incidents/%s/status", incidentID)
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
