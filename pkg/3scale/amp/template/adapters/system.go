package adapters

import (
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/common"
	templatev1 "github.com/openshift/api/template/v1"
)

type System struct {
	FileStorageType component.SystemFileStorageType
}

func NewSystemAdapter(fileStorageType component.SystemFileStorageType) Adapter {
	return NewAppenderAdapter(&System{FileStorageType: fileStorageType})
}

func (s *System) Parameters() []templatev1.Parameter {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:        "WILDCARD_DOMAIN",
			Description: "Root domain for the wildcard routes. Eg. example.com will generate 3scale-admin.example.com.",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_BACKEND_USERNAME",
			Description: "Internal 3scale API username for internal 3scale api auth.",
			Value:       "3scale_api_user",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_BACKEND_PASSWORD",
			Description: "Internal 3scale API password for internal 3scale api auth.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_BACKEND_SHARED_SECRET",
			Description: "Shared secret to import events from backend to system.",
			Generate:    "expression",
			From:        "[a-z0-9]{8}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_APP_SECRET_KEY_BASE",
			Description: "System application secret key base",
			Generate:    "expression",
			From:        "[a-f0-9]{128}",
			Required:    true,
		},
		templatev1.Parameter{
			Name:     "ADMIN_PASSWORD",
			Generate: "expression",
			From:     "[a-z0-9]{8}",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "ADMIN_USERNAME",
			Value:    "admin",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "ADMIN_EMAIL",
			Required: false,
		},
		templatev1.Parameter{
			Name:        "ADMIN_ACCESS_TOKEN",
			Description: "Admin Access Token with all scopes and write permissions for API access.",
			Generate:    "expression",
			From:        "[a-z0-9]{16}",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "MASTER_NAME",
			Description: "The root name which Master Admin UI will be available at.",
			Value:       "master",
			Required:    true,
		},
		templatev1.Parameter{
			Name:     "MASTER_USER",
			Value:    "master",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "MASTER_PASSWORD",
			Generate: "expression",
			From:     "[a-z0-9]{8}",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "MASTER_ACCESS_TOKEN",
			Generate: "expression",
			From:     "[a-z0-9]{8}",
			Required: true,
		},
		templatev1.Parameter{
			Name:        "RECAPTCHA_PUBLIC_KEY",
			Description: "reCAPTCHA site key (used in spam protection)",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "RECAPTCHA_PRIVATE_KEY",
			Description: "reCAPTCHA secret key (used in spam protection)",
			Required:    false,
		},
		templatev1.Parameter{
			Name:        "SYSTEM_REDIS_URL",
			Description: "Define the external system-redis to connect to",
			Required:    true,
			Value:       "redis://system-redis:6379/1",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_MESSAGE_BUS_REDIS_URL",
			Description: "Define the external system-redis message bus to connect to. By default the same value as SYSTEM_REDIS_URL but with the logical database incremented by 1 and the result applied mod 16",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_REDIS_NAMESPACE",
			Description: "Define the namespace to be used by System's Redis Database. The empty value means not namespaced",
		},
		templatev1.Parameter{
			Name:        "SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE",
			Description: "Define the namespace to be used by System's Message Bus Redis Database. The empty value means not namespaced",
		},
	}

	parameters = append(parameters, s.fileStorageParameters()...)
	return parameters
}

func (s *System) Objects() ([]common.KubernetesObject, error) {
	systemOptions, err := s.options()
	if err != nil {
		return nil, err
	}
	systemComponent := component.NewSystem(systemOptions)
	return systemComponent.Objects(), nil
}

func (s *System) options() (*component.SystemOptions, error) {
	sob := component.SystemOptionsBuilder{}
	sob.AdminAccessToken("${ADMIN_ACCESS_TOKEN}")
	sob.AdminPassword("${ADMIN_PASSWORD}")
	sob.AdminUsername("${ADMIN_USERNAME}")
	adminEmail := "${ADMIN_EMAIL}"
	sob.AdminEmail(&adminEmail)
	sob.AmpRelease("${AMP_RELEASE}")
	sob.ApicastAccessToken("${APICAST_ACCESS_TOKEN}")
	sob.ApicastRegistryURL("${APICAST_REGISTRY_URL}")
	sob.MasterAccessToken("${MASTER_ACCESS_TOKEN}")
	sob.MasterName("${MASTER_NAME}")
	sob.MasterUsername("${MASTER_USER}")
	sob.MasterPassword("${MASTER_PASSWORD}")
	sob.AppLabel("${APP_LABEL}")
	sob.RecaptchaPublicKey("${RECAPTCHA_PUBLIC_KEY}")
	sob.RecaptchaPrivateKey("${RECAPTCHA_PRIVATE_KEY}")
	redisUrl := "${SYSTEM_REDIS_URL}"
	sob.RedisURL(&redisUrl)
	redisNamespace := "${SYSTEM_REDIS_NAMESPACE}"
	sob.RedisNamespace(&redisNamespace)
	messageBusRedisURL := "${SYSTEM_MESSAGE_BUS_REDIS_URL}"
	sob.MessageBusRedisURL(&messageBusRedisURL)
	messageBusRedisNamespace := "${SYSTEM_MESSAGE_BUS_REDIS_NAMESPACE}"
	sob.MessageBusRedisNamespace(&messageBusRedisNamespace)
	sob.AppSecretKeyBase("${SYSTEM_APP_SECRET_KEY_BASE}")
	sob.BackendSharedSecret("${SYSTEM_BACKEND_SHARED_SECRET}")
	sob.TenantName("${TENANT_NAME}")
	sob.WildcardDomain("${WILDCARD_DOMAIN}")
	s.setFileStorageOptions(&sob)
	return sob.Build()
}

func (s *System) setFileStorageOptions(sob *component.SystemOptionsBuilder) {
	switch s.FileStorageType {
	case component.SystemFileStorageTypePVC:
		// We set the storageClassName as nil even though the value comes
		// as a template parameter in this case.
		// The reason for this is that the PersistentVolumeClaim Kubernetes object
		// requires a pointer to string to specify the StorageClassName. For this
		// you need to reference it using OpenShift double-brace parameter
		// expansion, which does not exist in Kubernetes. We take care of replacing
		// the value with the double brace parameter expansion as a postProcess
		// step after generating the template objects
		sob.PVCFileStorageOptions(component.PVCFileStorageOptions{StorageClass: nil})
	case component.SystemFileStorageTypeS3:
		sob.S3FileStorageOptions(component.S3FileStorageOptions{
			S3ConfigSecret: "${S3_CONFIGURATION_SECRET_NAME}",
		})
	default:
		panic("s3 or pvc has to be set as system's file storage option")
	}
}

func (s *System) fileStorageParameters() []templatev1.Parameter {
	switch s.FileStorageType {
	case component.SystemFileStorageTypePVC:
		return []templatev1.Parameter{
			templatev1.Parameter{
				Name:        "RWX_STORAGE_CLASS",
				Description: "The Storage Class to be used by ReadWriteMany System's PVC",
				Required:    false,
				Value:       "null", //TODO this is incorrect. The value that should be used at the end is null and not "null", however template parameters only accept strings.
			},
		}
	case component.SystemFileStorageTypeS3:
		return []templatev1.Parameter{
			templatev1.Parameter{
				Name:        "S3_CONFIGURATION_SECRET_NAME",
				Description: "Name of the secret that contains S3 configuration for System's File Storage. The secret name needs to exist previously.",
				Required:    true,
			},
		}
	default:
		panic("s3 or pvc has to be set as system's file storage option")
	}
}
