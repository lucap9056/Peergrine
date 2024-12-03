# Peergrine RtcBridge Service

The **RtcBridge** service acts as a WebRTC Signaling Server, enabling efficient communication for WebRTC applications. It relies on **Kafker** and **Kafka** for horizontal scaling and integrates with **Redis** for caching and authentication. The configuration of RtcBridge can be managed via environment variables, command-line arguments, or Zookeeper for centralized management.

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
|`APP_ID` |Unique identifier for the service (optional) |Randomly generated |
|`APP_ADDR` |Address where the service runs |`:80` |
|`APP_AUTH_ADDR` |Address for the authentication service |Empty |
|`APP_REDIS_ADDR` |Address of the Redis server |Empty |
|`APP_KAFKA_ADDR` |Address of the Kafka server |Empty |
|`APP_KAFKER_ADDR` |Address of the Kafker service |Empty |
|`APP_CONSUL_ADDR` |Address of the Consul service |Empty |
|`APP_SERVICE_NAME` |Name of the service |`RtcBridge` |
|`APP_SERVICE_ADDR` |Service address (optional) |Determined by service |
|`APP_SERVICE_PORT` |Port for health checks |`4000` |

>**Notes:**
- If `APP_CONSUL_ADDR` is not set, `APP_SERVICE_NAME`, `APP_SERVICE_ADDR`, and `APP_SERVICE_PORT` will not be used.
- If `APP_AUTH_ADDR` is empty but `APP_REDIS_ADDR` is set, the service will attempt to handle authentication independently.

----

## Zookeeper Configuration

Zookeeper allows RtcBridge to read settings from a specified configuration path, which is useful for managing configurations in distributed environments.

### Zookeeper Path

- **Default Path**: `/rtc-bridge`

- The default path can be overridden using the `CONFIG_PATH` environment variable or command-line arguments.

----

### Configuration Workflow

1. **Zookeeper**: If a Zookeeper address (`APP_ZOOKEEPER_ADDRS`) is provided, the service will attempt to read settings from the specified Zookeeper path. If the path does not exist, it will initialize with default settings and store them in Zookeeper.

2. **Environment Variables**: Configuration values like `APP_ADDR` or `APP_KAFKA_ADDR` can be set using environment variables.

3. **Command-Line Arguments**: Configuration values can also be passed using command-line arguments, such as `-zookeeper-addrs`.

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
export APP_KAFKA_ADDR="kafka:9092"
export APP_KAFKER_ADDR="kafker:50051"
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

- **Kafka**: For real-time message queuing and event-driven communication.

- **Redis**: For caching and authentication handling.

- **Kafker**: To manage Kafka partition assignments.

- **Zookeeper**: For centralized configuration management.

- **Consul**: For service registration and discovery (optional).

RtcBridge offers a scalable and reliable solution for WebRTC signaling, enabling efficient and real-time communication for WebRTC clients.