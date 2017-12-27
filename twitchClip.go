package twitchClip

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/net/context/ctxhttp"
)

type Twitch struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	Auth         *authResponse

	client *http.Client
}

type authResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
}

// Create new Twitch client
// Required: clientid, clientsecret
// Optional: accesstoken
func NewClient(clientid, clientsecret, accesstoken, refreshtoken string) (*Twitch, error) {
	if clientid == "" || clientsecret == "" {
		return nil, errors.New("Missing Client-Secret or Client-ID")
	}
	return &Twitch{
		ClientID:     clientid,
		ClientSecret: clientsecret,
		AccessToken:  accesstoken,
		RefreshToken: refreshtoken,
		client:       &http.Client{},
	}, nil
}

/*
curl -X POST https://api.twitch.tv/kraken/oauth2/token
    --data-urlencode
    ?grant_type=refresh_token
    &refresh_token=eyJfaWQmNzMtNGCJ9%6VFV5LNrZFUj8oU231/3Aj
    &client_id=fooid
    &client_secret=barbazsecret
*/

// Refresh the authtoken and refresh_token using Client-ID, Client-Secret and old Refresh_Token
// Returns new tokens on success
func (t *Twitch) RefreshAuthToken(ctx context.Context) (*authResponse, error) {
	return t.refreshAuthToken(ctx)
}

func (t *Twitch) refreshAuthToken(ctx context.Context) (*authResponse, error) {
	query := url.Values{
		"client_id":     []string{t.ClientID},
		"client_secret": []string{t.ClientSecret},
		"refresh_token": []string{t.RefreshToken},
		"grant_type":    []string{"refresh_token"},
		"scope":         []string{"clips:edit"},
	}

	req := newPostRequest("/kraken/oauth2/token", "", query)
	response, errResp := do(ctx, req)
	if errResp != nil {
		return nil, fmt.Errorf("failed POST request: %s", errResp.Message)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading body: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(string(body) + response.Status)
	}

	var auth authResponse
	err = json.Unmarshal(body, &auth)
	if err != nil {
		return nil, fmt.Errorf("failed unmarshaling body: %s with error: %v", body, err)
	}
	t.Auth = &auth
	t.AccessToken = auth.AccessToken
	t.RefreshToken = auth.RefreshToken
	return &auth, nil
}

/*
curl -H 'Authorization: Bearer cfabdegwdoklmawdzdo98xt2fo512y' \
-X POST 'https://api.twitch.tv/helix/clips?broadcaster_id=44322889'
*/

type clipCreateResponse struct {
	Data []struct {
		EditURL string `json:"edit_url"`
		ID      string `json:"id"`
	} `json:"data"`
}

// Create clip with given broadcaster id
// Returns clip id on success
func (t *Twitch) CreateClip(ctx context.Context, broadcasterid string) (string, error) {
	return t.createClip(ctx, broadcasterid)
}

func (t *Twitch) createClip(ctx context.Context, broadcasterid string) (string, error) {
	if t.AccessToken == "" && t.Auth == nil {
		return "", errors.New("Authenticate first!")
	}

	query := url.Values{
		"broadcaster_id": []string{broadcasterid},
	}

	req := newPostRequest("/helix/clips", t.AccessToken, query)
	resp, errResp := do(ctx, req)
	if errResp != nil {
		return "", fmt.Errorf("error: %s", errResp.Message)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading body: %v", err)
	}

	var cr clipCreateResponse
	err = json.Unmarshal(body, &cr)
	if err != nil {
		return "", fmt.Errorf("failed unmarshaling body: %s with error: %v", body, err)
	}
	return cr.Data[0].ID, nil
}

/*
curl -H 'Client-ID: uo6dggojyb8d6soh92zknwmi5ej1q2' \
-X GET 'https://api.twitch.tv/helix/clips?id=AwkwardHelplessSalamanderSwiftRage'
*/

type Clip struct {
	Data []struct {
		BroadcasterID string `json:"broadcaster_id"`
		CreatedAt     string `json:"created_at"`
		CreatorID     string `json:"creator_id"`
		EmbedURL      string `json:"embed_url"`
		GameID        string `json:"game_id"`
		ID            string `json:"id"`
		Language      string `json:"language"`
		ThumbnailURL  string `json:"thumbnail_url"`
		Title         string `json:"title"`
		URL           string `json:"url"`
		VideoID       string `json:"video_id"`
		ViewCount     int    `json:"view_count"`
	} `json:"data"`
}

// Get clip info with given clip id
// Returns clip info on success
func (t *Twitch) GetClip(ctx context.Context, clipid string) (*Clip, error) {
	return t.getClip(ctx, clipid)
}

func (t *Twitch) getClip(ctx context.Context, clipid string) (*Clip, error) {
	if t.ClientID == "" {
		return nil, errors.New("Client_ID missing!")
	}

	query := url.Values{
		"id": []string{clipid},
	}

	req := newGetRequest("/helix/clips", t.ClientID, query)
	resp, errResp := do(ctx, req)
	if errResp != nil {
		return nil, fmt.Errorf("failed GET request: %s", errResp.Message)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading body: %v", err)
	}

	var clip Clip
	err = json.Unmarshal(body, &clip)
	if err != nil {
		return &clip, fmt.Errorf("failed unmarshaling body: %s with error: %v", body, err)
	}

	return &clip, nil
}

func newPostRequest(path string, token string, query url.Values) *http.Request {
	u := url.URL{
		Scheme:   "https",
		Host:     "api.twitch.tv",
		Path:     path,
		RawQuery: query.Encode(),
	}
	req, _ := http.NewRequest("POST", u.String(), nil)
	if token == "" {
		return req
	}
	return withAccessToken(req, token)
}

func newGetRequest(path string, token string, query url.Values) *http.Request {
	u := url.URL{
		Scheme:   "https",
		Host:     "api.twitch.tv",
		Path:     path,
		RawQuery: query.Encode(),
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	if token == "" {
		return req
	}
	return withClientID(req, token)
}

func withAccessToken(req *http.Request, accesstoken string) *http.Request {
	req.Header.Set("Authorization", "Bearer "+accesstoken)
	return req
}

func withClientID(req *http.Request, clientid string) *http.Request {
	req.Header.Set("Client-ID", clientid)
	return req
}

/*
{
    "error": "Unauthorized",
    "message": "Token invalid or missing required scope",
    "status": 401
}
*/

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func do(ctx context.Context, req *http.Request) (*http.Response, *errorResponse) {
	var errResp errorResponse
	resp, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		errResp.Error = err.Error()
		errResp.Message = resp.Status
		errResp.Status = resp.StatusCode
		return resp, &errResp
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return resp, nil
	}

	err = json.NewDecoder(resp.Body).Decode(&errResp)
	if err != nil {
		errResp.Error = err.Error()
		errResp.Message = resp.Status
		errResp.Status = resp.StatusCode
		return resp, &errResp
	}
	return resp, &errResp
}
