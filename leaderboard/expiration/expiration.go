// podium

// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package expiration

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var unixRE = regexp.MustCompile("from([0-9]{10})to([0-9]{10})$")                                            // unix timestamp
var timestampRE = regexp.MustCompile("from([0-9]{4}[0|1][0-9][0-3][0-9])to([0-9]{4}[0|1][0-9][0-3][0-9])$") //YYYYMMDD
var yearlyRE = regexp.MustCompile("year([0-9]{4})$")                                                        // yearly
var quarterRE = regexp.MustCompile("year([0-9]{4})(week|quarter|month)([0-9]+)$")                           //week, quarter, mo

func checkExpireAtErrors(
	leaderboardPublicID string, startTimestamp, endTimestamp int64,
) (int64, error) {
	now := time.Now().UTC().Unix()
	durationInSeconds := endTimestamp - startTimestamp
	if durationInSeconds <= 0 {
		return -1, &InvalidDurationError{leaderboardPublicID, durationInSeconds}
	}
	expireAt := endTimestamp + durationInSeconds
	if expireAt <= now {
		return -1, &LeaderboardExpiredError{leaderboardPublicID}
	}
	return expireAt, nil
}

// WeeklyExpiration calculates the expireAt date for leaderboards with weekly format
func WeeklyExpiration(year, week int64) time.Time {
	dummyDate, _ := time.Parse("2006", strconv.Itoa(int(year)))
	dummyDateYear, dummyDateWeek := dummyDate.ISOWeek()
	startTime := dummyDate.AddDate(int(year)-dummyDateYear, 0, (int(week)-dummyDateWeek)*7)
	return startTime.AddDate(0, 0, 14)
}

// MonthlyExpiration calculates the expireAt date for leaderboards with quarterly format
func MonthlyExpiration(startTime time.Time) time.Time {
	return startTime.AddDate(0, 2, 0)
}

// QuarterlyExpiration calculates the expireAt date for leaderboards with quarterly format
func QuarterlyExpiration(year, quarter int64) time.Time {
	dummyDate, _ := time.Parse("2006", strconv.Itoa(int(year)))
	startTime := dummyDate.AddDate(0, (int(quarter)-1)*3, 0)
	return startTime.AddDate(0, 6, 0)
}

// GetExpireAt returns a timestamp when the key should expire or -1 if the key doesn't match any valid auto expire regexes
func GetExpireAt(leaderboardPublicID string) (int64, error) {
	substrings := unixRE.FindStringSubmatch(leaderboardPublicID)
	if len(substrings) == 3 {
		startTimestamp, _ := strconv.ParseInt(substrings[1], 10, 32)
		endTimestamp, _ := strconv.ParseInt(substrings[2], 10, 32)
		return checkExpireAtErrors(leaderboardPublicID, startTimestamp, endTimestamp)
	}

	substrings = timestampRE.FindStringSubmatch(leaderboardPublicID)
	if len(substrings) == 3 {
		startTime, err := time.Parse("20060102", substrings[1])
		if err != nil {
			return -1, err
		}
		endTime, err := time.Parse("20060102", substrings[2])
		if err != nil {
			return -1, err
		}
		return checkExpireAtErrors(leaderboardPublicID, startTime.Unix(), endTime.Unix())
	}

	now := time.Now().UTC().Unix()

	substrings = yearlyRE.FindStringSubmatch(leaderboardPublicID)
	if len(substrings) == 2 {
		startTime, _ := time.Parse("2006", substrings[1])
		endTime := startTime.AddDate(2, 0, 0)
		if endTime.Unix() <= now {
			return -1, &LeaderboardExpiredError{leaderboardPublicID}
		}
		return endTime.Unix(), nil
	}

	substrings = quarterRE.FindStringSubmatch(leaderboardPublicID)
	if len(substrings) == 4 {
		var startTime time.Time
		var endTime time.Time
		var err error

		if substrings[2] == "month" {
			startTime, err = time.Parse("200601", strings.Join([]string{substrings[1], substrings[3]}, ""))
			if err != nil {
				return -1, err
			}
			endTime = MonthlyExpiration(startTime)
		}

		year, _ := strconv.ParseInt(substrings[1], 10, 32)
		if substrings[2] == "week" {
			week, _ := strconv.ParseInt(substrings[3], 10, 32)
			endTime = WeeklyExpiration(year, week)
		}

		if substrings[2] == "quarter" {
			quarter, _ := strconv.ParseInt(substrings[3], 10, 32)
			endTime = QuarterlyExpiration(year, quarter)
		}

		if endTime.Unix() <= now {
			return -1, &LeaderboardExpiredError{leaderboardPublicID}
		}
		return endTime.Unix(), nil
	}

	return -1, nil
}
