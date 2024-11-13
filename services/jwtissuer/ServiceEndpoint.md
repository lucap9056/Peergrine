# gRPC API Documentation

## Overview

This document describes the gRPC API for the `AuthService`, which provides methods for verifying access tokens.

## Service Definition

### `AuthService`

The `AuthService` provides methods for verifying access tokens.

#### Methods

- **VerifyAccessToken**
  - **Description**: Verifies the provided access token and returns token information and verification status.
  - **Request**: `AccessTokenRequest`
  - **Response**: `TokenResponse`

## Protobuf Definitions

### `AccessTokenRequest`

The `AccessTokenRequest` message is used to send the access token and issuer information to the `VerifyAccessToken` method.

```protobuf
message AccessTokenRequest {
  string iss = 1;           // Issuer of the JWT token
  string access_token = 2;  // The access token to be verified
}
```

| Field Name    | Type   | Description                           |
|---------------|--------|---------------------------------------|
| `iss`         | string | Issuer of the JWT token.              |
| `access_token`| string | The access token to be verified.      |

### `TokenResponse`

The `TokenResponse` message contains the details of the token and the result of the verification process.

```protobuf
message TokenResponse {
  string iss = 1;      // Issuer of the JWT token
  int64 iat = 2;       // Issuance time (Unix timestamp)
  int64 exp = 3;       // Expiration time (Unix timestamp)
  string user_id = 4;  // User ID associated with the token
}
```

| Field Name | Type   | Description                                   |
|------------|--------|-----------------------------------------------|
| `iss`      | string | Issuer of the JWT token.                     |
| `iat`      | int64  | Issuance time of the token (Unix timestamp). |
| `exp`      | int64  | Expiration time of the token (Unix timestamp).|
| `user_id`  | string | User ID associated with the token.           |

## Usage

To verify an access token, clients send a `VerifyAccessToken` RPC call with the `AccessTokenRequest` message. The service responds with the `TokenResponse` message indicating the result of the verification and details about the token.

### Example

#### Request

```protobuf
message AccessTokenRequest {
  iss: "exampleIssuer",
  access_token: "yourAccessToken"
}
```

#### Response

```protobuf
message TokenResponse {
  iss: "exampleIssuer",
  iat: 1633024800,
  exp: 1633111200,
  user_id: "user123"
}
```

This document provides an overview of the `AuthService` gRPC API, detailing the request and response messages used for token verification.