// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var baseURL string

func doRequest(method, url, reqBody string) (int, string, error) {
	absURL := fmt.Sprintf("%s%s", baseURL, url)

	var req *http.Request
	if reqBody != "" {
		req, _ = http.NewRequest(method, absURL, bytes.NewBuffer([]byte(reqBody)))
	} else {
		req, _ = http.NewRequest(method, absURL, nil)
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return 500, "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 500, "", err
	}
	return response.StatusCode, string(body), nil
}

// smokeCmd represents the smoke command
var smokeCmd = &cobra.Command{
	Use:   "smoke",
	Short: "performs a smoke test",
	Long: `Runs a smoke test in a given instance of podium.
A smoke test will perform all the available operations in a leaderboard and then remove it.
`,
	Run: func(cmd *cobra.Command, args []string) {
		doHealthCheck()

		leaderboardID := uuid.NewV4().String()
		fmt.Printf("Creating leaderboard %s...\n\n", leaderboardID)

		fmt.Println("Adding member scores to leaderboard...")
		for i := 0; i < 100; i++ {
			addMemberScore(leaderboardID, fmt.Sprintf("member-%d", i), 100-i)
		}
		fmt.Println("Member scores added to leaderboard successfully.\n")

		fmt.Println("Getting member details from leaderboard...")
		for i := 0; i < 100; i++ {
			getMember(leaderboardID, fmt.Sprintf("member-%d", i))
		}
		fmt.Println("Member details retrieved successfully.\n")

		fmt.Println("Getting many members from leaderboard...")
		memberIDs := []string{}
		for i := 0; i < 100; i++ {
			memberIDs = append(memberIDs, fmt.Sprintf("member-%d", i))
		}
		getMembers(leaderboardID, strings.Join(memberIDs, ","))
		fmt.Println("Members retrieved successfully.\n")

		fmt.Println("Getting members ranks from leaderboard...")
		for i := 0; i < 100; i++ {
			getRank(leaderboardID, fmt.Sprintf("member-%d", i))
		}
		fmt.Println("Members ranks retrieved successfully.\n")

		fmt.Println("Getting members around a member from leaderboard...")
		for i := 0; i < 100; i++ {
			getAround(leaderboardID, fmt.Sprintf("member-%d", i))
		}
		fmt.Println("Members around a member retrieved successfully.\n")

		fmt.Println("Getting number of members in a leaderboard...")
		getNumberOfMembers(leaderboardID)
		fmt.Println("Number of members retrieved successfully.\n")

		fmt.Println("Getting top members in a leaderboard...")
		getTopMembers(leaderboardID)
		fmt.Println("Top members retrieved successfully.\n")

		fmt.Println("Getting top 5% members in a leaderboard...")
		getTopPercentage(leaderboardID)
		fmt.Println("Top 5% retrieved successfully.\n")

		fmt.Println("Removing members from leaderboard...")
		for i := 0; i < 100; i++ {
			removeMember(leaderboardID, fmt.Sprintf("member-%d", i))
		}
		fmt.Println("Members removed successfully.\n")

		fmt.Println("Removing leaderboard...")
		removeLeaderboard(leaderboardID)
		fmt.Println("Leaderboard removed successfully.")
	},
}

func doHealthCheck() {
	fmt.Println("Starting smoke test...")
	status, body, err := doRequest("GET", "/healthcheck", "")
	if err != nil {
		log.Fatalf("Could not reach %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 || body != "WORKING" {
		log.Fatalf("Could not reach %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func addMemberScore(leaderboardID, memberID string, score int) {
	url := fmt.Sprintf("/l/%s/members/%s/score", leaderboardID, memberID)
	status, body, err := doRequest(
		"PUT",
		url,
		fmt.Sprintf("{\"score\":%d}", score),
	)
	if err != nil {
		log.Fatalf("Could not set member score at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not set member score at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func getMember(leaderboardID, memberID string) {
	url := fmt.Sprintf("/l/%s/members/%s", leaderboardID, memberID)
	status, body, err := doRequest(
		"GET",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not get member at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not get member at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func getMembers(leaderboardID, memberIDs string) {
	url := fmt.Sprintf("/l/%s/members?ids=%s", leaderboardID, memberIDs)
	status, body, err := doRequest(
		"GET",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not get members at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not get members at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}
func getRank(leaderboardID, memberID string) {
	url := fmt.Sprintf("/l/%s/members/%s/rank", leaderboardID, memberID)
	status, body, err := doRequest(
		"GET",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not get member rank at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not get member rank at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func getAround(leaderboardID, memberID string) {
	url := fmt.Sprintf("/l/%s/members/%s/around", leaderboardID, memberID)
	status, body, err := doRequest(
		"GET",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not get member around at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not get member around at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func getNumberOfMembers(leaderboardID string) {
	url := fmt.Sprintf("/l/%s/members-count", leaderboardID)
	status, body, err := doRequest(
		"GET",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not get members count at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not get members count at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func getTopMembers(leaderboardID string) {
	url := fmt.Sprintf("/l/%s/top/1?pageSize=20", leaderboardID)

	status, body, err := doRequest(
		"GET",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not get top members at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not get top members at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func getTopPercentage(leaderboardID string) {
	url := fmt.Sprintf("/l/%s/top-percent/5", leaderboardID)

	status, body, err := doRequest(
		"GET",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not get top 5% at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not get top 5% at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func removeMember(leaderboardID, memberID string) {
	url := fmt.Sprintf("/l/%s/members/%s", leaderboardID, memberID)
	status, body, err := doRequest(
		"DELETE",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not remove member at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not remove member at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func removeLeaderboard(leaderboardID string) {
	url := fmt.Sprintf("/l/%s", leaderboardID)
	status, body, err := doRequest(
		"DELETE",
		url,
		"",
	)
	if err != nil {
		log.Fatalf("Could not delete leaderboard at %s. Error: %s", baseURL, err.Error())
	}
	if status != 200 {
		log.Fatalf("Could not delete leaderboard at %s (Status: %d). Error: %s", baseURL, status, body)
	}
}

func init() {
	RootCmd.AddCommand(smokeCmd)
	smokeCmd.Flags().StringVarP(&baseURL, "base-url", "b", "http://localhost:8888", "Base URL for podium.")
}
