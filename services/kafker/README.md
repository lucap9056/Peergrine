# Peergrine Kafker Service (Deprecated)


The **Kafker** service is dedicated to managing Kafka partition assignments and depends on both Kafka and Zookeeper. It uses Zookeeper for configuration management and service registration.

----

## Configuration Overview

Kafker supports configuration through **environment variables** and **command-line parameters**. These settings control how the service connects to Kafka, Zookeeper, and Consul, as well as how it operates (e.g., enabling cluster mode).

----

## Configuration Options

|**Setting Field** |**Description** |**Default Value** |
|-|-|-|
|`APP_ID` |Unique service identifier (optional) |Randomly generated |
|`APP_ADDR` |Address for the service (optional) |`:50051` |
|`APP_CLUSTER_MODE` |Toggle for cluster mode (optional) |`false` |
|`APP_KAFKA_ADDR` |Kafka server address (optional) |`kafka:9092` |
|`APP_CONSUL_ADDR` |Consul server address (optional) |`consul:8500` |
|`APP_SERVICE_NAME` |Registered service name (optional) |`Kafker` |
|`APP_SERVICE_ADDR` |Service address (optional) |Determined by service |
|`APP_SERVICE_PORT` |Health check port for the service (optional) |`4000` |
|`APP_ZOOKEEPER_ADDRS` |List of Zookeeper server addresses (required) |None |
|`CONFIG_PATH` |Configuration path in Zookeeper (optional) |`/kafker` |

>**Note:** If `APP_CONSUL_ADDR` is not set, the `APP_SERVICE_NAME`, `APP_SERVICE_ADDR`, and `APP_SERVICE_PORT` settings will not be used.

----

## Zookeeper Configuration

The service supports centralized configuration management using Zookeeper. If Zookeeper addresses and paths are provided, Kafker retrieves its configuration from Zookeeper.

### Zookeeper Path

- Default configuration path in Zookeeper: `/kafker`.

- Override the default path using the `APP_CONFIG_PATH` environment variable.

### Configuration Workflow

1. **Zookeeper**: The service attempts to read settings from the specified Zookeeper path. If the path does not exist, default settings are written to it.

2. **Environment Variables**: Use environment variables to set configuration values such as `APP_ADDR` or `APP_KAFKA_ADDR`.

3. **Command-Line Parameters**: Pass configuration values as command-line arguments, such as `-zookeeper-addrs`.

----

## Zookeeper Initialization

If the specified configuration path is missing in Zookeeper, Kafker automatically creates it and populates it with default settings. This ensures seamless startup even without prior configuration.

----

## Examples

### Setting Configuration via Environment Variables

```
export APP_ID="kafker-service-1234"
export APP_ADDR=":9090"
export APP_CLUSTER_MODE="true"
export APP_KAFKA_ADDR="kafka:9092"
export APP_CONSUL_ADDR="consul:8500"
export APP_SERVICE_NAME="Kafker"
export APP_SERVICE_PORT="4000"
export APP_ZOOKEEPER_ADDRS="zookeeper1:2181,zookeeper2:2181"
export CONFIG_PATH="/kafker"
```
### Setting Configuration via Command-Line Parameters

```
./app -zookeeper-addrs "zookeeper1:2181,zookeeper2:2181" -config-path "/kafker"
```

----

This flexibility in configuration makes Kafker adaptable for various deployment scenarios, allowing it to integrate smoothly within distributed environments while ensuring reliable Kafka partition management.