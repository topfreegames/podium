Library
=======

For detailed information, check our [reference](https://godoc.org/github.com/topfreegames/podium/leaderboard).
All examples below have imported the leaderboard module using:

```
import "github.com/topfreegames/podium/leaderboard"
```

## Creating the client

```
const host = "localhost"
const port = 6379
const password = ""
const db = 0
const connectionTimeout = 200

leaderboards, err := leaderboard.NewClient(host, port, password, db, connectionTimeout)
```

## Creating, updating or retrieving member scores

```
const leaderboardID = "lbID"
const playerID = "playerID"
const score = 100
const wantToKnowPreviousRank = false //do I want to receive also the previous rank on the user?
const scoreTTL = "100"               //how many seconds my score will be kept on the leaderboard

member, err := leaderboards.SetMemberScore(context.Background(), leaderboardID, playerID, score, wantToKnowPreviousRank, scoreTTL)
if err != nil {
    return err
}

playerPrinter := func(publicID string, score int64, rank int) {
    fmt.Printf("Player(id: %s, score: %d rank: %d)\n", publicID, score, rank)
}

playerPrinter(member.PublicID, member.Score, member.Rank)

const order = "desc"     //if set to asc, will treat the ranking with ascending scores (less is best)
const includeTTL = false //if set to true, will return the member's score expiration unix timestamp
member, err = leaderboards.GetMember(context.Background(), leaderboardID, playerID, order, includeTTL)
if err != nil {
    return err
}

playerPrinter(member.PublicID, member.Score, member.Rank)
```

## Setting and getting member scores in bulk

```
const leaderboardID = "lbID"

players := leaderboard.Members{
    &leaderboard.Member{Score: 1000, PublicID: "playerA"},
    &leaderboard.Member{Score: 2000, PublicID: "playerB"},
}

err := leaderboards.SetMembersScore(context.Background(), leaderboardID, players, false, "")
if err != nil {
    return err
}

const order = "desc"     //if set to asc, will treat the ranking with ascending scores (less is best)
const includeTTL = false //if set to true, will return the member's score expiration unix timestamp
members, err := leaderboards.GetMembers(context.Background(), leaderboardID, []string{"playerA", "playerB"}, order, includeTTL)

for _, member := range members {
    fmt.Printf("Player(id: %s, score: %d rank: %d)\n", member.PublicID, member.Score, member.Rank)
}
```

## Retrieving leaderboard leaders

```
const leaderboardID = "myleaderboardID"

players := leaderboard.Members{
    &leaderboard.Member{Score: 10, PublicID: "player1"},
    &leaderboard.Member{Score: 20, PublicID: "player2"},
}

err = leaderboards.SetMembersScore(context.Background(), leaderboardID, players, false, "")
if err != nil {
    log.Fatalf("leaderboards.SetMembersScore failed: %v", err)
}
const pageSize = 10
const pageIdx = 1 //starts at 1
leaders, err := leaderboards.GetLeaders(context.Background(), leaderboardID, pageSize, pageIdx, "desc")
if err != nil {
    log.Fatalf("leaderboards.GetLeaders failed: %v", err)
}

for _, player := range leaders {
    fmt.Printf("Player(id: %s, score: %d rank: %d)\n", player.PublicID, player.Score, player.Rank)
}
```

## Incrementing player scores

```
const leaderboardID = "lbID"
const playerID = "playerA"
const scoreIncrement = 500
const scoreTTL = ""
member, err := leaderboards.IncrementMemberScore(context.Background(), leaderboardID, playerID, scoreIncrement,
    scoreTTL)
if err != nil {
    return err
}

fmt.Printf("Player(id: %s, score: %d rank: %d)\n", member.PublicID, member.Score, member.Rank)
```

## Number of players on a leaderboard

```
const leaderboardID = "lbID"
count, err := leaderboards.TotalMembers(context.Background(), leaderboardID)
if err != nil {
    return err
}
fmt.Printf("Total number of players on leaderboard %s: %d\n", leaderboardID, count)
```

## Removing players from a leaderboard

```
const leaderboardID = "lbID"
const playerIdToRemove = "playerID"
err := leaderboards.RemoveMember(context.Background(), leaderboardID, playerIdToRemove) //removing a single player
if err != nil {
    return err
}

playerIDsToRemove := make([]interface{}, 2)
playerIDsToRemove[0] = "playerA"
playerIDsToRemove[1] = "playerB"

err = leaderboards.RemoveMembers(context.Background(), leaderboardID, playerIDsToRemove) //removing multiple players
if err != nil {
    return err
}
```

## Total pages of a leaderboard

```
const leaderboardID = "lbID"
const pageSize = 10
pageCount, err := leaderboards.TotalPages(context.Background(), leaderboardID, pageSize)
if err != nil {
    return err
}
fmt.Printf("total pages: %d\n", pageCount)
```

## Getting players around a player

```
const leaderboardID = "lbID"
const pageSize = 10
const getLastIfNotFound = false //if set to true, will treat members not in ranking as being in last position
//if set to false, will return 404 when the member is not in the ranking
const order = "asc"
members, err := leaderboards.GetAroundMe(context.Background(), leaderboardID, pageSize, "playerID",
    order, getLastIfNotFound)
if err != nil {
    return err
}
for _, member := range members {
    fmt.Printf("Player(id: %s, score: %d rank: %d)\n", member.PublicID, member.Score, member.Rank)
}
```

## Getting players around a score

```
const leaderboardID = "lbID"
const pageSize = 10
const score = 1500
const order = "desc"

members, err := leaderboards.GetAroundScore(context.Background(), leaderboardID, pageSize, score, order)
if err != nil {
    return err
}
for _, member := range members {
    fmt.Printf("Player(id: %s, score: %d rank: %d)\n", member.PublicID, member.Score, member.Rank)
}
```

## Top percentage of a leaderboard

```
const leaderboardID = "lbID"
const pageSize = 10
const percent = 10
const maxMembersToReturn = 100
const order = "asc"
top10, err := leaderboards.GetTopPercentage(context.Background(), leaderboardID, pageSize, percent,
    maxMembersToReturn, order)
if err != nil {
    return err
}
for _, member := range top10 {
    fmt.Printf("Player(id: %s, score: %d rank: %d)\n", member.PublicID, member.Score, member.Rank)
}
```

## Getting members in a range

```
const leaderboardID = "lbID"
const startOffset = 0
const endOffset = 10
const order = "asc"
members, err := leaderboards.GetMembersByRange(context.Background(), leaderboardID, startOffset, endOffset, order)
if err != nil {
    return err
}
for _, member := range members {
    fmt.Printf("Player(id: %s, score: %d rank: %d)\n", member.PublicID, member.Score, member.Rank)
} 
```

## Removing a leaderboard

```
const leaderboardID = "lbID"
err := leaderboards.RemoveLeaderboard(context.Background(), leaderboardID)
```