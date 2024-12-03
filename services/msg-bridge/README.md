# Peergrine MsgBridge Service

The **MsgBridge** service is responsible for forwarding messages between clients. It relies on **Kafker** and **Kafka** for horizontal scaling and uses **Redis** for caching and authentication when required. Configuration can be managed through environment variables, command-line arguments, or Zookeeper.

This high-performance messaging bridge integrates seamlessly with Kafka, Redis, and other services to provide efficient message forwarding.

---

## Configuration Overview

MsgBridge service settings can be loaded from:
1. **Zookeeper**: Reads settings from a specified configuration path in Zookeeper.
2. **Environment Variables**: Sets configuration options via environment variables.
3. **Command-Line Arguments**: Overrides configuration options using command-line arguments.

---

## Configuration Options

| **Setting Field**         | **Description**                                     | **Default Value** |
|-|-|-|
| `APP_ID`                  | Unique identifier for the service (optional)        | Randomly generated |
| `APP_ADDR`                | Address where the service runs                      | `:80`             |
| `APP_AUTH_ADDR`           | Address for authentication service                  | Empty             |
| `APP_REDIS_ADDR`          | Address of the Redis server                         | Empty             |
| `APP_KAFKA_ADDR`          | Address of the Kafka server                         | Empty             |
| `APP_KAFKER_ADDR`         | Address of the Kafker service                       | Empty             |
| `APP_CONSUL_ADDR`         | Address of the Consul service                       | Empty             |
| `APP_SERVICE_NAME`        | Name of the service                                 | `MsgBridge`       |
| `APP_SERVICE_ADDR`        | Service address (optional)                          | Determined by service |
| `APP_SERVICE_PORT`        | Port for health checks                              | `4000`            |

> **Notes:**
> - If `APP_CONSUL_ADDR` is not set, `APP_SERVICE_NAME`, `APP_SERVICE_ADDR`, and `APP_SERVICE_PORT` will not be used.
> - If `APP_AUTH_ADDR` is empty but `APP_REDIS_ADDR` is set, the service will attempt to handle authentication independently.

---

## Zookeeper Configuration

Zookeeper provides centralized management of MsgBridge settings. If Zookeeper addresses and paths are provided, MsgBridge retrieves its configuration from Zookeeper.

### Zookeeper Path

- **Default Path**: `/msg-bridge`  
- Override the default path using the `CONFIG_PATH` environment variable or command-line arguments.

---

### Configuration Workflow

1. **Zookeeper**: The service attempts to load settings from the specified Zookeeper path. If the path does not exist, it initializes with default settings and saves them to Zookeeper.
2. **Environment Variables**: Settings like `APP_ADDR` or `APP_KAFKA_ADDR` can be specified via environment variables.
3. **Command-Line Arguments**: Settings can also be passed via command-line arguments, such as `-zookeeper-addrs`.

---

### Zookeeper Initialization

If the specified configuration path in Zookeeper does not exist, MsgBridge will create it and populate it with default settings, ensuring it can operate without manual configuration.

---

## Examples

### Setting Configuration via Environment Variables

```bash
export APP_ADDR=":8080"
export APP_AUTH_ADDR="auth-service:5000"
export APP_REDIS_ADDR="redis:6379"
export APP_KAFKA_ADDR="kafka:9092"
export APP_KAFKER_ADDR="kafker:50051"
export APP_ZOOKEEPER_ADDRS="zookeeper1:2181,zookeeper2:2181"
export CONFIG_PATH="/msg-bridge"
```

### Setting Configuration via Command-Line Arguments

```bash
./app -zookeeper-addrs "zookeeper1:2181,zookeeper2:2181" -config-path "/msg-bridge"
```

---

## Integration

MsgBridge integrates with:
- **Kafka**: For high-performance message queuing and delivery.
- **Redis**: For caching and optional independent authentication.
- **Kafker**: To manage Kafka partition assignments.
- **Zookeeper**: For centralized configuration management.
- **Consul**: For service registration and discovery (optional).

This setup ensures that MsgBridge is both scalable and reliable for handling real-time message forwarding in distributed systems.