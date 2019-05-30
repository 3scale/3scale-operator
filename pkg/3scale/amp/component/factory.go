package component

func NewComponent(componentName string, componentOptions []string) Component {
	var component Component

	switch ComponentType(componentName) {
	case RedisType:
		// The *Redis is automatically converted to the Component interface when returning.
		// Another option would be:
		// var result Component = NewRedis(componentOptions)
		component = NewRedis(componentOptions)
		//Another option would be just return the *Redis object directly, and would be
		//automatically converted to the Component interface
		//return NewRedis(componentOptions)
	case ApicastType:
		component = NewApicast(componentOptions)
	case WildcardRouterType:
		component = NewWildcardRouter(componentOptions)
	case BackendType:
		component = NewBackend(componentOptions)
	case SystemType:
		component = NewSystem(componentOptions)
	case ZyncType:
		component = NewZync(componentOptions)
	case MySQLType:
		component = NewMysql(componentOptions)
	case AmpImagesType:
		component = NewAmpImages(componentOptions)
	case MemcachedType:
		component = NewMemcached(componentOptions)
	case S3Type:
		component = NewS3(componentOptions)
	case ProductizedType:
		component = NewProductized(componentOptions)
	case EvaluationType:
		component = NewEvaluation(componentOptions)
	case AmpTemplateType:
		component = NewAmpTemplate(componentOptions)
	case AmpS3TemplateType:
		component = NewAmpS3Template(componentOptions)
	case AmpEvalTemplateType:
		component = NewAmpEvalTemplate(componentOptions)
	case AmpEvalS3TemplateType:
		component = NewAmpEvalS3Template(componentOptions)
	case HighAvailabilityType:
		component = NewHighAvailability(componentOptions)
	case AmpHATemplateType:
		component = NewAmpHATemplate(componentOptions)
	case AmpPostgreSQLTemplateType:
		component = NewAmpPostgreSQLTemplate(componentOptions)

	default:
		panic("Error: Component not recognized")
	}

	return component
}
