package conf

import (
	"fmt"
)

const (
	// DefaultCheckInterval is the default CheckInterval if one is not provided in a Conf Object
	DefaultCheckInterval = 60
)

// Conf is the basic configuration structure to be used by this program. It defines kubernetes config locations as well
// as pod kicking criteria.
type Conf struct {
	// KubeConf defines a path to a configuration file to use when connecting to the Kubernetes Cluster we are managing
	// pod kicking for. If not provided it will attempt to use one setup within the home directory of the invokeing user.
	KubeConf string `yaml:"kubeConf"`

	// CheckInterval defines the interval in seconds between kicker evaluations
	CheckInterval int64 `yaml:"checkInterval"`

	// Criteria is the set of targetting strategies for this programm to use. At least one valid criteria must be provided.
	Criteria []Criteria `yaml:"criteria"`
}

func (c *Conf) validate() error {
	if c.CheckInterval <= 0 {
		c.CheckInterval = DefaultCheckInterval
	}

	if len(c.Criteria) <= 0 {
		return fmt.Errorf("Must provide at least one Criteria in conf")
	}

	for i := range c.Criteria {
		if err := c.Criteria[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

const (
	// DefaultNamespace is the namespace used if one is not provided in a Criteria Object
	DefaultNamespace = "default"

	// DefaultMaxAge is default MaxAge in seconds if one is not provided in a Criteria Object
	DefaultMaxAge = 86400

	// DefaultMinAge is the default MinAge in seconds if one is not provided in a Criteria Object
	DefaultMinAge = 90

	// DefaultStrategy is the default Strategy if one is not provided in a Criteria Object
	DefaultStrategy = StrategySpread

	// DefaultLimit is the default Limit is one is not provided in a Criteria Object
	DefaultLimit = 1

	// DefaultGracePeriod is the default GracePeriod if one is not provided in a Criteria Object
	DefaultGracePeriod = 30

	// DefaultCoolDown is the default CoolDown if one is not provided in a Criteria Object
	DefaultCoolDown = 300
)

// Criteria defines a set of criteria used for targetting pods to kick.
type Criteria struct {
	// Name is the name of the pods used to target this Criteria.
	// This is a required field
	Name string `yaml:"name"`

	// Namespace is the namespace to use when finding and working with these pods.
	// Defaults to DefaultNamespace if left empty.
	Namespace string `yaml:"namespace"`

	// MaxAge is the maximum age in seconds that a pod should live for to be eligible for kicking.
	// Must be greater then MinAge. Defaults to DefaultMaxAge if not provided or <= 0.
	MaxAge int64 `yaml:"maxAge"`

	// MinAge is the minimum age in seconds that a pod should be alive for to before being considered eligible for
	// kicking.
	// Must be less then MaxAge. Defaults to DefaultMinAge if not provided or <= 0.
	MinAge int64 `yaml:"minAge"`

	// Strategy is the strategy used to manage which pod is kicked if needed. Defaults to DefaultStrategy
	Strategy Strategy `yaml:"strategy"`

	// Limit is the maximum count of pods that can be kicked per evaluation period. Defaults to DefaultLimit
	Limit int64 `yaml:"limit"`

	// GracePeriod is the grace period in seconds that a kicked pod will have when shutting down
	GracePeriod int64 `yaml:"gracePeriod"`

	// DefaultCoolDown is the cool down in seconds that a strategy will wait before being elidgable to kick a pod
	CoolDown int64 `yaml:"coolDown"`
}

func (c *Criteria) validate() error {
	if c.Name == "" {
		return fmt.Errorf("Criteria must have a Name")
	}

	if c.Namespace == "" {
		c.Namespace = DefaultNamespace
	}

	if c.MinAge <= 0 {
		c.MinAge = DefaultMinAge
	}

	if c.MaxAge <= 0 {
		c.MaxAge = DefaultMaxAge
	}

	if c.MaxAge <= c.MinAge {
		return fmt.Errorf("MaxAge: %ds must be less then MinAge: %ds", c.MaxAge, c.MinAge)
	}

	if c.Strategy == "" {
		c.Strategy = DefaultStrategy
	}

	if c.Limit <= 0 {
		c.Limit = DefaultLimit
	}

	if c.GracePeriod <= 0 {
		c.GracePeriod = DefaultGracePeriod
	}

	if c.CoolDown <= 0 {
		c.CoolDown = DefaultCoolDown
	}

	return nil
}

// Strategy defines the strategy to use for kicking pods
type Strategy string

const (
	// StrategySpread attempts to spread out the kicking of pods to a evenly distributed schedule within the provided
	// MaxAge. Keep in mind that pods with ages closer then count/MaxAge may exist longer then the MaxAge.
	StrategySpread = "spread"
	// StrategyImmediate kicks any pod that is over MaxAge regardless of the state of other pods. This is a fairly
	// drastic approach and should be used with caution.
	StrategyImmediate = "immediate"
)
