# Peergrine Kafker 服務設定

此服務專門負責管理 Kafka 分區的分配，並且依賴於 Kafka 和 Zookeeper。在分散式環境中，Kafker 會通過 Zookeeper 來管理設定和服務註冊，並且可以將設定存儲於 Zookeeper 中，支援通過環境變數或命令行參數進行自定義設定。預計未來將逐步淘汰。

## 設定概述

Kafker 服務支持通過環境變數或命令行參數來設定服務的各項參數。這些設定會決定服務如何連接到 Kafka、Zookeeper 以及 Consul，並且控制服務的一些行為（例如是否啟用集群模式）。
## 設定選項

以下是 Kafker 服務的主要設定項目：
|設定欄位 |描述 |預設值 |
|---|
|**APP_ID** |服務的唯一識別碼 (可選) |隨機生成 |
|**APP_ADDR** |服務的地址 (可選) |:50051 |
|**APP_CLUSTER_MODE** |集群模式開關 (可選) |false |
|**APP_KAFKA_ADDR** |Kafka 伺服器地址 (可選) |kafka:9092 |
|**APP_CONSUL_ADDR** |Consul 伺服器地址 (可選) |consul:8500 |
|**APP_SERVICE_NAME** |註冊的服務名稱 (可選) |Kafker |
|**APP_SERVICE_ADDR** |服務地址 (可選) |服務自行判定 |
|**APP_SERVICE_PORT** |服務健康檢查端口 (可選) |4000 |
|**APP_ZOOKEEPER_ADDRS** |Zookeeper 伺服器地址 (必選) |無預設值 |
|**CONFIG_PATH** |Zookeeper 中的設定路徑 (可選) |/kafker |
> 若 **APP_CONSUL_ADDR** 為空，則 `APP_SERVICE_NAME`、`APP_SERVICE_ADDR`、`APP_SERVICE_PORT` 三項設置將不會生效。
### Zookeeper 設定

該服務還允許從 Zookeeper 獲取設定，這使得它在分散式環境中能夠進行集中化的設定管理。如果提供了 Zookeeper 地址和路徑，服務將會嘗試從 Zookeeper 獲取設定。

#### Zookeeper 路徑

- 預設的設定路徑為 `/kafker`，您可以使用環境變數 `APP_CONFIG_PATH` 指定不同的路徑。
### 設定流程

1. **Zookeeper**：服務會嘗試從指定的 Zookeeper 路徑讀取設定。如果該路徑不存在，將會寫入預設設定。
2. **環境變數**：可以通過環境變數設置設定值（例如 `APP_CLIENTENDPOINT_ADDR`）。
3. **命令行參數**：也可以通過命令行參數傳遞設定值（例如 `-zookeeper-addrs`）。

### Zookeeper 初始化

若設定路徑在 Zookeeper 中不存在，服務會自動創建該路徑並儲存預設設定。這樣即便未手動設定，服務也能使用預設值啟動。
### 範例

以下是如何設定 Kafker 服務的範例：
#### 使用環境變數設定

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

#### 使用命令行參數設定
```
./app -zookeeper-addrs "zookeeper1:2181,zookeeper2:2181" -config-path "/kafker"
```
