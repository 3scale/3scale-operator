package component

import templatev1 "github.com/openshift/api/template/v1"

type Component interface {
	AssembleIntoTemplate(*templatev1.Template, []Component)
	PostProcess(*templatev1.Template, []Component)
}

type ComponentType string

const (
	RedisType                 ComponentType = "redis"
	ApicastType               ComponentType = "apicast"
	WildcardRouterType        ComponentType = "wildcardrouter"
	BackendType               ComponentType = "backend"
	SystemType                ComponentType = "system"
	ZyncType                  ComponentType = "zync"
	MySQLType                 ComponentType = "mysql"
	AmpImagesType             ComponentType = "ampimages"
	MemcachedType             ComponentType = "memcached"
	S3Type                    ComponentType = "s3"
	ProductizedType           ComponentType = "productized"
	EvaluationType            ComponentType = "evaluation"
	HighAvailabilityType      ComponentType = "ha"
	AmpTemplateType           ComponentType = "amp-template"
	AmpS3TemplateType         ComponentType = "amp-s3-template"
	AmpEvalTemplateType       ComponentType = "amp-eval-template"
	AmpEvalS3TemplateType     ComponentType = "amp-eval-s3-template"
	AmpHATemplateType         ComponentType = "amp-ha-template"
	AmpPostgreSQLTemplateType ComponentType = "amp-postgresql-template"
)
