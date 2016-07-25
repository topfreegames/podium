Meaning of Leaderboard Names
============================

Leaderboard names carry a lot of semantic weight in Podium. Each leaderboard name is composed of two parts: leaderboard name and an optional season suffix.

## Seasonal Leaderboards

If you want a leaderboard to be seasonal and have an expiration, Podium allows you to do it just by adding a suffix to it.

Let's say you want a weekly leaderboard for your Cario Sisters game. You would name that leaderboard `cario-sisters-year2016week01` when reporting scores for the first week, `cario-sisters-year2016week02` when reporting for the next week and so on.

Podium will expire the leaderboard in twice as many time as you provisioned your leaderboard to contain. That means a leaderboard with a week of data will be expired within 2 weeks after it's appointed start.

## Available expirations

Podium supports many different expirations:

* Unix timestamps from and to;
* yyyymmdd timestamps from and to;
* Yearly expiration;
* Quarterly expiration;
* Monthly expiration;
* Weekly expiration.

### Unix Timestamp Expiration

In order to use this type of expiration use leaderboard names like `cario-sisters-from1469487752to1469487753`. This means a leaderboard from the first timestamp to the second timestamp.

This kind of leaderboard has the ultimate flexibility, allow for configuration of a leaderboard duration up to the second. Just remember this is UTC timestamps.

### yyyymmdd Timestamp Expiration

In order to use this type of expiration use leaderboard names like `cario-sisters-from20201010to20201011`. This means a leaderboard from the first timestamp to the second timestamp.

### Yearly Expiration

In order to use this type of expiration use leaderboard names like `cario-sisters-year2016`. This means a leaderboard ranging from 1st of January of 2016 to the 1st of January of 2017(not included).

### Quarterly Expiration

In order to use this type of expiration use leaderboard names like `cario-sisters-year2016quarter01`. This means a leaderboard ranging from 1st of January of 2016 to the 1st of April of 2016(not included).

### Monthly Expiration

In order to use this type of expiration use leaderboard names like `cario-sisters-year2016month03`. This means a leaderboard ranging from 1st of March of 2016 to the 1st of April of 2016(not included).

### Weekly Expiration

In order to use this type of expiration use leaderboard names like `cario-sisters-year2016week21`. This means a leaderboard ranging from the 23rd of May of 2016 to the 30th of May of 2016(not included).

This mode is a little odd as it uses week numbers and Week 1 does not start in the first of january. For more information about week numbers, refer to [this page](https://en.wikipedia.org/wiki/ISO_week_date).
