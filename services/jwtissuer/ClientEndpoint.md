# ClientEndpoint API Documentation

## ClientEndpoint API Endpoints Overview

| Method | Endpoint           | Description                                           |
|--------|---------------------|-------------------------------------------------------|
| GET    | `/initialize` | Initialize a WebSocket connection and generate tokens |
| POST   | `/refresh`    | Refresh access token using a refresh token            |
| POST   | `/transfer`   | Generate new tokens and replace the current refresh token |

### GET `/initialize`

#### Description:
This endpoint handles WebSocket connection upgrades, generates access and refresh tokens, and sends the authorization message to the client.

#### Possible Status Codes:

- `101 Switching Protocols`: Successfully upgraded the HTTP connection to a WebSocket connection.
- `400 Bad Request`: The request is invalid or missing required parameters.
- `401 Unauthorized`: The authorization header is missing or invalid.
- `500 Internal Server Error`: A server error occurred while generating tokens or upgrading the connection.

#### Response:
WebSocket connection is established, and the following JSON message is sent over the WebSocket:
```json
{
  "type": "Authorization",
  "content": {
    "refresh_token": "string",
    "access_token": "string",
    "expires_at": 1234567890
  }
}
```

| Field Name   | Type   | Description                                    |
|--------------|--------|------------------------------------------------|
| `type`       | string | Type of message ("Authorization")              |
| `content`    | object | Object containing authorization details        |
| `refresh_token` | string | The generated refresh token                    |
| `access_token`  | string | The generated access token                     |
| `expires_at`    | int64  | Expiration timestamp of the access token (in seconds) |

---

### POST `/refresh`

#### Description:
This endpoint generates a new access token based on the provided refresh token. If the refresh token is invalid or expired, it returns an unauthorized error.

#### Headers:
- `Authorization: <refresh_token>`

#### Possible Status Codes:

- `200 OK`: Successfully generated a new access token.
- `400 Bad Request`: The request is invalid or missing required parameters.
- `401 Unauthorized`: The refresh token is invalid or expired.
- `500 Internal Server Error`: A server error occurred while generating the new access token.

#### Request Body:
- None (the refresh token is expected to be in the Authorization header).

#### Response Body:
```json
{
  "access_token": "string",
  "expires_at": 1234567890
}
```

| Field Name   | Type   | Description                                    |
|--------------|--------|------------------------------------------------|
| `access_token`  | string | The newly generated access token               |
| `expires_at`    | int64  | Expiration timestamp of the new access token (in seconds) |

---

### POST `/transfer`

#### Description:
This endpoint generates a new refresh token and access token based on an existing refresh token. The endpoint will delete the old refresh token and store the new one on the server.

#### Headers:
- `Authorization: <refresh_token>`

#### Possible Status Codes:

- `200 OK`: Successfully generated and returned new refresh and access tokens.
- `400 Bad Request`: The request is invalid or missing required parameters.
- `401 Unauthorized`: The provided refresh token is invalid or expired.
- `500 Internal Server Error`: An error occurred while generating the new tokens.

#### Request Body:
- None (the refresh token is expected to be in the Authorization header).

#### Response Body:
```json
{
  "refresh_token": "string",
  "access_token": "string",
  "expires_at": 1234567890
}
```

|Field Name |Type |Description |
|--------------|--------|------------------------------------------------|
|`refresh_token` |string |The newly generated refresh token |
|`access_token` |string |The newly generated access token |
|`expires_at` |int64 |Expiration timestamp of the new access token (in seconds) |
