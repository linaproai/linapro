// This file defines cluster configuration value objects owned by the cluster
// topology component. The config service aliases these types when loading
// config.yaml so topology consumers can depend on cluster.Service without
// introducing a package cycle.

package cluster

import "time"

// ClusterConfig holds cluster topology configuration.
type ClusterConfig struct {
	Enabled      bool               `json:"enabled"`      // Enabled reports whether clustered deployment is enabled.
	Coordination string             `json:"coordination"` // Coordination names the cluster coordination backend.
	Election     ElectionConfig     `json:"election"`     // Election contains primary-election settings for clustered mode.
	Redis        ClusterRedisConfig `json:"redis"`        // Redis stores Redis coordination connection settings.
}

// ElectionConfig holds leader election configuration.
type ElectionConfig struct {
	Lease         time.Duration `json:"lease"`         // Lease is the lock lease duration.
	RenewInterval time.Duration `json:"renewInterval"` // RenewInterval is the lease renewal interval.
}

// ClusterRedisConfig holds Redis coordination settings for clustered mode.
type ClusterRedisConfig struct {
	Address        string        `json:"address"`        // Address is the host:port endpoint for Redis.
	DB             int           `json:"db"`             // DB selects the Redis logical database.
	Password       string        `json:"password"`       // Password authenticates to Redis when configured.
	ConnectTimeout time.Duration `json:"connectTimeout"` // ConnectTimeout bounds Redis connection establishment.
	ReadTimeout    time.Duration `json:"readTimeout"`    // ReadTimeout bounds Redis read operations.
	WriteTimeout   time.Duration `json:"writeTimeout"`   // WriteTimeout bounds Redis write operations.
}
