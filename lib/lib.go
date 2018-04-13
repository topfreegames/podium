package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

// Podium is a struct that represents a podium API application
type Podium struct {
	httpClient        *http.Client
	Config            *viper.Viper
	URL               string
	User              string
	Pass              string
	baseLeaderboard   string
	localeLeaderboard string
}

var (
	client *http.Client
	once   sync.Once
)

// Member maps an member identified by their publicID to their score and rank
type Member struct {
	PublicID     string
	Score        int
	Rank         int
	previousRank int
}

//MemberList is a list of member
type MemberList struct {
	Members []*Member
	Member  *Member
}

//Score will represent a Player Score in a Leaderboard
type Score struct {
	LeaderboardID string
	PublicID      string
	Score         int
	Rank          int
}

//ScoreList is a list of Scores
type ScoreList struct {
	Scores []*Score
}

//Response will determine if a request has been succeeded
type Response struct {
	Success bool
	Reason  string
}

func getHTTPClient() *http.Client {
	once.Do(func() {
		client = &http.Client{}
	})
	return client
}

// NewPodium returns a new podium API application
func NewPodium(config *viper.Viper) *Podium {
	p := &Podium{
		httpClient:        getHTTPClient(),
		Config:            config,
		URL:               config.GetString("podium.url"),
		User:              config.GetString("podium.user"),
		Pass:              config.GetString("podium.pass"),
		baseLeaderboard:   config.GetString("leaderboards.globalLeaderboard"),
		localeLeaderboard: config.GetString("leaderboards.localeLeaderboard"),
	}
	return p
}

func (p *Podium) sendTo(method, url string, payload map[string]interface{}) (int, []byte, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return -1, nil, err
	}

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(payloadJSON))
		if err != nil {
			return -1, nil, err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return -1, nil, err
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(p.User, p.Pass)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer resp.Body.Close()

	body, respErr := ioutil.ReadAll(resp.Body)

	if respErr != nil {
		return -1, nil, respErr
	}

	return resp.StatusCode, body, nil
}

func (p *Podium) buildURL(pathname string) string {
	return fmt.Sprintf("%s%s", p.URL, pathname)
}

func (p *Podium) buildDeleteLeaderboardURL(leaderboard string) string {
	var pathname = fmt.Sprintf("/l/%s", leaderboard)
	return p.buildURL(pathname)
}

func (p *Podium) buildGetTopPercentURL(leaderboard string, percentage int) string {
	var pathname = fmt.Sprintf("/l/%s/top-percent/%d", leaderboard, percentage)
	return p.buildURL(pathname)
}

func (p *Podium) buildUpdateScoreURL(leaderboard string, playerID string) string {
	var pathname = fmt.Sprintf("/l/%s/members/%s/score", leaderboard, playerID)
	return p.buildURL(pathname)
}

func (p *Podium) buildIncrementScoreURL(leaderboard string, playerID string) string {
	return p.buildUpdateScoreURL(leaderboard, playerID)
}

func (p *Podium) buildUpdateScoresURL(playerID string) string {
	var pathname = fmt.Sprintf("/m/%s/scores", playerID)
	return p.buildURL(pathname)
}

func (p *Podium) buildRemoveMemberFromLeaderboardURL(leaderboard string, member string) string {
	var pathname = fmt.Sprintf("/l/%s/members/%s", leaderboard, member)
	return p.buildURL(pathname)
}

// page is 1-based
func (p *Podium) buildGetTopURL(leaderboard string, page int, pageSize int) string {
	var pathname = fmt.Sprintf("/l/%s/top/%d?pageSize=%d", leaderboard, page, pageSize)
	return p.buildURL(pathname)
}

func (p *Podium) buildGetPlayerURL(leaderboard string, playerID string) string {
	var pathname = fmt.Sprintf("/l/%s/members/%s", leaderboard, playerID)
	return p.buildURL(pathname)
}

func (p *Podium) buildHealthcheckURL() string {
	var pathname = "/healthcheck"
	return p.buildURL(pathname)
}

// GetBaseLeaderboards shows the global leaderboard
func (p *Podium) GetBaseLeaderboards() string {
	return p.baseLeaderboard
}

