# Peergrine RtcBridge 服務
此服務為WebRTC的Signaling Server。該服務目前依賴於 `Kafker`，`Kafka` 進行水平擴展，並且配置與設定由環境變數、命令行參數或 Zookeeper 來管理。RtcBridge 能夠與 Kafka、Redis、Kafker 及其他服務進行整合，提供一個高效能的訊息橋接解決方案。

## 服務設定

RtcBridge 服務的設定檔可以由以下來源載入：
1. **Zookeeper**：如果指定 Zookeeper 伺服器和配置路徑，服務會從 Zookeeper 讀取配置。
2. **環境變數**：可以透過環境變數來設置配置項。
3. **命令行參數**：使用命令行參數來覆蓋配置項。

## 設定概述

RtcBridge 的主要設定項目包括：
- **地址設定**：服務的運行地址（`APP_ADDR`）。
- **授權地址**：服務用於認證的地址（`APP_AUTH_ADDR`）。
- **Redis 地址**：服務使用的 Redis 伺服器地址（`APP_REDIS_ADDR`）。
- **Kafka 地址**：服務連接的 Kafka 伺服器地址（`APP_KAFKA_ADDR`）。
- **Kafker 地址**：Kafker 服務的地址（`APP_KAFKER_ADDR`）。
- **Consul 配置**：服務註冊的 Consul 地址及相關配置（`APP_CONSUL_ADDR`）。

每個設定項都可以通過環境變數、命令行參數，或從 Zookeeper 配置中加載。
## 設定選項
|設定項 |描述 |預設值 |
|-|-|-|
|**APP_ID** |服務的唯一識別碼 (可選) |隨機生成 |
|**APP_ADDR** |服務的運行地址 |:80 |
|**APP_AUTH_ADDR** |授權服務的地址 |空 |
|**APP_REDIS_ADDR** |Redis 服務的地址 |空 |
|**APP_KAFKA_ADDR** |Kafka 服務的地址 |空 |
|**APP_KAFKER_ADDR** |Kafker 服務的地址 |空 |
|**APP_CONSUL_ADDR** |Consul 服務的地址 |空 |
|**APP_SERVICE_NAME** |服務名稱 |RtcBridge |
|**APP_SERVICE_ADDR** |服務地址 (可選) |服務自行判定 |
|**APP_SERVICE_PORT** |服務健康檢查端口 |4000 |
> 若 **APP_CONSUL_ADDR** 為空，則 `APP_SERVICE_NAME`、`APP_SERVICE_ADDR`、`APP_SERVICE_PORT` 三項設置將不會生效。
> 若 **APP_AUTH_ADDR** 為空 **APP_REDIS_ADDR** 不為空，服務將會嘗試自行處理身份驗證。
## Zookeeper 設定
Zookeeper 配置允許 RtcBridge 服務從指定的 Zookeeper 路徑讀取設定。這對於分散式環境中管理配置非常有用。
### Zookeeper 路徑

RtcBridge 服務會從以下 Zookeeper 路徑讀取配置：
- **預設路徑**：`/rtc-bridge`
### 配置流程

1. **Zookeeper**：如果提供了 Zookeeper 地址 (`APP_ZOOKEEPER_ADDRS`)，服務會嘗試從指定的 Zookeeper 路徑讀取設定。如果該路徑不存在，將會寫入預設設定。
2. **環境變數**：可以通過環境變數設置設定值（例如 `APP_ID`）。
3. **命令行參數**：也可以通過命令行參數傳遞設定值（例如 `-zookeeper-addrs`）。
### Zookeeper 初始化

若設定路徑在 Zookeeper 中不存在，服務會自動創建該路徑並儲存預設設定。這樣即便未手動設定，服務也能使用預設值啟動。

### 範例

以下是如何設定服務的範例：
#### 使用環境變數

```
export APP_ADDR=":8080"
export APP_KAFKA_ADDR="kafka:9092"
export APP_ZOOKEEPER_ADDRS="zookeeper:2181"
```

#### 使用命令行參數
```
./app -ZOOKEEPER_ADDRS=zookeeper:2181 -CONFIG_PATH="/rtc-bridge"
```