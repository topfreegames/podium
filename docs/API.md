Podium API
==========

## Healthcheck Routes

  ### Healthcheck

  `GET /healthcheck`

  Validates that the app is still up, including the connection to Redis.

  * Success Response
    * Code: `200`
    * Content:

      ```
        "WORKING"
      ```

  * Error Response

    It will return an error if it failed to connect to Redis.

    * Code: `500`
    * Content:

      ```
        "<error-details>"
      ```

## Status Routes

  ### Status

  `GET /status`

  Returns statistics on the health of Podium.

  * Success Response
    * Code: `200`
    * Content:

      ```
        {
          "app": {
            "errorRate": [float]        // Exponentially Weighted Moving Average Error Rate
          },
        }
      ```

## Leaderboard Routes

  ### Create or Update a Member Score
  `PUT /l/:leaderboardID/members/:memberPublicID/score`

  ##### optional query string
  * prevRank=[true|false]
    * if set to true, it will also return the previous rank of the player in the leaderboard, -1 if the player didn't exist in the leaderboard
    * e.g. `PUT /l/:leaderboardID/members/:memberPublicID/score?prevRank=true`
    * defaults to "false"
  * scoreTTL=[integer]
    * if set, the score of the player will be expired from the leaderboard past [integer] seconds if it does not update it within this interval
    * e.g. `PUT /l/:leaderboardID/members/:memberPublicID/score?scoreTTL=100`
    * defaults to none (the score will never expire)

  Atomically creates a new member within a leaderboard or if member already exists in leaderboard, update their score.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and memberPublicID should be a unique identifier for the member associated with the score.

  * Payload

    ```
    {
      "score":      [integer]  // Integer representing member score
    }
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "member": {
          "publicID":     [string]  // member public id
          "score":        [int]     // member updated score
          "rank":         [int]     // member current rank in leaderboard
          "previousRank": [int]     // the previous rank of the player in the leaderboard, if requests
          "expireAt":     [int]     // unix timestamp of when the score will be expired, if scoreTTL is sent
        }
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Create or Update many Members Score
  `PUT /l/:leaderboardID/scores`

  ##### optional query string
  * prevRank=[true|false]
    * if set to true, it will also return the previous rank of the player in the leaderboard, -1 if the player didn't exist in the leaderboard
    * e.g. `PUT /l/:leaderboardID/scores?prevRank=true`
    * defaults to "false"
  * scoreTTL=[integer]
    * if set, the score of the player will be expired from the leaderboard past [integer] seconds if it does not update it within this interval
    * e.g. `PUT /l/:leaderboardID/scores?scoreTTL=100`
    * defaults to none (the score will never expire)

  Atomically creates many new members within a leaderboard or if some members already exists in leaderboard, update their scores.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and publicID should be a unique identifier for the member associated with the score.

  * Payload

    ```
    {
      "members": [{
          "publicID": [string]  // member public id
          "score":    [int],    // member updated score
        }, ...]
    }
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "members": [{
          "publicID":     [string]  // member public id
          "score":        [int]     // member updated score
          "rank":         [int]     // member current rank in leaderboard
          "previousRank": [int]     // the previous rank of the player in the leaderboard, if requests
          "expireAt":     [int]     // unix timestamp of when the score will be expired, if scoreTTL is sent
        }, ...]
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Increment a Member Score
  `PATCH /l/:leaderboardID/members/:memberPublicID/score`

  ##### optional query string
  * scoreTTL=[integer]
    * if set, the score of the player will be expired from the leaderboard past [integer] seconds if it does not update it within this interval
    * e.g. `PUT /l/:leaderboardID/members/:memberPublicID/score?scoreTTL=100`
    * defaults to none (the score will never expire)

  Atomically creates a new member within a leaderboard with the given increment as score. If member already exists in leaderboard just increment their score.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and memberPublicID should be a unique identifier for the member associated with the score.

  **WARNING:** Incrementing a member score by 0 is not a valid operation and will return a 400 Bad Request result.

  * Payload

    ```
    {
      "increment":      [integer]  // Integer representing increment in member score
    }
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "member": {
          "publicID": [string]  // member public id
          "score":    [int]     // member updated score
          "rank":     [int]     // member current rank in leaderboard
          "expireAt": [int]     // unix timestamp of when the score will be expired, if scoreTTL is sent
        }
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```


  ### Remove a leaderboard
  `DELETE /l/:leaderboardID`

  Remove the entire leaderboard from Podium.

  **WARNING: This operation cannot be undone and all the information in the leaderboard will be destroyed.**

  `leaderboardID` should be a valid [leaderboard name](leaderboard-names.html).

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
      }
      ```

  * Error Response

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Get a member score and rank
  `GET /l/:leaderboardID/members/:memberPublicID`

  ##### optional query string
  * order=[asc|desc]
    * if set to asc, will treat the ranking with ascending scores (less is best)
    * e.g. `GET /l/:leaderboardID/members/:memberPublicID?order=asc`
    * defaults to "desc"
  * scoreTTL=[true|false]
    * if set to true, will return the member's score expiration unix timestamp
    * e.g. `GET /l/:leaderboardID/members/:memberPublicID?scoreTTL=true`
    * defaults to "false"

  Gets a member score and rank within a leaderboard.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and memberPublicID should be a unique identifier for the desired member.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "publicID": [string]  // member public id
        "score":    [int]     // member updated score
        "rank":     [int]     // member current rank in leaderboard
        "expireAt": [int]     // unix timestamp of when the member's score will be erased (only if scoreTTL is true)
      }
      ```

  * Error Response

    It will return an error if the member is not found.

    * Code: `404`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Get multiple member scores and rank
  `GET /l/:leaderboardID/members?ids=publicIDcsv`

  ##### optional query string
  * order=[asc|desc]
    * if set to asc, will treat the ranking with ascending scores (less is best)
    * e.g. `GET /l/:leaderboardID/members?ids=publicIDcsv?order=asc`
    * defaults to "desc"
  * scoreTTL=[true|false]
    * if set to true, will return the member's score expiration unix timestamp
    * e.g. `GET /l/:leaderboardID/members?ids=publicIDcsv?scoreTTL=true`
    * defaults to "false"


  Gets multiple members' score and ranks within a leaderboard.

  If any public IDs are not found, they will be returned in the `notFound` list in the response. This is so a list of all the desired members (i.e.: player's friends) can be retrieved and only the ones in the leaderboard get returned.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and publicIDcsv should be a comma-separated list of the desired members Public IDs.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "members": [
          {
            "publicID": [string]    // member public id
            "rank":     [int]       // member rank in the specific leaderboard
            "position": [int]       // member rank for all members returned in this request
            "score":    [int]       // member score in the leaderboard
            "expireAt": [int]       // unix timestamp of when the member's score will be erased (only if scoreTTL is true)
          }
        ],
        "notFound": [
          "[string]"                // list of public ids that were not found in the leaderboard
        ],
        "success": true
      }
      ```

  * Error Response

    It will return an error if a list of member ids is not supplied.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Remove members from leaderboard
  `DELETE /l/:leaderboardID/members?ids=memberPublicID1,memberPublicID2,...`

  Removes specified members from leaderboard. If a member is not in leaderboard, do nothing.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and ids should be a list of unique identifier for the members being removed, separated by commas.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
      }
      ```

  * Error Response

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```


  ### Get a member score and rank in many leaderboards
  `GET /m/:memberPublicID/scores?leaderboardIds=leaderboard1,leaderboard2,...`

  ##### optional query string
  * order=[asc|desc]
    * if set to asc, will treat the ranking with ascending scores (less is best)
    * e.g. `GET /m/:memberPublicID/scores?leaderboardIds=leaderboard1,leaderboard2,...?order=asc`
    * defaults to "desc"
  * scoreTTL=[true|false]
    * if set to true, will return the member's score expiration unix timestamp
    * e.g. `GET /m/:memberPublicID/scores?leaderboardIds=leaderboard1,leaderboard2,...?scoreTTL=true`
    * defaults to "false"

  Get a member score and rank within many leaderboards.

  Leaderboard Ids should be valid leaderboard names separated by commas.

  * Sucess Response

    * Code: 200
    * Content:
      ```
      {
        "scores": [
          {
            "leaderboardID": "teste",
            "rank": 1,
            "score": 100,
            "expireAt": [int]     // unix timestamp of when the member's score will be erased (only if scoreTTL is true)
          },
          {
            "leaderboardID": "teste2",
            "rank": 1,
            "score": 100,
            "expireAt": [int]     // unix timestamp of when the member's score will be erased (only if scoreTTL is true)
          }
        ],
        "success": true
      }
      ```

  * Error Response

    * Code: 500
    * Content:
    ```
    {
      "reason": "Could not find data for member teste3 in leaderboard teste3.",
      "success": false
    }
    ```

  ### Get a member rank
  `GET /l/:leaderboardID/members/:memberPublicID/rank`

  ##### optional query string
  * order=[asc|desc]
    * if set to asc, will treat the ranking with ascending scores (less is best)
    * e.g. `GET /l/:leaderboardID/members/:memberPublicID/rank?order=asc`
    * defaults to "desc"

  Gets a member rank within a leaderboard.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and memberPublicID should be a unique identifier for the desired member.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "publicID": [string]  // member public id
        "rank":     [int],    // member current rank in leaderboard
      }
      ```

  * Error Response

    It will return an error if the member is not found.

    * Code: `404`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Get members around a member
  `GET /l/:leaderboardID/members/:memberPublicID/around?pageSize=10`

  ##### optional query string
  * order=[asc|desc]
    * if set to asc, will treat the ranking with ascending scores (less is best)
    * e.g. `GET /l/:leaderboardID/members/:memberPublicID/around?pageSize=10?order=asc`
    * defaults to "desc"
  * getLastIfNotFound=[true|false]
    * if set to true, will return the last members of the ranking when the member is not in the ranking
    * if set to false, will return 404 when the member is not in the ranking
    * e.g. `GET /l/:leaderboardID/members/:memberPublicID/around?getLastIfNotFound=true`
    * defaults to "false"

  Gets a list of members with ranking around that of the specified member within a leaderboard.

  The `pageSize` querystring parameter specifies the number of members that will be returned from this operation. This means that `pageSize/2` members will be above the specified member and the other `pageSize/2` will be below.

  Podium will compensate if no more members can be found above or below (first or last member in the leaderboard ranking) to ensure that the desired number of members is returned (up to the number of members in the leaderboard).

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and memberPublicID should be a unique identifier for the desired member.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "members": [
          {
            "publicID": [string]  // member public id
            "score":    [int],    // member updated score
            "rank":     [int],    // member current rank in leaderboard
          },
          {
            "publicID": [string]  // member public id
            "score":    [int],    // member updated score
            "rank":     [int],    // member current rank in leaderboard
          },
          //...
        ],
      }
      ```

  * Error Response

    It will return an error if the member is not found.

    * Code: `404`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```
  ### Get members around a score
  `GET /l/:leaderboardID/scores/:score/around?pageSize=10`

  ##### optional query string
  * order=[asc|desc]
    * if set to asc, will treat the ranking with ascending scores (less is best)
    * e.g. `GET /l/:leaderboardID/scores/:score/around?pageSize=10?order=asc`
    * defaults to "desc"

  Gets a list of members with score around that of the specified specified in the request. If the `score` parameter falls outside the leaderboard [minScore, maxScore], it will return the bottom/top rank members in the leaderboard, respectively.

  The `pageSize` querystring parameter specifies the number of members that will be returned from this operation. That means there will be around `pageSize/2` (+-1) members with score above the specified score, and `pageSize/2`(+-1) with score below.

  Podium will compensate if no more members can be found above or below (first or last member in the leaderboard ranking) to ensure that the desired number of members is returned (up to the number of members in the leaderboard).

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and `score` should be a valid number.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "members": [
          {
            "publicID": [string]  // member public id
            "score":    [int],    // member updated score
            "rank":     [int],    // member current rank in leaderboard
          },
          {
            "publicID": [string]  // member public id
            "score":    [int],    // member updated score
            "rank":     [int],    // member current rank in leaderboard
          },
          //...
        ],
      }
      ```

  * Error Response

    It will return an error if the leaderboard is not found or the request has invalid parameters.

    * Code: `404`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Get the number of members in a leaderboard
  `GET /l/:leaderboardID/members-count/`

  Gets the number of members in a leaderboard.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html).

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "count": [int],      // number of members in leaderboard
      }
      ```

  * Error Response

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Get the top N members in a leaderboard (by page)
  `GET /l/:leaderboardID/top/:pageNumber?pageSize=:pageSize`

  ##### optional query string
  * order=[asc|desc]
    * if set to asc, will treat the ranking with ascending scores (less is best)
    * e.g. `GET /l/:leaderboardID/top/:pageNumber?pageSize=:pageSize?order=asc`
    * defaults to "desc"

  Gets the top N members in a leaderboard, by page.

  `leaderboardID` should be a valid [leaderboard name](leaderboard-names.html), `pageNumber` is the current page you are looking for and `pageSize` is the number of members per page that will be returned.

  This means that if you want the top 20 members, you'll call `/l/my-leaderboard/top/1?pageSize=20` for the first 20, `/l/my-leaderboard/top/2?pageSize=20` for members 21-40 and so on.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "members": [
          {
            "publicID": [string]  // member public id
            "score":    [int],    // member updated score
            "rank":     [int],    // member current rank in leaderboard
          },
          {
            "publicID": [string]  // member public id
            "score":    [int],    // member updated score
            "rank":     [int],    // member current rank in leaderboard
          },
          //...
        ]
      }
      ```

  * Error Response

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

  ### Get the top x% members in a leaderboard
  `GET /l/:leaderboardID/top-percent/:percentage`

  ##### optional query string
  * order=[asc|desc]
    * if set to asc, will treat the ranking with ascending scores (less is best)
    * e.g. `GET /l/:leaderboardID/top-percent/:percentage?order=asc`
    * defaults to "desc"

  Gets the top x% members in a leaderboard.

  `leaderboardID` should be a valid [leaderboard name](leaderboard-names.html), `percentage` is the % of members you want to return.

  The number of members is bound by the configuration `api.maxReturnedMembers`, that defaults to 2000 members.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "members": [
          {
            "publicID": [string]  // member public id
            "score":    [int],    // member updated score
            "rank":     [int],    // member current rank in leaderboard
          },
          {
            "publicID": [string]  // member public id
            "score":    [int],    // member updated score
            "rank":     [int],    // member current rank in leaderboard
          },
          //...
        ]
      }
      ```

  * Error Response

    If the percentage is not a valid integer between 1 and 100, you'll get a 400.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

