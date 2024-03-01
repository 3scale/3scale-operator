package component

type SystemMySQLImage struct {
	Options *SystemMySQLImageOptions
}

func NewSystemMySQLImage(options *SystemMySQLImageOptions) *SystemMySQLImage {
	return &SystemMySQLImage{Options: options}
}