// GetLocalizedLeaderboard receives a locale and returns its leaderboard
func (p *Podium) GetLocalizedLeaderboard(locale string) string {
	localeLeaderboard := p.localeLeaderboard
	result := strings.Replace(localeLeaderboard, "%{locale}", locale, -1)
	return result
}

// GetTop returns the top players for this leaderboard. Page is 1-index
func (p *Podium) GetTop(leaderboard string, page int, pageSize int) (int, *MemberList, error) {
	route := p.buildGetTopURL(leaderboard, page, pageSize)
	status, body, err := p.sendTo("GET", route, nil)

	if err != nil {
		return -1, nil, err
	}

	var members MemberList
	err = json.Unmarshal(body, &members)

	return status, &members, err
}

// GetTopPercent returns the top x% of players in a leaderboard
func (p *Podium) GetTopPercent(leaderboard string, percentage int) (int, *MemberList, error) {
	route := p.buildGetTopPercentURL(leaderboard, percentage)
	status, body, err := p.sendTo("GET", route, nil)

	if err != nil {
		return -1, nil, err
	}

	var members MemberList
	err = json.Unmarshal(body, &members)

	return status, &members, err
}

// UpdateScore updates the score of a particular player in a leaderboard
func (p *Podium) UpdateScore(leaderboard string, playerID string, score int) (int, *MemberList, error) {
	route := p.buildUpdateScoreURL(leaderboard, playerID)
	payload := map[string]interface{}{
		"score": score,
	}
	status, body, err := p.sendTo("PUT", route, payload)

	if err != nil {
		return -1, nil, err
	}

	var member MemberList
	err = json.Unmarshal(body, &member)

	return status, &member, err
}

// IncrementScore increments the score of a particular player in a leaderboard
func (p *Podium) IncrementScore(leaderboard string, playerID string, increment int) (int, *MemberList, error) {
	route := p.buildIncrementScoreURL(leaderboard, playerID)
	payload := map[string]interface{}{
		"increment": increment,
	}
	status, body, err := p.sendTo("PATCH", route, payload)

	if err != nil {
		return -1, nil, err
	}

	var member MemberList
	err = json.Unmarshal(body, &member)

	return status, &member, err
}

// UpdateScores updates the score of a player in more than one leaderboard
func (p *Podium) UpdateScores(leaderboards []string, playerID string, score int) (int, *ScoreList, error) {
	route := p.buildUpdateScoresURL(playerID)
	payload := map[string]interface{}{
		"score":        score,
		"leaderboards": leaderboards,
	}
	status, body, err := p.sendTo("PUT", route, payload)

	if err != nil {
		return -1, nil, err
	}

	var scores ScoreList
	err = json.Unmarshal(body, &scores)

	return status, &scores, err
}

// RemoveMemberFromLeaderboard removes a player from a leaderboard
func (p *Podium) RemoveMemberFromLeaderboard(leaderboard string, member string) (int, *Response, error) {
	route := p.buildRemoveMemberFromLeaderboardURL(leaderboard, member)
	status, body, err := p.sendTo("DELETE", route, nil)

	if err != nil {
		return -1, nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)

	return status, &response, err
}

// GetPlayer shows score and rank of a particular player in a leaderboard
func (p *Podium) GetPlayer(leaderboard string, playerID string) (int, *Member, error) {
	route := p.buildGetPlayerURL(leaderboard, playerID)
	status, body, err := p.sendTo("GET", route, nil)

	if err != nil {
		return -1, nil, err
	}

	var member Member
	err = json.Unmarshal(body, &member)

	return status, &member, err
}

// Healthcheck verifies if podium is still up
func (p *Podium) Healthcheck() (int, string, error) {
	route := p.buildHealthcheckURL()
	status, body, err := p.sendTo("GET", route, nil)
	return status, string(body), err
}

// DeleteLeaderboard deletes the leaderboard from podium
func (p *Podium) DeleteLeaderboard(leaderboard string) (int, *Response, error) {
	route := p.buildDeleteLeaderboardURL(leaderboard)
	status, body, err := p.sendTo("DELETE", route, nil)

	if err != nil {
		return -1, nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)

	return status, &response, err
}
