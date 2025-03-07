// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package harness

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/drone/go-scm/scm"
)

type repositoryService struct {
	client *wrapper
}

func (s *repositoryService) Find(ctx context.Context, repo string) (*scm.Repository, *scm.Response, error) {
	repo = buildHarnessURI(s.client.account, s.client.organization, s.client.project, repo)
	path := fmt.Sprintf("api/v1/repos/%s", repo)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	if err != nil {
		return nil, res, err
	}
	convertedRepo := convertRepository(out)
	if convertedRepo == nil {
		return nil, res, errors.New("Harness returned an unexpected null repository")
	}
	return convertedRepo, res, err
}

func (s *repositoryService) FindHook(ctx context.Context, repo string, id string) (*scm.Hook, *scm.Response, error) {
	harnessURI := buildHarnessURI(s.client.account, s.client.organization, s.client.project, repo)
	path := fmt.Sprintf("api/v1/repos/%s/webhooks/%s", harnessURI, id)
	out := new(hook)
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertHook(out), res, err
}

func (s *repositoryService) FindPerms(ctx context.Context, repo string) (*scm.Perm, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) List(ctx context.Context, opts scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	harnessURI := buildHarnessURI(s.client.account, s.client.organization, s.client.project, "")
	path := fmt.Sprintf("api/v1/spaces/%s/repos?sort=path&order=asc&%s", harnessURI, encodeListOptions(opts))
	out := []*repository{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertRepositoryList(out), res, err
}

func (s *repositoryService) ListHooks(ctx context.Context, repo string, opts scm.ListOptions) ([]*scm.Hook, *scm.Response, error) {
	harnessURI := buildHarnessURI(s.client.account, s.client.organization, s.client.project, repo)
	path := fmt.Sprintf("api/v1/repos/%s/webhooks?sort=display_name&order=asc&%s", harnessURI, encodeListOptions(opts))
	out := []*hook{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertHookList(out), res, err
}

func (s *repositoryService) ListStatus(ctx context.Context, repo string, ref string, opts scm.ListOptions) ([]*scm.Status, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) CreateHook(ctx context.Context, repo string, input *scm.HookInput) (*scm.Hook, *scm.Response, error) {
	harnessURI := buildHarnessURI(s.client.account, s.client.organization, s.client.project, repo)
	path := fmt.Sprintf("api/v1/repos/%s/webhooks", harnessURI)
	in := new(hook)
	in.Enabled = true
	in.DisplayName = input.Name
	in.Secret = input.Secret
	in.Insecure = input.SkipVerify
	in.URL = input.Target
	in.Triggers = append(
		input.NativeEvents,
	)
	out := new(hook)
	res, err := s.client.do(ctx, "POST", path, in, out)
	return convertHook(out), res, err
}

func (s *repositoryService) CreateStatus(ctx context.Context, repo string, ref string, input *scm.StatusInput) (*scm.Status, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) UpdateHook(ctx context.Context, repo, id string, input *scm.HookInput) (*scm.Hook, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) DeleteHook(ctx context.Context, repo string, id string) (*scm.Response, error) {
	harnessURI := buildHarnessURI(s.client.account, s.client.organization, s.client.project, repo)
	path := fmt.Sprintf("api/v1/repos/%s/webhooks/%s", harnessURI, id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

//
// native data structures
//

type (
	// harness repository resource.
	repository struct {
		ID             int    `json:"id"`
		ParentID       int    `json:"parent_id"`
		UID            string `json:"uid"`
		Path           string `json:"path"`
		Description    string `json:"description"`
		IsPublic       bool   `json:"is_public"`
		CreatedBy      int    `json:"created_by"`
		Created        int64  `json:"created"`
		Updated        int64  `json:"updated"`
		DefaultBranch  string `json:"default_branch"`
		ForkID         int    `json:"fork_id"`
		NumForks       int    `json:"num_forks"`
		NumPulls       int    `json:"num_pulls"`
		NumClosedPulls int    `json:"num_closed_pulls"`
		NumOpenPulls   int    `json:"num_open_pulls"`
		NumMergedPulls int    `json:"num_merged_pulls"`
		GitURL         string `json:"git_url"`
	}
	hook struct {
		Created               int      `json:"created"`
		CreatedBy             int      `json:"created_by"`
		Description           string   `json:"description"`
		DisplayName           string   `json:"display_name"`
		Enabled               bool     `json:"enabled"`
		HasSecret             bool     `json:"has_secret"`
		Secret                string   `json:"secret"`
		ID                    int      `json:"id"`
		Insecure              bool     `json:"insecure"`
		LatestExecutionResult string   `json:"latest_execution_result"`
		ParentID              int      `json:"parent_id"`
		ParentType            string   `json:"parent_type"`
		Triggers              []string `json:"triggers"`
		Updated               int      `json:"updated"`
		URL                   string   `json:"url"`
		Version               int      `json:"version"`
	}
)

//
// native data structure conversion
//

func convertRepositoryList(src []*repository) []*scm.Repository {
	var dst []*scm.Repository
	for _, v := range src {
		dst = append(dst, convertRepository(v))
	}
	return dst
}

func convertRepository(src *repository) *scm.Repository {
	return &scm.Repository{
		ID:        strconv.Itoa(src.ID),
		Namespace: src.Path,
		Name:      src.UID,
		Branch:    src.DefaultBranch,
		Private:   !src.IsPublic,
		Clone:     src.GitURL,
		CloneSSH:  src.GitURL,
		Link:      src.GitURL,
		// Created:   time.Unix(src.Created, 0),
		//		Updated:   time.Unix(src.Updated, 0),
	}
}

func convertHookList(from []*hook) []*scm.Hook {
	to := []*scm.Hook{}
	for _, v := range from {
		to = append(to, convertHook(v))
	}
	return to
}

func convertHook(from *hook) *scm.Hook {
	return &scm.Hook{
		ID:         strconv.Itoa(from.ID),
		Name:       from.DisplayName,
		Active:     from.Enabled,
		Target:     from.URL,
		Events:     from.Triggers,
		SkipVerify: from.Insecure,
	}
}
