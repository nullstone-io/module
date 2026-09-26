package platformdata

import "sort"

// Platform identifies the runtime an application module targets.
// The provider's env data sources use it to decide which templates are legal
// (e.g. `k8s.field(...)` only on Kubernetes); the UI uses it to render refs.
type Platform struct {
	Name string
	// SupportsK8sRefs enables `k8s.field/configMap/resourceField/fileKey(...)` templates.
	SupportsK8sRefs bool
	// SupportsSecretRefs enables `secret(...)` templates that point at an existing cloud secret.
	SupportsSecretRefs bool
}

const (
	PlatformEcs               = "ecs"
	PlatformBatch             = "batch"
	PlatformLambda            = "lambda"
	PlatformBeanstalk         = "beanstalk"
	PlatformS3                = "s3"
	PlatformK8s               = "k8s"
	PlatformCloudRun          = "cloudrun"
	PlatformCloudFunctions    = "cloudfunctions"
	PlatformComposer          = "composer"
	PlatformGce               = "gce"
	PlatformGcs               = "gcs"
	PlatformAzureContainerApp = "azure_container_app"
	PlatformAzureFunction     = "azure_function"
	PlatformAzureAppService   = "azure_app_service"
)

var platforms = map[string]Platform{
	PlatformEcs:               {Name: PlatformEcs, SupportsSecretRefs: true},
	PlatformBatch:             {Name: PlatformBatch, SupportsSecretRefs: true},
	PlatformLambda:            {Name: PlatformLambda, SupportsSecretRefs: true},
	PlatformBeanstalk:         {Name: PlatformBeanstalk, SupportsSecretRefs: true},
	PlatformS3:                {Name: PlatformS3},
	PlatformK8s:               {Name: PlatformK8s, SupportsSecretRefs: true, SupportsK8sRefs: true},
	PlatformCloudRun:          {Name: PlatformCloudRun, SupportsSecretRefs: true},
	PlatformCloudFunctions:    {Name: PlatformCloudFunctions, SupportsSecretRefs: true},
	PlatformComposer:          {Name: PlatformComposer, SupportsSecretRefs: true},
	PlatformGce:               {Name: PlatformGce, SupportsSecretRefs: true},
	PlatformGcs:               {Name: PlatformGcs},
	PlatformAzureContainerApp: {Name: PlatformAzureContainerApp},
	PlatformAzureFunction:     {Name: PlatformAzureFunction},
	PlatformAzureAppService:   {Name: PlatformAzureAppService},
}

// LookupPlatform returns the platform definition; ok is false for an unknown name.
func LookupPlatform(name string) (Platform, bool) {
	p, ok := platforms[name]
	return p, ok
}

// Platforms returns every known platform name, sorted.
func Platforms() []string {
	names := make([]string, 0, len(platforms))
	for name := range platforms {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
