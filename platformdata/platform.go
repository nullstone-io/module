package platformdata

import "sort"

// Platform identifies the runtime an application module targets.
// The provider's env data sources use it to decide which templates are legal
// (e.g. `k8s.field(...)` only on Kubernetes); the UI uses it to render refs.
//
// Identifiers are prefixed with the cloud so they are unambiguous on their own:
// `aws_ecs`, `gcp_gke`, `azure_aks`, never bare `ecs` or `k8s`.
type Platform struct {
	Name string
	// SupportsK8sRefs enables `k8s.field/configMap/resourceField/fileKey(...)` templates.
	SupportsK8sRefs bool
	// SupportsSecretRefs enables `secret(...)` templates that point at an existing cloud secret.
	SupportsSecretRefs bool
}

const (
	PlatformAwsEcs       = "aws_ecs"
	PlatformAwsBatch     = "aws_batch"
	PlatformAwsLambda    = "aws_lambda"
	PlatformAwsBeanstalk = "aws_beanstalk"
	PlatformAwsEc2       = "aws_ec2"
	PlatformAwsS3        = "aws_s3"
	PlatformAwsEks       = "aws_eks"

	PlatformGcpGke            = "gcp_gke"
	PlatformGcpCloudRun       = "gcp_cloudrun"
	PlatformGcpCloudFunctions = "gcp_cloudfunctions"
	PlatformGcpComposer       = "gcp_composer"
	PlatformGcpGce            = "gcp_gce"
	PlatformGcpGcs            = "gcp_gcs"

	PlatformAzureAks          = "azure_aks"
	PlatformAzureContainerApp = "azure_container_app"
	PlatformAzureFunction     = "azure_function"
	PlatformAzureAppService   = "azure_app_service"
	PlatformAzureStaticWebApp = "azure_static_web_app"
)

var platforms = map[string]Platform{
	PlatformAwsEcs:       {Name: PlatformAwsEcs, SupportsSecretRefs: true},
	PlatformAwsBatch:     {Name: PlatformAwsBatch, SupportsSecretRefs: true},
	PlatformAwsLambda:    {Name: PlatformAwsLambda, SupportsSecretRefs: true},
	PlatformAwsBeanstalk: {Name: PlatformAwsBeanstalk, SupportsSecretRefs: true},
	PlatformAwsEc2:       {Name: PlatformAwsEc2, SupportsSecretRefs: true},
	PlatformAwsS3:        {Name: PlatformAwsS3},
	PlatformAwsEks:       {Name: PlatformAwsEks, SupportsSecretRefs: true, SupportsK8sRefs: true},

	PlatformGcpGke:            {Name: PlatformGcpGke, SupportsSecretRefs: true, SupportsK8sRefs: true},
	PlatformGcpCloudRun:       {Name: PlatformGcpCloudRun, SupportsSecretRefs: true},
	PlatformGcpCloudFunctions: {Name: PlatformGcpCloudFunctions, SupportsSecretRefs: true},
	PlatformGcpComposer:       {Name: PlatformGcpComposer, SupportsSecretRefs: true},
	PlatformGcpGce:            {Name: PlatformGcpGce, SupportsSecretRefs: true},
	PlatformGcpGcs:            {Name: PlatformGcpGcs},

	PlatformAzureAks:          {Name: PlatformAzureAks, SupportsSecretRefs: true, SupportsK8sRefs: true},
	PlatformAzureContainerApp: {Name: PlatformAzureContainerApp},
	PlatformAzureFunction:     {Name: PlatformAzureFunction},
	PlatformAzureAppService:   {Name: PlatformAzureAppService},
	PlatformAzureStaticWebApp: {Name: PlatformAzureStaticWebApp},
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
