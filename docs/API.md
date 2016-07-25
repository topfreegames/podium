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
          "publicID": [string]  // member public id
          "score":    [int],    // member updated score
          "rank":     [int],    // member current rank in leaderboard
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

  ### Get a member score and rank
  `GET /l/:leaderboardID/members/:memberPublicID`

  Gets a member score and rank within a leaderboard.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and memberPublicID should be a unique identifier for the desired member.

  * Success Response
    * Code: `200`
    * Content:
      ```
      {
        "success": true,
        "publicID": [string]  // member public id
        "score":    [int],    // member updated score
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

  ### Remove member from leaderboard
  `DELETE /l/:leaderboardID/members/:memberPublicID`

  Removes specified member from leaderboard. If member is not in leaderboard, do nothing.

  Leaderboard ID should be a valid [leaderboard name](leaderboard-names.html) and memberPublicID should be a unique identifier for the member being removed.

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

  ### Get a member rank
  `GET /l/:leaderboardID/members/:memberPublicID/rank`

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
          },
          {
            "leaderboardID": [string] // leaderboard where this score was set
            "publicID": [string]      // member public id
            "score":    [int],        // member updated score
            "rank":     [int],        // member current rank in leaderboard
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
