// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package expiration_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/expiration"
)

var _ = Describe("Expires Helper", func() {
	Describe("No timestamps in leaderboard name", func() {
		It("should get null expiration name without expiration", func() {
			exp, err := expiration.GetExpireAt("my_leaderboard")
			Expect(err).NotTo(HaveOccurred())
			Expect(exp).To(BeEquivalentTo(-1))
		})
	})

	Describe("Yearly expiration", func() {
		It("should get expiration for year 2020", func() {
			exp, err := expiration.GetExpireAt("leaderboard_year2020")
			Expect(err).NotTo(HaveOccurred())

			startTime, err := time.Parse("2006", "2020")
			Expect(err).NotTo(HaveOccurred())

			endTime := startTime.AddDate(2, 0, 0)
			Expect(exp).To(BeEquivalentTo(endTime.Unix()))
		})
	})

	Describe("Custom Day expiration", func() {
		It("should get expiration", func() {
			formatedStartTime := time.Now().AddDate(0, 0, 1).Format("20060102")
			formatedEndTime := time.Now().AddDate(0, 0, 2).Format("20060102")
			expireString := fmt.Sprintf("leaderboard_from%sto%s", formatedStartTime, formatedEndTime)
			exp, err := expiration.GetExpireAt(expireString)
			Expect(err).NotTo(HaveOccurred())

			startTime, err := time.Parse("20060102", formatedStartTime)
			endTime, err := time.Parse("20060102", formatedEndTime)

			ts := endTime.Add(endTime.Sub(startTime)).Unix()
			Expect(exp).To(BeEquivalentTo(ts))
		})

		It("should return error for invalid timestamp", func() {
			exp, err := expiration.GetExpireAt("leaderboard_from20201039to20201011")
			Expect(exp).To(BeEquivalentTo(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("parsing time \"20201039\": day out of range"))

			exp, err = expiration.GetExpireAt("leaderboard_from20201010to20201139")
			Expect(exp).To(BeEquivalentTo(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("parsing time \"20201139\": day out of range"))
		})

		It("should return error for negative duration", func() {
			exp, err := expiration.GetExpireAt("leaderboard_from20201011to20201010")
			Expect(exp).To(BeEquivalentTo(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("has invalid duration -86400"))
		})

		It("should return error if duration is 0", func() {
			exp, err := expiration.GetExpireAt("leaderboard_from20201010to20201010")
			Expect(exp).To(BeEquivalentTo(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("has invalid duration 0"))
		})
	})

	Describe("Unix Timestamp expiration", func() {
		It("should get expiration", func() {
			start := time.Now()
			end := time.Now().Add(time.Hour * 24)
			exp, err := expiration.GetExpireAt(fmt.Sprintf("leaderboard_from%dto%d", start.Unix(), end.Unix()))
			Expect(err).NotTo(HaveOccurred())

			ts := end.Add(end.Sub(start)).Unix()
			Expect(exp).To(BeEquivalentTo(ts))
		})

		It("should get invalid expiration if timestamps reversed", func() {
			start := time.Now()
			end := time.Now().Add(time.Hour * 24)
			exp, err := expiration.GetExpireAt(fmt.Sprintf("leaderboard_from%dto%d", end.Unix(), start.Unix()))
			Expect(exp).To(BeEquivalentTo(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("has invalid duration -86400"))
		})

		It("should get invalid expiration if same timestamps", func() {
			start := time.Now()
			exp, err := expiration.GetExpireAt(fmt.Sprintf("leaderboard_from%dto%d", start.Unix(), start.Unix()))
			Expect(exp).To(BeEquivalentTo(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("has invalid duration 0"))
		})
	})

	Describe("Montly expiration", func() {
		It("should get monthly expiration", func() {
			now := time.Now().UTC()
			year := now.Year()
			month := now.Month()
			monthS := fmt.Sprintf("%d", month)
			if month < 10 {
				monthS = fmt.Sprintf("0%s", monthS)
			}
			exp, err := expiration.GetExpireAt(fmt.Sprintf("leaderboard_year%dmonth%s", year, monthS))
			Expect(err).NotTo(HaveOccurred())

			startTime, _ := time.Parse("200601", fmt.Sprintf("%d%s", year, monthS))
			end := expiration.MonthlyExpiration(startTime)
			ts := end.Unix()
			Expect(exp).To(BeEquivalentTo(ts))
		})

		It("should return error for invalid timestamp", func() {
			exp, err := expiration.GetExpireAt("leaderboard_year2016month99")
			Expect(exp).To(BeEquivalentTo(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("parsing time \"201699\": month out of range"))
		})
	})

	Describe("Weekly expiration", func() {
		It("should get weekly expiration", func() {
			year, week := time.Now().UTC().ISOWeek()
			weekS := fmt.Sprintf("%d", week)
			if week < 10 {
				weekS = fmt.Sprintf("0%s", weekS)
			}
			exp, err := expiration.GetExpireAt(fmt.Sprintf("leaderboard_year%dweek%s", year, weekS))
			Expect(err).NotTo(HaveOccurred())

			twoWeeksFromNow := expiration.WeeklyExpiration(int64(year), int64(week))
			ts := twoWeeksFromNow.Unix()
			Expect(exp).To(BeEquivalentTo(ts))
		})
	})

	Describe("Quarter expiration", func() {
		It("should get quarter expiration", func() {
			now := time.Now().UTC()
			year := now.Year()
			quarter := int(now.Month())/3 + 1
			exp, err := expiration.GetExpireAt(fmt.Sprintf("leaderboard_year%dquarter0%d", year, quarter))
			Expect(err).NotTo(HaveOccurred())

			end := expiration.QuarterlyExpiration(int64(year), int64(quarter))
			ts := end.Unix()
			Expect(exp).To(BeEquivalentTo(ts))
		})
	})
})
