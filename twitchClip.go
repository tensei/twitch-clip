package twitchClip

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
func (t *Twitch) RefreshAuthToken() (*authResponse, error) {
	return t.refreshAuthToken()
}

func (t *Twitch) refreshAuthToken() (*authResponse, error) {
	u := url.URL{
		Scheme:   "https",
		Host:     "api.twitch.tv",
		Path:     "/kraken/oauth2/token",
		RawQuery: fmt.Sprintf("client_id=%s&client_secret=%s&refresh_token=%s&grant_type=refresh_token", t.ClientID, t.ClientSecret, t.RefreshToken),
	}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating POST request: %v", err)
	}

	response, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed POST request: %v", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading body: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(string(body) + response.Status)
	}

	auth := authResponse{}
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
func (t *Twitch) CreateClip(broadcasterid string) (string, error) {
	return t.createClip(broadcasterid)
}

func (t *Twitch) createClip(broadcasterid string) (string, error) {
	if t.AccessToken == "" && t.Auth == nil {
		return "", errors.New("Authenticate first!")
	}
	u := url.URL{
		Scheme:   "https",
		Host:     "api.twitch.tv",
		Path:     "/helix/clips",
		RawQuery: fmt.Sprintf("broadcaster_id=%s", broadcasterid),
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed creating POST request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed doing POST request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading body: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return "", errors.New(string(body))
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
func (t *Twitch) GetClip(clipid string) (*Clip, error) {
	return t.getClip(clipid)
}

func (t *Twitch) getClip(clipid string) (*Clip, error) {
	if t.ClientID == "" {
		return nil, errors.New("Client_ID missing!")
	}
	u := url.URL{
		Scheme:   "https",
		Host:     "api.twitch.tv",
		Path:     "/helix/clips",
		RawQuery: fmt.Sprintf("id=%s", clipid),
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating POST request: %v", err)
	}
	req.Header.Set("Client-ID", t.ClientID)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed doing POST request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	var clip Clip
	err = json.Unmarshal(body, &clip)
	if err != nil {
		return &clip, fmt.Errorf("failed unmarshaling body: %s with error: %v", body, err)
	}

	return &clip, nil
}
