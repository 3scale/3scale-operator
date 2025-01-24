package component

type SystemDatabaseType string

const (
	SystemDatabaseTypeExternal SystemDatabaseType = "external"
)

type DeploymentsLister struct {
	ExternalZyncDatabase bool
	IsZyncEnabled        bool
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
		SystemSearchdDeploymentName,
	)

	if d.IsZyncEnabled {
		deployments = append(deployments, ZyncName)
		deployments = append(deployments, ZyncQueDeploymentName)
	}

	if !d.ExternalZyncDatabase && d.IsZyncEnabled {
		deployments = append(deployments, ZyncDatabaseDeploymentName)
	}

	return deployments
}
