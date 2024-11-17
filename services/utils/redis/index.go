package redis

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type Manager struct {
	ctx           context.Context
	client        *redis.Client
	clusterClient *redis.ClusterClient
}

func Test(client *redis.Client) *Manager {
	ctx := context.Background()
	return &Manager{ctx, client, nil}
}

func New(addr string) (*Manager, error) {

	m := &Manager{
		ctx: context.Background(),
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}

	isCluster := m.isClusterNode()

	if isCluster {
		clusterAddrs, err := m.getClusterNodes()
		if err != nil {
			return nil, err
		}

		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: clusterAddrs,
		})

		m.clusterClient = clusterClient
	}

	log.Printf("Redis cluster: %v", isCluster)

	return m, nil
}

func (r *Manager) isClusterNode() bool {

	info, err := r.client.ClusterInfo(r.ctx).Result()
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

	result, err := r.client.ClusterNodes(r.ctx).Result()
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

	if r.clusterClient != nil {
		return r.clusterClient.Get(r.ctx, key).Bytes()
	}

	return r.client.Get(r.ctx, key).Bytes()
}

func (r *Manager) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {

	if r.clusterClient != nil {
		return r.clusterClient.Scan(r.ctx, cursor, match, count).Result()
	}

	return r.client.Scan(r.ctx, cursor, match, count).Result()
}

func (r *Manager) MGet(keys []string) ([]interface{}, error) {

	if r.clusterClient != nil {
		return r.clusterClient.MGet(r.ctx, keys...).Result()
	}

	return r.client.MGet(r.ctx, keys...).Result()
}

func (r *Manager) Set(key string, data []byte, expiration time.Duration) error {

	if r.clusterClient != nil {
		return r.clusterClient.Set(r.ctx, key, data, expiration).Err()
	}

	return r.client.Set(r.ctx, key, data, expiration).Err()
}

func (r *Manager) Exists(key string) (bool, error) {

	if r.clusterClient != nil {
		exists, err := r.clusterClient.Exists(r.ctx, key).Result()
		if err != nil {
			return false, err
		}
		return exists > 0, nil
	}

	exists, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *Manager) Del(key string) error {

	if r.clusterClient != nil {
		return r.clusterClient.Del(r.ctx, key).Err()
	}

	return r.client.Del(r.ctx, key).Err()
}

func (r *Manager) Close() error {

	if r.clusterClient != nil {
		r.clusterClient.Close()
	}

	return r.client.Close()
}
