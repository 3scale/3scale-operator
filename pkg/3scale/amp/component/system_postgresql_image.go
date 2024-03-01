package component

type SystemPostgreSQLImage struct {
	Options *SystemPostgreSQLImageOptions
}

func NewSystemPostgreSQLImage(options *SystemPostgreSQLImageOptions) *SystemPostgreSQLImage {
	return &SystemPostgreSQLImage{Options: options}
}
