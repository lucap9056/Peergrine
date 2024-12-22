# Peergrine MsgBridge Service

The **MsgBridge** service is responsible for forwarding messages between clients. It relies on **Pulsar** for horizontal scaling and uses **Redis** for caching and authentication when required. Configuration can be managed through environment variables, command-line arguments, or Zookeeper.

---

## Configuration Overview

MsgBridge service settings can be loaded from:
1. **Zookeeper**: Reads settings from a specified configuration path in Zookeeper.
2. **Environment Variables**: Sets configuration options via environment variables.
3. **Command-Line Arguments**: Overrides configuration options using command-line arguments.

---

## Configuration Options

|**SettingField**|**Description**|**DefaultValue**|
|-|-|-|
|`APP_ID`| Unique service identifier (optional | Randomly generated |
|`APP_ADDR`|Address where the service runs (optional) | `:80` |
|`APP_AUTH_ADDR`|Address for authentication service | None |
|`APP_REDIS_ADDR` |Redis server address (optional) |None (no Redis used) |
|`APP_PULSAR_ADDRS` |List of Pulsar broker addresses (optional, comma-separated) |None |
|`APP_PULSAR_TOPIC` |Pulsar topic name for communication (optional) |None |

> **Notes:**
> - If `APP_AUTH_ADDR` is empty but `APP_REDIS_ADDR` is set, the service will attempt to handle authentication independently.

---

## Zookeeper Configuration

Zookeeper provides centralized management of MsgBridge settings. If Zookeeper addresses and paths are provided, MsgBridge retrieves its configuration from Zookeeper.

### Zookeeper Path

- **Default Path**: `/msg-bridge`  
- Override the default path using the `CONFIG_PATH` environment variable or command-line arguments.

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
export APP_PULSAR_ADDRS="pulsar://pulsar-broker-0:6650,pulsar://pulsar-broker-1:6650"
export APP_PULSAR_TOPIC="MsgBridge"
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
- **Pulsar**: For high-performance horizontal message delivery.
- **Redis**: For caching and optional independent authentication.
- **Zookeeper**: For centralized configuration management (optional).

This setup ensures that MsgBridge is both scalable and reliable for handling real-time message forwarding in distributed systems.