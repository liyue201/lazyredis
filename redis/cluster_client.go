package redis

import "github.com/wuxibin89/redis-go-cluster"

type ClusterClient struct {
	cluster *redis.Cluster
} 

func NewClusterClient(opts *redis.Options) (Client, error) {
	cluster, err := redis.NewCluster(opts)
	return &ClusterClient{cluster: cluster}, err
}

func (c *ClusterClient) Do(cmd string, args ...interface{}) (string, error) {
	reply, err := c.cluster.Do(cmd, args...)
	if err != nil {
		return "", err
	}
	return replyText(reply), nil
}

func (c *ClusterClient) Close() error {
	c.cluster.Close()
	return nil
}
