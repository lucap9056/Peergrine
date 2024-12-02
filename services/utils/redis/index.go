package redis

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const _MAX_RETRY_ATTEMPTS = 10
const _RETRY_INTERVAL_TIME = time.Second * 5

type Manager struct {
	client        *redis.Client
	clusterClient *redis.ClusterClient
}

func Test(client *redis.Client) (*Manager, error) {
	return clientInit(client)
}

func New(addr string) (*Manager, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return clientInit(client)
}

func clientInit(client *redis.Client) (*Manager, error) {

	m := &Manager{
		client: client,
	}

	for i := 0; i < _MAX_RETRY_ATTEMPTS; i++ {
		if err := ping(m.client); err == nil {
			break
		} else if i == _MAX_RETRY_ATTEMPTS-1 {
			return nil, errors.New("failed to connect to Redis after maximum retries")
		}
		time.Sleep(_RETRY_INTERVAL_TIME)
	}

	isCluster := m.isClusterNode()

	if isCluster {
		clusterAddrs, err := m.getClusterNodes()
		if err != nil {
			return nil, err
		}

		m.client.Close()
		m.client = nil

		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: clusterAddrs,
		})

		m.clusterClient = clusterClient
	}

	log.Printf("Redis cluster: %v", isCluster)

	return m, nil
}

func ping(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	return err
}

func (r *Manager) isClusterNode() bool {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	info, err := r.client.ClusterInfo(ctx).Result()
	if err != nil {
		log.Printf("Error fetching cluster info: %v", err)
		return false
	}
	log.Printf("Redis info: \n%s\n", info)
	stateOk := strings.Contains(info, "cluster_state:ok")
	clusterEnabled := strings.Contains(info, "cluster_enabled:1")
	return stateOk || clusterEnabled
}

func (r *Manager) getClusterNodes() ([]string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	result, err := r.client.ClusterNodes(ctx).Result()
	if err != nil {
		return nil, err
	}

	var clusterAddrs []string
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}

		// 格式: <node_id> <ip:port> ...
		addr := strings.Split(parts[1], "@")[0]
		clusterAddrs = append(clusterAddrs, addr)
	}

	return clusterAddrs, nil
}

func (r *Manager) Get(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	if r.clusterClient != nil {
		return r.clusterClient.Get(ctx, key).Bytes()
	}

	return r.client.Get(ctx, key).Bytes()
}

func (r *Manager) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	if r.clusterClient != nil {
		return r.clusterClient.Scan(ctx, cursor, match, count).Result()
	}

	return r.client.Scan(ctx, cursor, match, count).Result()
}

func (r *Manager) MGet(keys []string) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	if r.clusterClient != nil {
		return r.clusterClient.MGet(ctx, keys...).Result()
	}

	return r.client.MGet(ctx, keys...).Result()
}

func (r *Manager) Set(key string, data []byte, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	if r.clusterClient != nil {
		return r.clusterClient.Set(ctx, key, data, expiration).Err()
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *Manager) Exists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	if r.clusterClient != nil {
		exists, err := r.clusterClient.Exists(ctx, key).Result()
		if err != nil {
			return false, err
		}
		return exists > 0, nil
	}

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *Manager) Del(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	if r.clusterClient != nil {
		return r.clusterClient.Del(ctx, key).Err()
	}

	return r.client.Del(ctx, key).Err()
}

func (r *Manager) Close() error {
	if r.clusterClient != nil {
		return r.clusterClient.Close()
	}

	return r.client.Close()
}
