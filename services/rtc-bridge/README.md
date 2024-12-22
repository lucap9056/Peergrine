# Peergrine RtcBridge Service

The **RtcBridge** service acts as a WebRTC Signaling Server, enabling efficient communication for WebRTC applications. It relies on **Pulsar** for horizontal scaling and integrates with **Redis** for caching and authentication. The configuration of RtcBridge can be managed via environment variables, command-line arguments, or Zookeeper for centralized management.

----

## Configuration Overview

RtcBridge's settings can be loaded from the following sources:

1. **Zookeeper**: If Zookeeper servers and a configuration path are specified, the service will read its settings from Zookeeper.

2. **Environment Variables**: Configuration options can be set via environment variables.

3. **Command-Line Arguments**: Configuration options can be passed through command-line arguments.

----

## Configuration Options

|**Setting Field** |**Description** |**Default Value** |
|-|-|-|
|`APP_ID` |Unique service identifier (optional) |Randomly generated |
|`APP_ADDR` |Service running address (optional) |`:80` |
|`APP_AUTH_ADDR` |Authentication service address (optional) |None |
|`APP_REDIS_ADDR` |Redis server address (optional) |None |
|`APP_PULSAR_ADDRS` |List of Pulsar broker addresses (optional, comma-separated) |None |
|`APP_PULSAR_TOPIC` |Pulsar topic name for communication (optional) |None |
|`APP_ZOOKEEPER_ADDRS` |List of Zookeeper server addresses (optional, comma-separated) |None |
|`APP_CONFIG_PATH` |Configuration path in Zookeeper (optional) |None |

>**Notes:**
>- If `APP_AUTH_ADDR` is empty but `APP_REDIS_ADDR` is set, the service will attempt to handle authentication independently.

----

## Zookeeper Configuration

Zookeeper allows RtcBridge to read settings from a specified configuration path, which is useful for managing configurations in distributed environments.

### Zookeeper Path

- **Default Path**: `/rtc-bridge`

- The default path can be overridden using the `CONFIG_PATH` environment variable or command-line arguments.

----

### Zookeeper Initialization

If the specified configuration path in Zookeeper does not exist, RtcBridge will automatically create the path and save default settings, ensuring that the service can start even if manual configuration is not performed.

----

## Examples

### Setting Configuration via Environment Variables

```
export APP_ADDR=":8080"
export APP_AUTH_ADDR="auth-service:5000"
export APP_REDIS_ADDR="redis:6379"
export APP_PULSAR_ADDRS="pulsar://pulsar-broker-0:6650,pulsar://pulsar-broker-1:6650"
expoort APP_PULSAR_TOPIC="RTCBridge"
export APP_ZOOKEEPER_ADDRS="zookeeper1:2181,zookeeper2:2181"
export CONFIG_PATH="/rtc-bridge"
```
### Setting Configuration via Command-Line Arguments

```
./app -zookeeper-addrs "zookeeper1:2181,zookeeper2:2181" -config-path "/rtc-bridge"
```

----

## Integration

RtcBridge integrates with:

- **Pulsar**: For high-performance horizontal message delivery.
- **Redis**: For caching and optional independent authentication.
- **Zookeeper**: For centralized configuration management (optional).

RtcBridge offers a scalable and reliable solution for WebRTC signaling, enabling efficient and real-time communication for WebRTC clients.