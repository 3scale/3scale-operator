package component

type SystemDatabaseType string

const (
	SystemDatabaseTypeInternalMySQL      SystemDatabaseType = "mysql"
	SystemDatabaseTypeInternalPostgreSQL SystemDatabaseType = "postgresql"
	SystemDatabaseTypeExternal           SystemDatabaseType = "external"
)

type DeploymentsLister struct {
	SystemDatabaseType     SystemDatabaseType
	ExternalRedisDatabases bool
	ExternalZyncDatabase   bool
}

func (d *DeploymentsLister) DeploymentNames() []string {
	var deployments []string
	deployments = append(deployments,
		ApicastStagingName,
		ApicastProductionName,
		BackendListenerName,
		BackendWorkerName,
		BackendCronName,
		SystemMemcachedDeploymentName,
		SystemAppDeploymentName,
		SystemSidekiqName,
		SystemSphinxDeploymentName,
		ZyncName,
		ZyncQueDeploymentName,
	)

	switch d.SystemDatabaseType {
	case SystemDatabaseTypeInternalMySQL:
		deployments = append(deployments, SystemMySQLDeploymentName)
	case SystemDatabaseTypeInternalPostgreSQL:
		deployments = append(deployments, SystemPostgreSQLDeploymentName)
	}

	if !d.ExternalRedisDatabases {
		deployments = append(deployments,
			BackendRedisDeploymentName,
			SystemRedisDeploymentName,
		)
	}

	if !d.ExternalZyncDatabase {
		deployments = append(deployments, ZyncDatabaseDeploymentName)
	}

	return deployments
}
