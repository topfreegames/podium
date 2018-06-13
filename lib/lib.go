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

type requestError struct {
	statusCode int
	body       string
}

func newRequestError(statusCode int, body string) *requestError {
	return &requestError{
		statusCode: statusCode,
		body:       body,
	}
}

func (r *requestError) Error() string {
	return fmt.Sprintf("Request error. Status code: %d. Body: %s", r.statusCode, r.body)
}

// Podium is a struct that represents a podium API application
type Podium struct {
	httpClient *http.Client
	Config     *viper.Viper
	URL        string
	User       string
	Pass       string
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
	Members  []*Member
	Member   *Member
	NotFound []string
}

//Score will represent a member Score in a Leaderboard
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
func NewPodium(config *viper.Viper) PodiumInterface {
	p := &Podium{
		httpClient: getHTTPClient(),
		Config:     config,
		URL:        config.GetString("podium.url"),
		User:       config.GetString("podium.user"),
		Pass:       config.GetString("podium.pass"),
	}
	return p
}

func (p *Podium) sendTo(method, url string, payload map[string]interface{}) ([]byte, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var req *http.Request

	if payload != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(payloadJSON))
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return nil, err
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(p.User, p.Pass)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, respErr := ioutil.ReadAll(resp.Body)
	if respErr != nil {
		return nil, respErr
	}

	if resp.StatusCode > 399 {
		return nil, newRequestError(resp.StatusCode, string(body))
	}

	return body, nil
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

func (p *Podium) buildUpdateScoreURL(leaderboard string, memberID string) string {
	var pathname = fmt.Sprintf("/l/%s/members/%s/score", leaderboard, memberID)
	return p.buildURL(pathname)
}

func (p *Podium) buildIncrementScoreURL(leaderboard string, memberID string) string {
	return p.buildUpdateScoreURL(leaderboard, memberID)
}

func (p *Podium) buildUpdateScoresURL(memberID string) string {
	var pathname = fmt.Sprintf("/m/%s/scores", memberID)
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

func (p *Podium) buildGetMemberURL(leaderboard string, memberID string) string {
	pathname := fmt.Sprintf("/l/%s/members/%s", leaderboard, memberID)
	return p.buildURL(pathname)
}

func (p *Podium) buildGetMembersURL(leaderboard string, memberIDs []string) string {
	memberIDsCsv := strings.Join(memberIDs, ",")
	pathname := fmt.Sprintf("/l/%s/members?ids=%s", leaderboard, memberIDsCsv)
	return p.buildURL(pathname)
}

func (p *Podium) buildHealthcheckURL() string {
	var pathname = "/healthcheck"
	return p.buildURL(pathname)
}

// GetTop returns the top members for this leaderboard. Page is 1-index
func (p *Podium) GetTop(leaderboard string, page int, pageSize int) (*MemberList, error) {
	route := p.buildGetTopURL(leaderboard, page, pageSize)
	body, err := p.sendTo("GET", route, nil)
	if err != nil {
		return nil, err
	}

	var members MemberList
	err = json.Unmarshal(body, &members)

	return &members, err
}

// GetTopPercent returns the top x% of members in a leaderboard
func (p *Podium) GetTopPercent(leaderboard string, percentage int) (*MemberList, error) {
	route := p.buildGetTopPercentURL(leaderboard, percentage)
	body, err := p.sendTo("GET", route, nil)

	if err != nil {
		return nil, err
	}

	var members MemberList
	err = json.Unmarshal(body, &members)

	return &members, err
}

// UpdateScore updates the score of a particular member in a leaderboard
func (p *Podium) UpdateScore(leaderboard string, memberID string, score int) (*MemberList, error) {
	route := p.buildUpdateScoreURL(leaderboard, memberID)
	payload := map[string]interface{}{
		"score": score,
	}
	body, err := p.sendTo("PUT", route, payload)

	if err != nil {
		return nil, err
	}

	var member MemberList
	err = json.Unmarshal(body, &member)

	return &member, err
}

// IncrementScore increments the score of a particular member in a leaderboard
func (p *Podium) IncrementScore(leaderboard string, memberID string, increment int) (*MemberList, error) {
	route := p.buildIncrementScoreURL(leaderboard, memberID)
	payload := map[string]interface{}{
		"increment": increment,
	}
	body, err := p.sendTo("PATCH", route, payload)

	if err != nil {
		return nil, err
	}

	var member MemberList
	err = json.Unmarshal(body, &member)

	return &member, err
}

// UpdateScores updates the score of a member in more than one leaderboard
func (p *Podium) UpdateScores(leaderboards []string, memberID string, score int) (*ScoreList, error) {
	route := p.buildUpdateScoresURL(memberID)
	payload := map[string]interface{}{
		"score":        score,
		"leaderboards": leaderboards,
	}
	body, err := p.sendTo("PUT", route, payload)

	if err != nil {
		return nil, err
	}

	var scores ScoreList
	err = json.Unmarshal(body, &scores)

	return &scores, err
}

// RemoveMemberFromLeaderboard removes a member from a leaderboard
func (p *Podium) RemoveMemberFromLeaderboard(leaderboard string, member string) (*Response, error) {
	route := p.buildRemoveMemberFromLeaderboardURL(leaderboard, member)
	body, err := p.sendTo("DELETE", route, nil)

	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)

	return &response, err
}

// GetMember shows score and rank of a particular member in a leaderboard
func (p *Podium) GetMember(leaderboard string, memberID string) (*Member, error) {
	route := p.buildGetMemberURL(leaderboard, memberID)
	body, err := p.sendTo("GET", route, nil)

	if err != nil {
		return nil, err
	}

	var member Member
	err = json.Unmarshal(body, &member)

	return &member, err
}

// GetMembers returns the members for this leaderboard. Page is 1-index
func (p *Podium) GetMembers(leaderboard string, memberIDs []string) (*MemberList, error) {
	route := p.buildGetMembersURL(leaderboard, memberIDs)
	body, err := p.sendTo("GET", route, nil)

	if err != nil {
		return nil, err
	}

	var members MemberList
	err = json.Unmarshal(body, &members)

	return &members, err
}

// Healthcheck verifies if podium is still up
func (p *Podium) Healthcheck() (string, error) {
	route := p.buildHealthcheckURL()
	body, err := p.sendTo("GET", route, nil)
	return string(body), err
}

// DeleteLeaderboard deletes the leaderboard from podium
func (p *Podium) DeleteLeaderboard(leaderboard string) (*Response, error) {
	route := p.buildDeleteLeaderboardURL(leaderboard)
	body, err := p.sendTo("DELETE", route, nil)

	if err != nil {
		return nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)

	return &response, err
}
