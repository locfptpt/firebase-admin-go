// Copyright 2019 Google Inc. All Rights Reserved.
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

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"google.golang.org/api/iterator"
)

func TestAuthForTenantEmptyTenantID(t *testing.T) {
	s := echoServer(testGetUserResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("")
	if client != nil || err == nil {
		t.Errorf("AuthForTenant() = (%v, %v); want = (nil, error)", client, err)
	}
}

func TestTenantID(t *testing.T) {
	s := echoServer(testGetUserResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	const want = "tenantID"
	tenantID := client.TenantID()
	if tenantID != want {
		t.Errorf("TenantID() = %q; want = %q", tenantID, want)
	}

	if client.userManagementClient.tenantID != want {
		t.Errorf("userManagementClient.tenantID = %q; want = %q", client.userManagementClient.tenantID, want)
	}

	if client.providerConfigClient.tenantID != want {
		t.Errorf("providerConfigClient.tenantID = %q; want = %q", client.providerConfigClient.tenantID, want)
	}
}

func TestTenantGetUser(t *testing.T) {
	s := echoServer(testGetUserResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	user, err := client.GetUser(context.Background(), "ignored_id")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(user, testUser) {
		t.Errorf("GetUser() = %#v; want = %#v", user, testUser)
	}

	want := `{"localId":["ignored_id"]}`
	got := string(s.Rbody)
	if got != want {
		t.Errorf("GetUser() Req = %v; want = %v", got, want)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:lookup"
	if s.Req[0].RequestURI != wantPath {
		t.Errorf("GetUser() URL = %q; want = %q", s.Req[0].RequestURI, wantPath)
	}
}

func TestTenantGetUserByEmail(t *testing.T) {
	s := echoServer(testGetUserResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	user, err := client.GetUserByEmail(context.Background(), "test@email.com")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(user, testUser) {
		t.Errorf("GetUserByEmail() = %#v; want = %#v", user, testUser)
	}

	want := `{"email":["test@email.com"]}`
	got := string(s.Rbody)
	if got != want {
		t.Errorf("GetUserByEmail() Req = %v; want = %v", got, want)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:lookup"
	if s.Req[0].RequestURI != wantPath {
		t.Errorf("GetUserByEmail() URL = %q; want = %q", s.Req[0].RequestURI, wantPath)
	}
}

func TestTenantGetUserByPhoneNumber(t *testing.T) {
	s := echoServer(testGetUserResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	user, err := client.GetUserByPhoneNumber(context.Background(), "+1234567890")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(user, testUser) {
		t.Errorf("GetUserByPhoneNumber() = %#v; want = %#v", user, testUser)
	}

	want := `{"phoneNumber":["+1234567890"]}`
	got := string(s.Rbody)
	if got != want {
		t.Errorf("GetUserByPhoneNumber() Req = %v; want = %v", got, want)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:lookup"
	if s.Req[0].RequestURI != wantPath {
		t.Errorf("GetUserByPhoneNumber() URL = %q; want = %q", s.Req[0].RequestURI, wantPath)
	}
}

func TestTenantListUsers(t *testing.T) {
	testListUsersResponse, err := ioutil.ReadFile("../testdata/list_users.json")
	if err != nil {
		t.Fatal(err)
	}
	s := echoServer(testListUsersResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	want := []*ExportedUserRecord{
		{UserRecord: testUser, PasswordHash: "passwordhash1", PasswordSalt: "salt1"},
		{UserRecord: testUser, PasswordHash: "passwordhash2", PasswordSalt: "salt2"},
		{UserRecord: testUser, PasswordHash: "passwordhash3", PasswordSalt: "salt3"},
	}

	testIterator := func(iter *UserIterator, token string, req string) {
		count := 0
		for i := 0; i < len(want); i++ {
			user, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(user.UserRecord, want[i].UserRecord) {
				t.Errorf("Users(%q) = %#v; want = %#v", token, user, want[i])
			}
			if user.PasswordHash != want[i].PasswordHash {
				t.Errorf("Users(%q) PasswordHash = %q; want = %q", token, user.PasswordHash, want[i].PasswordHash)
			}
			if user.PasswordSalt != want[i].PasswordSalt {
				t.Errorf("Users(%q) PasswordSalt = %q; want = %q", token, user.PasswordSalt, want[i].PasswordSalt)
			}
			count++
		}
		if count != len(want) {
			t.Errorf("Users(%q) = %d; want = %d", token, count, len(want))
		}
		if _, err := iter.Next(); err != iterator.Done {
			t.Errorf("Users(%q) = %v, want = %v", token, err, iterator.Done)
		}

		hr := s.Req[len(s.Req)-1]
		// Check the query string of the last HTTP request made.
		gotReq := hr.URL.Query().Encode()
		if gotReq != req {
			t.Errorf("Users(%q) = %q, want = %v", token, gotReq, req)
		}

		wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:batchGet"
		if hr.URL.Path != wantPath {
			t.Errorf("Users(%q) URL = %q; want = %q", token, hr.URL.Path, wantPath)
		}
	}

	testIterator(
		client.Users(context.Background(), ""),
		"",
		"maxResults=1000")
	testIterator(
		client.Users(context.Background(), "pageToken"),
		"pageToken",
		"maxResults=1000&nextPageToken=pageToken")
}

func TestTenantCreateUser(t *testing.T) {
	resp := `{
		"kind": "identitytoolkit#SignupNewUserResponse",
		"localId": "expectedUserID"
	}`
	s := echoServer([]byte(resp), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts"
	for _, tc := range createUserCases {
		uid, err := client.createUser(context.Background(), tc.params)
		if uid != "expectedUserID" || err != nil {
			t.Errorf("createUser(%#v) = (%q, %v); want = (%q, nil)", tc.params, uid, err, "expectedUserID")
		}

		want, err := json.Marshal(tc.req)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(s.Rbody, want) {
			t.Errorf("createUser(%#v) request = %v; want = %v", tc.params, string(s.Rbody), string(want))
		}

		if s.Req[0].RequestURI != wantPath {
			t.Errorf("createUser(%#v) URL = %q; want = %q", tc.params, s.Req[0].RequestURI, wantPath)
		}
	}
}

func TestTenantUpdateUser(t *testing.T) {
	resp := `{
		"kind": "identitytoolkit#SetAccountInfoResponse",
		"localId": "expectedUserID"
	}`
	s := echoServer([]byte(resp), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:update"
	for _, tc := range updateUserCases {
		err := client.updateUser(context.Background(), "uid", tc.params)
		if err != nil {
			t.Errorf("updateUser(%v) = %v; want = nil", tc.params, err)
		}

		tc.req["localId"] = "uid"
		want, err := json.Marshal(tc.req)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(s.Rbody, want) {
			t.Errorf("updateUser() request = %v; want = %v", string(s.Rbody), string(want))
		}

		if s.Req[0].RequestURI != wantPath {
			t.Errorf("updateUser(%#v) URL = %q; want = %q", tc.params, s.Req[0].RequestURI, wantPath)
		}
	}
}

func TestTenantRevokeRefreshTokens(t *testing.T) {
	resp := `{
		"kind": "identitytoolkit#SetAccountInfoResponse",
		"localId": "expectedUserID"
	}`
	s := echoServer([]byte(resp), t)
	defer s.Close()

	tc, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	before := time.Now().Unix()
	if err := tc.RevokeRefreshTokens(context.Background(), "some_uid"); err != nil {
		t.Error(err)
	}

	after := time.Now().Unix()
	var req struct {
		ValidSince string `json:"validSince"`
	}
	if err := json.Unmarshal(s.Rbody, &req); err != nil {
		t.Fatal(err)
	}

	validSince, err := strconv.ParseInt(req.ValidSince, 10, 64)
	if err != nil {
		t.Fatal(err)
	}

	if validSince > after || validSince < before {
		t.Errorf("validSince = %d, expecting time between %d and %d", validSince, before, after)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:update"
	if s.Req[0].RequestURI != wantPath {
		t.Errorf("RevokeRefreshTokens() URL = %q; want = %q", s.Req[0].RequestURI, wantPath)
	}
}

func TestTenantSetCustomUserClaims(t *testing.T) {
	resp := `{
		"kind": "identitytoolkit#SetAccountInfoResponse",
		"localId": "uid"
	}`
	s := echoServer([]byte(resp), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:update"
	for _, tc := range setCustomUserClaimsCases {
		err := client.SetCustomUserClaims(context.Background(), "uid", tc)
		if err != nil {
			t.Errorf("SetCustomUserClaims(%v) = %v; want nil", tc, err)
		}

		input := tc
		if input == nil {
			input = map[string]interface{}{}
		}
		b, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}

		m := map[string]interface{}{
			"localId":          "uid",
			"customAttributes": string(b),
		}
		want, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(s.Rbody, want) {
			t.Errorf("SetCustomUserClaims() = %v; want = %v", string(s.Rbody), string(want))
		}

		hr := s.Req[len(s.Req)-1]
		if hr.RequestURI != wantPath {
			t.Errorf("RevokeRefreshTokens() URL = %q; want = %q", hr.RequestURI, wantPath)
		}
	}
}

func TestTenantImportUsers(t *testing.T) {
	s := echoServer([]byte("{}"), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	users := []*UserToImport{
		(&UserToImport{}).UID("user1"),
		(&UserToImport{}).UID("user2"),
	}
	result, err := client.ImportUsers(context.Background(), users)
	if err != nil {
		t.Fatal(err)
	}

	if result.SuccessCount != 2 || result.FailureCount != 0 {
		t.Errorf("ImportUsers() = %#v; want = {SuccessCount: 2, FailureCount: 0}", result)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:batchCreate"
	if s.Req[0].RequestURI != wantPath {
		t.Errorf("ImportUsers() URL = %q; want = %q", s.Req[0].RequestURI, wantPath)
	}
}

func TestTenantImportUsersWithHash(t *testing.T) {
	s := echoServer([]byte("{}"), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	users := []*UserToImport{
		(&UserToImport{}).UID("user1").PasswordHash([]byte("password")),
		(&UserToImport{}).UID("user2"),
	}
	result, err := client.ImportUsers(context.Background(), users, WithHash(mockHash{
		key:        "key",
		saltSep:    ",",
		rounds:     8,
		memoryCost: 14,
	}))
	if err != nil {
		t.Fatal(err)
	}

	if result.SuccessCount != 2 || result.FailureCount != 0 {
		t.Errorf("ImportUsers() = %#v; want = {SuccessCount: 2, FailureCount: 0}", result)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(s.Rbody, &got); err != nil {
		t.Fatal(err)
	}

	want := map[string]interface{}{
		"hashAlgorithm": "MOCKHASH",
		"signerKey":     "key",
		"saltSeparator": ",",
		"rounds":        float64(8),
		"memoryCost":    float64(14),
	}
	for k, v := range want {
		gv, ok := got[k]
		if !ok || gv != v {
			t.Errorf("ImportUsers() request(%q) = %v; want = %v", k, gv, v)
		}
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:batchCreate"
	if s.Req[0].RequestURI != wantPath {
		t.Errorf("ImportUsers() URL = %q; want = %q", s.Req[0].RequestURI, wantPath)
	}
}

func TestTenantDeleteUser(t *testing.T) {
	resp := `{
		"kind": "identitytoolkit#SignupNewUserResponse",
		"email": "",
		"localId": "expectedUserID"
	}`
	s := echoServer([]byte(resp), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	if err := client.DeleteUser(context.Background(), "uid"); err != nil {
		t.Errorf("DeleteUser() = %v; want = nil", err)
	}

	wantPath := "/projects/mock-project-id/tenants/tenantID/accounts:delete"
	if s.Req[0].RequestURI != wantPath {
		t.Errorf("DeleteUser() URL = %q; want = %q", s.Req[0].RequestURI, wantPath)
	}
}

func TestTenantSessionCookie(t *testing.T) {
	resp := `{
		"sessionCookie": "expectedCookie"
	}`
	s := echoServer([]byte(resp), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	for _, tc := range createSessionCookieCases {
		cookie, err := client.SessionCookie(context.Background(), "idToken", tc.expiresIn)
		if cookie != "expectedCookie" || err != nil {
			t.Errorf("SessionCookie() = (%q, %v); want = (%q, nil)", cookie, err, "expectedCookie")
		}

		wantURL := "/projects/mock-project-id/tenants/tenantID:createSessionCookie"
		if s.Req[0].URL.Path != wantURL {
			t.Errorf("SesionCookie() URL = %q; want = %q", s.Req[0].URL.Path, wantURL)
		}

		var got map[string]interface{}
		if err := json.Unmarshal(s.Rbody, &got); err != nil {
			t.Fatal(err)
		}
		want := map[string]interface{}{
			"idToken":       "idToken",
			"validDuration": tc.want,
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("SessionCookie(%f) request =%#v; want = %#v", tc.want, got, want)
		}
	}
}

const wantEmailActionURL = "/projects/mock-project-id/tenants/tenantID/accounts:sendOobCode"

func TestTenantEmailVerificationLink(t *testing.T) {
	s := echoServer(testActionLinkResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	link, err := client.EmailVerificationLink(context.Background(), testEmail)
	if err != nil {
		t.Fatal(err)
	}
	if link != testActionLink {
		t.Errorf("EmailVerificationLink() = %q; want = %q", link, testActionLink)
	}

	want := map[string]interface{}{
		"requestType":   "VERIFY_EMAIL",
		"email":         testEmail,
		"returnOobLink": true,
	}
	if err := checkActionLinkRequestWithURL(want, wantEmailActionURL, s); err != nil {
		t.Fatalf("EmailVerificationLink() %v", err)
	}
}

func TestTenantPasswordResetLink(t *testing.T) {
	s := echoServer(testActionLinkResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	link, err := client.PasswordResetLink(context.Background(), testEmail)
	if err != nil {
		t.Fatal(err)
	}
	if link != testActionLink {
		t.Errorf("PasswordResetLink() = %q; want = %q", link, testActionLink)
	}

	want := map[string]interface{}{
		"requestType":   "PASSWORD_RESET",
		"email":         testEmail,
		"returnOobLink": true,
	}
	if err := checkActionLinkRequestWithURL(want, wantEmailActionURL, s); err != nil {
		t.Fatalf("PasswordResetLink() %v", err)
	}
}

func TestTenantEmailSignInLink(t *testing.T) {
	s := echoServer(testActionLinkResponse, t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	link, err := client.EmailSignInLink(context.Background(), testEmail, testActionCodeSettings)
	if err != nil {
		t.Fatal(err)
	}
	if link != testActionLink {
		t.Errorf("EmailSignInLink() = %q; want = %q", link, testActionLink)
	}

	want := map[string]interface{}{
		"requestType":   "EMAIL_SIGNIN",
		"email":         testEmail,
		"returnOobLink": true,
	}
	for k, v := range testActionCodeSettingsMap {
		want[k] = v
	}
	if err := checkActionLinkRequestWithURL(want, wantEmailActionURL, s); err != nil {
		t.Fatalf("EmailSignInLink() %v", err)
	}
}

func TestTenantOIDCProviderConfig(t *testing.T) {
	s := echoServer([]byte(oidcConfigResponse), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	oidc, err := client.OIDCProviderConfig(context.Background(), "oidc.provider")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(oidc, oidcProviderConfig) {
		t.Errorf("OIDCProviderConfig() = %#v; want = %#v", oidc, oidcProviderConfig)
	}

	req := s.Req[0]
	if req.Method != http.MethodGet {
		t.Errorf("OIDCProviderConfig() Method = %q; want = %q", req.Method, http.MethodGet)
	}

	wantURL := "/projects/mock-project-id/tenants/tenantID/oauthIdpConfigs/oidc.provider"
	if req.URL.Path != wantURL {
		t.Errorf("OIDCProviderConfig() URL = %q; want = %q", req.URL.Path, wantURL)
	}
}

func TestTenantCreateOIDCProviderConfig(t *testing.T) {
	s := echoServer([]byte(oidcConfigResponse), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	options := (&OIDCProviderConfigToCreate{}).
		ID(oidcProviderConfig.ID).
		DisplayName(oidcProviderConfig.DisplayName).
		Enabled(oidcProviderConfig.Enabled).
		ClientID(oidcProviderConfig.ClientID).
		Issuer(oidcProviderConfig.Issuer)
	oidc, err := client.CreateOIDCProviderConfig(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(oidc, oidcProviderConfig) {
		t.Errorf("CreateOIDCProviderConfig() = %#v; want = %#v", oidc, oidcProviderConfig)
	}

	wantBody := map[string]interface{}{
		"displayName": oidcProviderConfig.DisplayName,
		"enabled":     oidcProviderConfig.Enabled,
		"clientId":    oidcProviderConfig.ClientID,
		"issuer":      oidcProviderConfig.Issuer,
	}
	wantURL := "/projects/mock-project-id/tenants/tenantID/oauthIdpConfigs"
	if err := checkCreateOIDCConfigRequestWithURL(s, wantBody, wantURL); err != nil {
		t.Fatal(err)
	}
}

func TestTenantUpdateOIDCProviderConfig(t *testing.T) {
	s := echoServer([]byte(oidcConfigResponse), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	options := (&OIDCProviderConfigToUpdate{}).
		DisplayName(oidcProviderConfig.DisplayName).
		Enabled(oidcProviderConfig.Enabled).
		ClientID(oidcProviderConfig.ClientID).
		Issuer(oidcProviderConfig.Issuer)
	oidc, err := client.UpdateOIDCProviderConfig(context.Background(), "oidc.provider", options)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(oidc, oidcProviderConfig) {
		t.Errorf("UpdateOIDCProviderConfig() = %#v; want = %#v", oidc, oidcProviderConfig)
	}

	wantBody := map[string]interface{}{
		"displayName": oidcProviderConfig.DisplayName,
		"enabled":     oidcProviderConfig.Enabled,
		"clientId":    oidcProviderConfig.ClientID,
		"issuer":      oidcProviderConfig.Issuer,
	}
	wantMask := []string{
		"clientId",
		"displayName",
		"enabled",
		"issuer",
	}
	wantURL := "/projects/mock-project-id/tenants/tenantID/oauthIdpConfigs/oidc.provider"
	if err := checkUpdateOIDCConfigRequestWithURL(s, wantBody, wantMask, wantURL); err != nil {
		t.Fatal(err)
	}
}

func TestTenantDeleteOIDCProviderConfig(t *testing.T) {
	s := echoServer([]byte("{}"), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	if err := client.DeleteOIDCProviderConfig(context.Background(), "oidc.provider"); err != nil {
		t.Fatal(err)
	}

	req := s.Req[0]
	if req.Method != http.MethodDelete {
		t.Errorf("DeleteOIDCProviderConfig() Method = %q; want = %q", req.Method, http.MethodDelete)
	}

	wantURL := "/projects/mock-project-id/tenants/tenantID/oauthIdpConfigs/oidc.provider"
	if req.URL.Path != wantURL {
		t.Errorf("DeleteOIDCProviderConfig() URL = %q; want = %q", req.URL.Path, wantURL)
	}
}

func TestTenantOIDCProviderConfigs(t *testing.T) {
	template := `{
                "oauthIdpConfigs": [
                    %s,
                    %s,
                    %s
                ],
                "nextPageToken": ""
        }`
	response := fmt.Sprintf(template, oidcConfigResponse, oidcConfigResponse, oidcConfigResponse)
	s := echoServer([]byte(response), t)
	defer s.Close()

	want := []*OIDCProviderConfig{
		oidcProviderConfig,
		oidcProviderConfig,
		oidcProviderConfig,
	}
	wantPath := "/projects/mock-project-id/tenants/tenantID/oauthIdpConfigs"

	testIterator := func(iter *OIDCProviderConfigIterator, token string, req string) {
		count := 0
		for i := 0; i < len(want); i++ {
			config, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(config, want[i]) {
				t.Errorf("OIDCProviderConfigs(%q) = %#v; want = %#v", token, config, want[i])
			}
			count++
		}
		if count != len(want) {
			t.Errorf("OIDCProviderConfigs(%q) = %d; want = %d", token, count, len(want))
		}
		if _, err := iter.Next(); err != iterator.Done {
			t.Errorf("OIDCProviderConfigs(%q) = %v; want = %v", token, err, iterator.Done)
		}

		url := s.Req[len(s.Req)-1].URL
		if url.Path != wantPath {
			t.Errorf("OIDCProviderConfigs(%q) = %q; want = %q", token, url.Path, wantPath)
		}

		// Check the query string of the last HTTP request made.
		gotReq := url.Query().Encode()
		if gotReq != req {
			t.Errorf("OIDCProviderConfigs(%q) = %q; want = %v", token, gotReq, req)
		}
	}

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	testIterator(
		client.OIDCProviderConfigs(context.Background(), ""),
		"",
		"pageSize=100")
	testIterator(
		client.OIDCProviderConfigs(context.Background(), "pageToken"),
		"pageToken",
		"pageSize=100&pageToken=pageToken")
}

func TestTenantSAMLProviderConfig(t *testing.T) {
	s := echoServer([]byte(samlConfigResponse), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	saml, err := client.SAMLProviderConfig(context.Background(), "saml.provider")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(saml, samlProviderConfig) {
		t.Errorf("SAMLProviderConfig() = %#v; want = %#v", saml, samlProviderConfig)
	}

	req := s.Req[0]
	if req.Method != http.MethodGet {
		t.Errorf("SAMLProviderConfig() Method = %q; want = %q", req.Method, http.MethodGet)
	}

	wantURL := "/projects/mock-project-id/tenants/tenantID/inboundSamlConfigs/saml.provider"
	if req.URL.Path != wantURL {
		t.Errorf("SAMLProviderConfig() URL = %q; want = %q", req.URL.Path, wantURL)
	}
}

func TestTenantCreateSAMLProviderConfig(t *testing.T) {
	s := echoServer([]byte(samlConfigResponse), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	options := (&SAMLProviderConfigToCreate{}).
		ID(samlProviderConfig.ID).
		DisplayName(samlProviderConfig.DisplayName).
		Enabled(samlProviderConfig.Enabled).
		IDPEntityID(samlProviderConfig.IDPEntityID).
		SSOURL(samlProviderConfig.SSOURL).
		RequestSigningEnabled(samlProviderConfig.RequestSigningEnabled).
		X509Certificates(samlProviderConfig.X509Certificates).
		RPEntityID(samlProviderConfig.RPEntityID).
		CallbackURL(samlProviderConfig.CallbackURL)
	saml, err := client.CreateSAMLProviderConfig(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(saml, samlProviderConfig) {
		t.Errorf("CreateSAMLProviderConfig() = %#v; want = %#v", saml, samlProviderConfig)
	}

	wantBody := map[string]interface{}{
		"displayName": samlProviderConfig.DisplayName,
		"enabled":     samlProviderConfig.Enabled,
		"idpConfig": map[string]interface{}{
			"idpEntityId":     samlProviderConfig.IDPEntityID,
			"ssoUrl":          samlProviderConfig.SSOURL,
			"signRequest":     samlProviderConfig.RequestSigningEnabled,
			"idpCertificates": idpCertsMap,
		},
		"spConfig": map[string]interface{}{
			"spEntityId":  samlProviderConfig.RPEntityID,
			"callbackUri": samlProviderConfig.CallbackURL,
		},
	}
	wantURL := "/projects/mock-project-id/tenants/tenantID/inboundSamlConfigs"
	if err := checkCreateSAMLConfigRequestWithURL(s, wantBody, wantURL); err != nil {
		t.Fatal(err)
	}
}

func TestTenantUpdateSAMLProviderConfig(t *testing.T) {
	s := echoServer([]byte(samlConfigResponse), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	options := (&SAMLProviderConfigToUpdate{}).
		DisplayName(samlProviderConfig.DisplayName).
		Enabled(samlProviderConfig.Enabled).
		IDPEntityID(samlProviderConfig.IDPEntityID).
		SSOURL(samlProviderConfig.SSOURL).
		RequestSigningEnabled(samlProviderConfig.RequestSigningEnabled).
		X509Certificates(samlProviderConfig.X509Certificates).
		RPEntityID(samlProviderConfig.RPEntityID).
		CallbackURL(samlProviderConfig.CallbackURL)
	saml, err := client.UpdateSAMLProviderConfig(context.Background(), "saml.provider", options)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(saml, samlProviderConfig) {
		t.Errorf("UpdateSAMLProviderConfig() = %#v; want = %#v", saml, samlProviderConfig)
	}

	wantBody := map[string]interface{}{
		"displayName": samlProviderConfig.DisplayName,
		"enabled":     samlProviderConfig.Enabled,
		"idpConfig": map[string]interface{}{
			"idpEntityId":     samlProviderConfig.IDPEntityID,
			"ssoUrl":          samlProviderConfig.SSOURL,
			"signRequest":     samlProviderConfig.RequestSigningEnabled,
			"idpCertificates": idpCertsMap,
		},
		"spConfig": map[string]interface{}{
			"spEntityId":  samlProviderConfig.RPEntityID,
			"callbackUri": samlProviderConfig.CallbackURL,
		},
	}
	wantMask := []string{
		"displayName",
		"enabled",
		"idpConfig.idpCertificates",
		"idpConfig.idpEntityId",
		"idpConfig.signRequest",
		"idpConfig.ssoUrl",
		"spConfig.callbackUri",
		"spConfig.spEntityId",
	}
	wantURL := "/projects/mock-project-id/tenants/tenantID/inboundSamlConfigs/saml.provider"
	if err := checkUpdateSAMLConfigRequestWithURL(s, wantBody, wantMask, wantURL); err != nil {
		t.Fatal(err)
	}
}

func TestTenantDeleteSAMLProviderConfig(t *testing.T) {
	s := echoServer([]byte("{}"), t)
	defer s.Close()

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	if err := client.DeleteSAMLProviderConfig(context.Background(), "saml.provider"); err != nil {
		t.Fatal(err)
	}

	req := s.Req[0]
	if req.Method != http.MethodDelete {
		t.Errorf("DeleteSAMLProviderConfig() Method = %q; want = %q", req.Method, http.MethodDelete)
	}

	wantURL := "/projects/mock-project-id/tenants/tenantID/inboundSamlConfigs/saml.provider"
	if req.URL.Path != wantURL {
		t.Errorf("DeleteSAMLProviderConfig() URL = %q; want = %q", req.URL.Path, wantURL)
	}
}

func TestTenantSAMLProviderConfigs(t *testing.T) {
	template := `{
                "inboundSamlConfigs": [
                    %s,
                    %s,
                    %s
                ],
                "nextPageToken": ""
        }`
	response := fmt.Sprintf(template, samlConfigResponse, samlConfigResponse, samlConfigResponse)
	s := echoServer([]byte(response), t)
	defer s.Close()

	want := []*SAMLProviderConfig{
		samlProviderConfig,
		samlProviderConfig,
		samlProviderConfig,
	}
	wantPath := "/projects/mock-project-id/tenants/tenantID/inboundSamlConfigs"

	testIterator := func(iter *SAMLProviderConfigIterator, token string, req string) {
		count := 0
		for i := 0; i < len(want); i++ {
			config, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(config, want[i]) {
				t.Errorf("SAMLProviderConfigs(%q) = %#v; want = %#v", token, config, want[i])
			}
			count++
		}
		if count != len(want) {
			t.Errorf("SAMLProviderConfigs(%q) = %d; want = %d", token, count, len(want))
		}
		if _, err := iter.Next(); err != iterator.Done {
			t.Errorf("SAMLProviderConfigs(%q) = %v; want = %v", token, err, iterator.Done)
		}

		url := s.Req[len(s.Req)-1].URL
		if url.Path != wantPath {
			t.Errorf("SAMLProviderConfigs(%q) = %q; want = %q", token, url.Path, wantPath)
		}

		// Check the query string of the last HTTP request made.
		gotReq := url.Query().Encode()
		if gotReq != req {
			t.Errorf("SAMLProviderConfigs(%q) = %q; want = %v", token, gotReq, req)
		}
	}

	client, err := s.Client.TenantManager.AuthForTenant("tenantID")
	if err != nil {
		t.Fatalf("AuthForTenant() = %v", err)
	}

	testIterator(
		client.SAMLProviderConfigs(context.Background(), ""),
		"",
		"pageSize=100")
	testIterator(
		client.SAMLProviderConfigs(context.Background(), "pageToken"),
		"pageToken",
		"pageSize=100&pageToken=pageToken")
}
