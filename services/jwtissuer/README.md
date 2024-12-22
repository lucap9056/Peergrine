# Peergrine JWTIssuer Service

The **JWTIssuer** service is designed to manage token generation and validation, utilizing various configuration options including environment variables, command-line parameters, and settings stored in Zookeeper. The service initializes with default values and supports custom configurations based on deployment needs.

----

## Configuration Overview

MsgBridge service settings can be loaded from:
1. **Zookeeper**: Reads settings from a specified configuration path in Zookeeper.
2. **Environment Variables**: Sets configuration options via environment variables.
3. **Command-Line Arguments**: Overrides configuration options using command-line arguments.

### Configuration Options

|**Setting Field** |**Description** |**Default Value** |
|-|-|-|
|`APP_ID` |Unique service identifier (optional) |Randomly generated |
|`APP_CLIENTENDPOINT_ADDR` |Client endpoint address (optional) |`:80` |
|`APP_SERVICEENDPOINT_ADDR` |Service endpoint address (optional) |`:50051` |
|`APP_REDIS_ADDR` |Redis server address (optional) |None (no Redis used) |
|`APP_BEARER_TOKEN_DURATION` |Bearer token validity duration (seconds, optional) |`3600` (1 hour) |
|`APP_REFRESH_TOKEN_DURATION` |Refresh token validity duration (seconds, optional) |`7200` (2 hours) |
|`APP_PULSAR_ADDRS` |List of Pulsar broker addresses (optional, comma-separated) |None |
|`APP_PULSAR_TOPIC` |Pulsar topic name for communication (optional) |None |
|`APP_ZOOKEEPER_ADDRS` |List of Zookeeper server addresses (optional, comma-separated) |None |
|`APP_CONFIG_PATH` |Configuration path in Zookeeper (optional) |None |

----

## Zookeeper Configuration

The service supports centralized configuration management using Zookeeper. If Zookeeper addresses and paths are provided, the service will attempt to retrieve its configuration from Zookeeper.

### Zookeeper Path

- The default configuration path in Zookeeper is `/jwtissuer`.

- You can override this path by setting the `APP_CONFIG_PATH` environment variable.

----

### Zookeeper Initialization

If the specified configuration path does not exist in Zookeeper, the service will automatically create it and populate it with default settings. This ensures that the service can start even without manual configuration.

----

## Examples
### Setting via Environment Variables
```
export APP_ID="jwt-service-1234"
export APP_CLIENTENDPOINT_ADDR=":8080"
export APP_SERVICEENDPOINT_ADDR=":9090"
export APP_BEARER_TOKEN_DURATION="3600"
export APP_REFRESH_TOKEN_DURATION="7200"
export APP_PULSAR_ADDRS="pulsar://pulsar-broker-0:6650,pulsar://pulsar-broker-1:6650"
expoort APP_PULSAR_TOPIC="JwtIssuer"
export APP_ZOOKEEPER_ADDRS="zookeeper1:2181,zookeeper2:2181"
export APP_CONFIG_PATH="/jwtissuer"
```
### Setting via Command-Line Parameters

```
./app -zookeeper-addrs "zookeeper1:2181,zookeeper2:2181" -config-path "/jwtissuer"
```
This flexibility in configuration options allows the JWTIssuer service to be adapted to various deployment scenarios, whether using standalone setups or integrated within a distributed environment.