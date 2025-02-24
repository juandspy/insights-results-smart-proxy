// Copyright 2020 Red Hat, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/insights-results-smart-proxy/tests/helpers"

	"github.com/RedHatInsights/insights-results-smart-proxy/server"
	types "github.com/RedHatInsights/insights-results-types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name             string
	identity         string
	expectedError    string
	expectedIdentity types.Identity
	expectedUserID   types.UserID
	expectedOrgID    types.OrgID
}

var (
	validIdentityXRH = types.Identity{
		AccountNumber: types.UserID("1"),
		OrgID:         1,
		User: types.User{
			UserID: types.UserID("1"),
		},
	}
)

func TestGetAuthToken(t *testing.T) {
	testCases := []testCase{
		{
			name:             "valid token",
			identity:         "valid",
			expectedError:    "",
			expectedIdentity: validIdentityXRH,
		},
		{
			name:          "no token",
			identity:      "empty",
			expectedError: "token is not provided",
		},
		{
			name:          "invalid token",
			identity:      "bad",
			expectedError: "contextKeyUser has wrong type",
		},
	}

	testServer := server.HTTPServer{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := getRequest(t, tc.identity)

			identity, err := testServer.GetAuthToken(req)
			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, &tc.expectedIdentity, identity)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestGetCurrentUserID(t *testing.T) {
	testCases := []testCase{
		{
			name:           "valid token",
			identity:       "valid",
			expectedError:  "",
			expectedUserID: validIdentityXRH.User.UserID,
		},
		{
			name:          "no token",
			identity:      "empty",
			expectedError: "token is not provided",
		},
		{
			name:          "invalid token",
			identity:      "bad",
			expectedError: "contextKeyUser has wrong type",
		},
	}

	testServer := server.HTTPServer{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := getRequest(t, tc.identity)

			userID, err := testServer.GetCurrentUserID(req)
			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedUserID, userID)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestGetCurrentOrgID(t *testing.T) {
	testCases := []testCase{
		{
			name:          "valid token",
			identity:      "valid",
			expectedError: "",
			expectedOrgID: validIdentityXRH.OrgID,
		},
		{
			name:          "no token",
			identity:      "empty",
			expectedError: "token is not provided",
		},
		{
			name:          "invalid token",
			identity:      "bad",
			expectedError: "contextKeyUser has wrong type",
		},
	}

	testServer := server.HTTPServer{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := getRequest(t, tc.identity)

			orgID, err := testServer.GetCurrentOrgID(req)
			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedOrgID, orgID)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestGetCurrentOrgIDUserIDFromToken(t *testing.T) {
	testCases := []testCase{
		{
			name:           "valid token",
			identity:       "valid",
			expectedError:  "",
			expectedOrgID:  validIdentityXRH.OrgID,
			expectedUserID: validIdentityXRH.User.UserID,
		},
		{
			name:          "no token",
			identity:      "empty",
			expectedError: "token is not provided",
		},
		{
			name:          "invalid token",
			identity:      "bad",
			expectedError: "contextKeyUser has wrong type",
		},
	}

	testServer := server.HTTPServer{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := getRequest(t, tc.identity)

			userID, err := testServer.GetCurrentUserID(req)
			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedUserID, userID)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}

			orgID, err := testServer.GetCurrentOrgID(req)
			if tc.expectedError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedOrgID, orgID)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestGetAuthTokenJWTHeaderMalformed(t *testing.T) {
	s := helpers.CreateHTTPServer(
		&helpers.DefaultServerConfig,
		&helpers.DefaultServicesConfig,
		nil, nil, nil, nil, nil,
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.GetAuthTokenHeader(s, w, r)
	})

	request, err := http.NewRequest(http.MethodGet, "an url", http.NoBody)
	assert.NoError(t, err)
	request.Header.Set(server.JWTAuthTokenHeader, "token")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	assert.Equal(t, recorder.Code, http.StatusForbidden)
	assert.Contains(t, recorder.Body.String(), "Invalid/Malformed auth token")
}

func TestGetAuthTokenXRHHeader(t *testing.T) {
	s := helpers.CreateHTTPServer(
		&helpers.DefaultServerConfigXRH,
		&helpers.DefaultServicesConfig,
		nil, nil, nil, nil, nil,
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.GetAuthTokenHeader(s, w, r)
	})

	request, err := http.NewRequest(http.MethodGet, "an url", http.NoBody)
	assert.NoError(t, err)
	request.Header.Set(server.XRHAuthTokenHeader, "token")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	// if xrh auth type is used, contents of header are checked in different function, thus only calling GetAuthTokenHeader is successful
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestGetXRHAuthTokenEmpty(t *testing.T) {
	s := helpers.CreateHTTPServer(
		&helpers.DefaultServerConfigXRH,
		&helpers.DefaultServicesConfig,
		nil, nil, nil, nil, nil,
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.GetAuthTokenHeader(s, w, r)
	})

	request, err := http.NewRequest(http.MethodGet, "an url", http.NoBody)
	assert.NoError(t, err)
	request.Header.Set(server.XRHAuthTokenHeader, "")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	assert.Equal(t, recorder.Code, http.StatusForbidden)
	assert.Contains(t, recorder.Body.String(), "Missing auth token")
}

func TestGetAuthTokenHeaderMissing(t *testing.T) {
	s := helpers.CreateHTTPServer(
		&helpers.DefaultServerConfig,
		&helpers.DefaultServicesConfig,
		nil, nil, nil, nil, nil,
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.GetAuthTokenHeader(s, w, r)
	})

	request, err := http.NewRequest(http.MethodGet, "an url", http.NoBody)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	assert.Equal(t, recorder.Code, http.StatusForbidden)
	assert.Contains(t, recorder.Body.String(), "Missing auth token")
}

func getRequest(t *testing.T, identity string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, "an url", http.NoBody)
	assert.NoError(t, err)

	if identity == "valid" {
		ctx := context.WithValue(req.Context(), types.ContextKeyUser, validIdentityXRH)
		req = req.WithContext(ctx)
	}

	if identity == "bad" {
		ctx := context.WithValue(req.Context(), types.ContextKeyUser, "not an identity")
		req = req.WithContext(ctx)
	}

	return req
}