## Member Routes

  ### Create or update score for a member in several leaderboards
  `PUT /m/:memberPublicID/scores`

  ##### optional query string
  * prevRank=[true|false]
    * if set to true, it will also return the previous rank of the player in the leaderboard, -1 if the player didn't exist in the leaderboard
    * e.g. `PUT /l/:leaderboardID/members/:memberPublicID/score?prevRank=true`
    * defaults to "false"
  * scoreTTL=[integer]
    * if set, the score of the player will be expired from the leaderboards past [integer] seconds if it does not update it within this interval
    * e.g. `PUT /l/:leaderboardID/members/:memberPublicID/score?scoreTTL=100`
    * defaults to none (the score will never expire

  Atomically creates a new member within many leaderboard or if member already exists in each leaderboard, updates their score.

  `memberPublicID` should be a unique identifier for the member associated with the score. Each `leaderboardID` should be a valid [leaderboard name](leaderboard-names.html).

  * Payload

    ```
    {
      "score": [integer],                       // Integer representing member score
      "leaderboards": [array of leaderboardID]  // List of all leaderboards to update
    }
    ```

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "scores": [
          {
            "leaderboardID": [string] // leaderboard where this score was set
            "publicID": [string]      // member public id
            "score":    [int],        // member updated score
            "rank":     [int],        // member current rank in leaderboard
            "previousRank": [int]     // the previous rank of the player in the leaderboard, if requests
          },
          {
            "leaderboardID": [string] // leaderboard where this score was set
            "publicID": [string]      // member public id
            "score":    [int],        // member updated score
            "rank":     [int],        // member current rank in leaderboard
            "previousRank": [int]     // the previous rank of the player in the leaderboard, if requests
          },
          //...
        ]
      }
      ```

  * Error Response

    It will return an error if an invalid payload is sent or if there are missing parameters.

    * Code: `400`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```

    * Code: `500`
    * Content:
      ```
      {
        "success": false,
        "reason": [string]
      }
      ```
