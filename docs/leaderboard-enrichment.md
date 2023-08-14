**Leaderboard Enrichment**

Leaderboard Enrichment is a feature in the Podium service that allows you to enhance the information retrieved from read operations on leaderboards by registering a webhook. 
By utilizing this feature, you can dynamically add metadata to each leaderboard member, providing additional details without the need for additional API requests. 
This documentation will guide you through the process of setting up and using Leaderboard Enrichment.

## Configuration Setup

To enable Leaderboard Enrichment, follow these steps:

1. Open your Podium configuration file or set the necessary environment variables.

```yaml
webhook_urls:
    "{TENANT_ID}": "{BASE_WEBHOOK_URL}"
```

> Replace `{TENANT_ID}` with the unique identifier of the tenant and `{BASE_WEBHOOK_URL}` with the base URL of the webhook endpoint.

## Webhook Endpoint

The webhook endpoint, denoted by `/leaderboards/enrich`, will be called by the Podium service to retrieve additional metadata for leaderboard members. The endpoint should be set up on your server to handle incoming requests.

### Request

The webhook endpoint will receive a `POST` request with the following JSON body:

```json
{
  "members": [
    {
      "leaderboard_id": "leaderboard_id",
      "id": "id"
    }
  ]
}
```

- `"leaderboard_id"`: The unique identifier of the leaderboard.
- `"member_public_id"`: The public identifier of the member.

### Response

The webhook endpoint is expected to return a `JSON` response with metadata for the specified member. Here's an example response:

```json
{
  "members": [
    {
      "leaderboard_id": "leaderboard_id",
      "id": "id",
      "metadata": {
        "custom_field1": "value1",
        "custom_field2": "value2"
      }
    }
  ]
}
```

- `"publicID"`: The public identifier of the member (must match the request).
- `"metadata"`: Additional metadata fields to enrich the member's information.

## Enabling Enrichment

Once the webhook endpoint is set up, you will need to add the information to your header when making read requests to the Podium API:

```json
"tenant-id": "tenant-id",
```

Podium will automatically call the endpoint for each read operation that retrieves information about leaderboard members. The enriched metadata will be included in the response, enhancing the details available for each member. If the corresponding configuration for the tenant-id sent is not found, of if none is specified, Podium will  return the response without any metadata.