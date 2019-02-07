// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	api "code.gitea.io/sdk/gitea"
	"github.com/stretchr/testify/assert"
)

type APITestContext struct {
	Reponame     string
	Session      *TestSession
	Token        string
	Username     string
	ExpectedCode int
}

func NewAPITestContext(t *testing.T, username, reponame string) APITestContext {
	session := loginUser(t, username)
	token := getTokenForLoggedInUser(t, session)
	return APITestContext{
		Session:  session,
		Token:    token,
		Username: username,
		Reponame: reponame,
	}
}

func (ctx APITestContext) GitPath() string {
	return fmt.Sprintf("%s/%s.git", ctx.Username, ctx.Reponame)
}

func doAPICreateRepository(ctx APITestContext, empty bool, callback ...func(*testing.T, api.Repository)) func(*testing.T) {
	return func(t *testing.T) {
		createRepoOption := &api.CreateRepoOption{
			AutoInit:    !empty,
			Description: "Temporary repo",
			Name:        ctx.Reponame,
			Private:     true,
			Gitignores:  "",
			License:     "WTFPL",
			Readme:      "Default",
		}
		req := NewRequestWithJSON(t, "POST", "/api/v1/user/repos?token="+ctx.Token, createRepoOption)
		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		resp := ctx.Session.MakeRequest(t, req, http.StatusCreated)

		var repository api.Repository
		DecodeJSON(t, resp, &repository)
		if len(callback) > 0 {
			callback[0](t, repository)
		}
	}
}

func doAPIGetRepository(ctx APITestContext, callback ...func(*testing.T, api.Repository)) func(*testing.T) {
	return func(t *testing.T) {
		urlStr := fmt.Sprintf("/api/v1/repos/%s/%s?token=%s", ctx.Username, ctx.Reponame, ctx.Token)

		req := NewRequest(t, "GET", urlStr)
		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		resp := ctx.Session.MakeRequest(t, req, http.StatusOK)

		var repository api.Repository
		DecodeJSON(t, resp, &repository)
		if len(callback) > 0 {
			callback[0](t, repository)
		}
	}
}

func doAPIDeleteRepository(ctx APITestContext) func(*testing.T) {
	return func(t *testing.T) {
		urlStr := fmt.Sprintf("/api/v1/repos/%s/%s?token=%s", ctx.Username, ctx.Reponame, ctx.Token)

		req := NewRequest(t, "DELETE", urlStr)
		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		ctx.Session.MakeRequest(t, req, http.StatusNoContent)
	}
}

func doAPICreateUserKey(ctx APITestContext, keyname, keyFile string, callback ...func(*testing.T, api.PublicKey)) func(*testing.T) {
	return func(t *testing.T) {
		urlStr := fmt.Sprintf("/api/v1/user/keys?token=%s", ctx.Token)

		dataPubKey, err := ioutil.ReadFile(keyFile + ".pub")
		assert.NoError(t, err)
		req := NewRequestWithJSON(t, "POST", urlStr, &api.CreateKeyOption{
			Title: keyname,
			Key:   string(dataPubKey),
		})
		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		resp := ctx.Session.MakeRequest(t, req, http.StatusCreated)
		var publicKey api.PublicKey
		DecodeJSON(t, resp, &publicKey)
		if len(callback) > 0 {
			callback[0](t, publicKey)
		}
	}
}

func doAPIDeleteUserKey(ctx APITestContext, keyID int64) func(*testing.T) {
	return func(t *testing.T) {
		urlStr := fmt.Sprintf("/api/v1/user/keys/%d?token=%s", keyID, ctx.Token)

		req := NewRequest(t, "DELETE", urlStr)
		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		ctx.Session.MakeRequest(t, req, http.StatusNoContent)
	}
}

func doAPICreateDeployKey(ctx APITestContext, keyname, keyFile string, readOnly bool) func(*testing.T) {
	return func(t *testing.T) {
		urlStr := fmt.Sprintf("/api/v1/repos/%s/%s/keys?token=%s", ctx.Username, ctx.Reponame, ctx.Token)

		dataPubKey, err := ioutil.ReadFile(keyFile + ".pub")
		assert.NoError(t, err)
		req := NewRequestWithJSON(t, "POST", urlStr, api.CreateKeyOption{
			Title:    keyname,
			Key:      string(dataPubKey),
			ReadOnly: readOnly,
		})

		if ctx.ExpectedCode != 0 {
			ctx.Session.MakeRequest(t, req, ctx.ExpectedCode)
			return
		}
		ctx.Session.MakeRequest(t, req, http.StatusCreated)
	}
}
