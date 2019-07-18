package redis

import "github.com/wuxibin89/redis-go-cluster"

type ClusterClient struct {
	cluster *redis.Cluster
}

func NewClusterClient(opts *redis.Options) (Client, error) {
	cluster, err := redis.NewCluster(opts)
	return &ClusterClient{cluster: cluster}, err
}

func (c *ClusterClient) Do(cmd string, args ...interface{}) (interface{}, error) {
	return c.cluster.Do(cmd, args...)
}

func (c *ClusterClient) Close() error {
	c.cluster.Close()
	return nil
}