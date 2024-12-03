# Peergrine JWTIssuer Service

The **JWTIssuer** service is designed to manage token generation and validation, utilizing various configuration options including environment variables, command-line parameters, and settings stored in Zookeeper. The service initializes with default values and supports custom configurations based on deployment needs.

----

## Configuration Overview

The service provides multiple configurable options that can be set using **environment variables** or **command-line parameters**. These include network addresses, token validity durations, Redis and Consul configurations, and Zookeeper paths.

### Configuration Options

|**Setting Field** |**Description** |**Default Value** |
|-|-|-|
|`APP_ID` |Unique service identifier (optional) |Randomly generated |
|`APP_CLIENTENDPOINT_ADDR` |Client endpoint address (optional) |`:80` |
|`APP_SERVICEENDPOINT_ADDR` |Service endpoint address (optional) |`:50051` |
|`APP_REDIS_ADDR` |Redis server address (optional) |None (no Redis used) |
|`APP_BEARER_TOKEN_DURATION` |Bearer token validity duration (seconds, optional) |`3600` (1 hour) |
|`APP_REFRESH_TOKEN_DURATION` |Refresh token validity duration (seconds, optional) |`7200` (2 hours) |
|`APP_CONSUL_ADDR` |Consul service address (optional) |None |
|`APP_SERVICE_NAME` |Service name (optional) |`JWTIssuer` |
|`APP_SERVICE_ADDR` |Service address (optional) |Determined by service |
|`APP_SERVICE_PORT` |Service health check port (optional) |`4000` |
|`APP_ZOOKEEPER_ADDRS` |List of Zookeeper server addresses (comma-separated) |None |
|`APP_CONFIG_PATH` |Configuration path in Zookeeper (optional) |None |

>**Note:** If `APP_CONSUL_ADDR` is not set, the settings `APP_SERVICE_NAME`, `APP_SERVICE_ADDR`, and `APP_SERVICE_PORT` will not be used.

----

## Zookeeper Configuration

The service supports centralized configuration management using Zookeeper. If Zookeeper addresses and paths are provided, the service will attempt to retrieve its configuration from Zookeeper.

### Zookeeper Path

- The default configuration path in Zookeeper is `/jwtissuer`.

- You can override this path by setting the `APP_CONFIG_PATH` environment variable.

### Configuration Workflow
1. **Zookeeper**: If `APP_ZOOKEEPER_ADDRS` is provided, the service attempts to read settings from the specified Zookeeper path. If the path does not exist, default settings are written to the path.
2. **Environment Variables**: Settings can be defined through environment variables, such as `APP_CLIENTENDPOINT_ADDR`.
3. **Command-Line Parameters**: Settings can also be passed as command-line arguments, such as `-zookeeper-addrs`.

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
export APP_ZOOKEEPER_ADDRS="zookeeper1:2181,zookeeper2:2181"
export APP_CONFIG_PATH="/jwtissuer"
```
### Setting via Command-Line Parameters

```
./app -zookeeper-addrs "zookeeper1:2181,zookeeper2:2181" -config-path "/jwtissuer"
```
This flexibility in configuration options allows the JWTIssuer service to be adapted to various deployment scenarios, whether using standalone setups or integrated within a distributed environment.