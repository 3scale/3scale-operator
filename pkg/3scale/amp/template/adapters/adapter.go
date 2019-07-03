package adapters

import (
	templatev1 "github.com/openshift/api/template/v1"
)

type Adapter interface {
	Adapt(*templatev1.Template)
}
