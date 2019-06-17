package component

import (
	"fmt"

	"github.com/3scale/3scale-operator/pkg/apis/common"
	"github.com/3scale/3scale-operator/pkg/helper"

	imagev1 "github.com/openshift/api/image/v1"
	templatev1 "github.com/openshift/api/template/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	insecureImportPolicy = false
)

type AmpImages struct {
	options []string
	Options *AmpImagesOptions
}

type AmpImagesOptions struct {
	appLabel                    string
	ampRelease                  string
	apicastImage                string
	backendImage                string
	systemImage                 string
	zyncImage                   string
	ZyncDatabasePostgreSQLImage string
	backendRedisImage           string
	systemRedisImage            string
	systemMemcachedImage        string
	insecureImportPolicy        bool
}

func NewAmpImages(options []string) *AmpImages {
	ampImages := &AmpImages{
		options: options,
	}
	return ampImages
}

type AmpImagesOptionsProvider interface {
	GetAmpImagesOptions() *AmpImagesOptions
}
type CLIAmpImagesOptionsProvider struct {
}

func (o *CLIAmpImagesOptionsProvider) GetAmpImagesOptions() (*AmpImagesOptions, error) {
	aob := AmpImagesOptionsBuilder{}
	aob.AppLabel("${APP_LABEL}")
	aob.AMPRelease("${AMP_RELEASE}")
	aob.ApicastImage("${AMP_APICAST_IMAGE}")
	aob.BackendImage("${AMP_BACKEND_IMAGE}")
	aob.SystemImage("${AMP_SYSTEM_IMAGE}")
	aob.ZyncImage("${AMP_ZYNC_IMAGE}")
	aob.ZyncDatabasePostgreSQLImage("${ZYNC_DATABASE_IMAGE}")
	aob.BackendRedisImage("${REDIS_IMAGE}")
	aob.SystemRedisImage("${REDIS_IMAGE}")
	aob.SystemMemcachedImage("${MEMCACHED_IMAGE}")

	aob.InsecureImportPolicy(false)

	res, err := aob.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to create AMPImages Options - %s", err)
	}
	return res, nil
}

func (ampImages *AmpImages) AssembleIntoTemplate(template *templatev1.Template, otherComponents []Component) {
	// TODO move this outside this specific method
	optionsProvider := CLIAmpImagesOptionsProvider{}
	ampImagesOpts, err := optionsProvider.GetAmpImagesOptions()
	_ = err
	ampImages.Options = ampImagesOpts
	ampImages.buildParameters(template)
	ampImages.addObjectsIntoTemplate(template)
}

func (ampImages *AmpImages) GetObjects() ([]common.KubernetesObject, error) {
	objects := ampImages.buildObjects()
	return objects, nil
}

func (ampImages *AmpImages) addObjectsIntoTemplate(template *templatev1.Template) {
	objects := ampImages.buildObjects()
	template.Objects = append(template.Objects, helper.WrapRawExtensions(objects)...)
}

func (ampImages *AmpImages) PostProcess(template *templatev1.Template, otherComponents []Component) {

}

func (ampImages *AmpImages) buildObjects() []common.KubernetesObject {
	backendImageStream := ampImages.buildAmpBackendImageStream()
	zyncImageStream := ampImages.buildAmpZyncImageStream()
	apicastImageStream := ampImages.buildApicastImageStream()
	systemImageStream := ampImages.buildAmpSystemImageStream()
	zyncDatabasePostgreSQL := ampImages.buildZyncDatabasePostgreSQLImageStream()
	backendRedisImageStream := ampImages.buildBackendRedisImageStream()
	systemRedisImageStream := ampImages.buildSystemRedisImageStream()
	systemMemcachedImageStream := ampImages.buildSystemMemcachedImageStream()

	deploymentsServiceAccount := ampImages.buildDeploymentsServiceAccount()

	objects := []common.KubernetesObject{
		backendImageStream,
		zyncImageStream,
		apicastImageStream,
		systemImageStream,
		zyncDatabasePostgreSQL,
		backendRedisImageStream,
		systemRedisImageStream,
		systemMemcachedImageStream,
		deploymentsServiceAccount,
	}
	return objects
}

func (ampImages *AmpImages) buildAmpBackendImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-backend",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "backend",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP backend",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "amp-backend (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "amp-backend " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.backendImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						// TODO this was originally a double brace expansion from a variable, that is not possible
						// natively with kubernetes so we replaced it with a const
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildAmpZyncImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-zync",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "zync",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP Zync",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP Zync (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP Zync " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.zyncImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						// TODO this was originally a double brace expansion from a variable, that is not possible
						// natively with kubernetes so we replaced it with a const
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildApicastImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-apicast",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "apicast",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP APIcast",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP APIcast (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP APIcast " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.apicastImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						// TODO this was originally a double brace expansion from a variable, that is not possible
						// natively with kubernetes so we replaced it with a const
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildAmpSystemImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp-system",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "AMP System",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP System (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "AMP system " + ampImages.Options.ampRelease,
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.systemImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildZyncDatabasePostgreSQLImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "zync-database-postgresql",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "Zync database PostgreSQL",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "Zync PostgreSQL (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "Zync " + ampImages.Options.ampRelease + " PostgreSQL",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.ZyncDatabasePostgreSQLImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildBackendRedisImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "backend-redis",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "backend",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "Backend Redis",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "Backend Redis (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "Backend " + ampImages.Options.ampRelease + " Redis",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.backendRedisImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildSystemRedisImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-redis",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "System Redis",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "System Redis (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + ampImages.Options.ampRelease + " Redis",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.backendRedisImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildSystemMemcachedImageStream() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "system-memcached",
			Labels: map[string]string{
				"app":                  ampImages.Options.appLabel,
				"threescale_component": "system",
			},
			Annotations: map[string]string{
				"openshift.io/display-name": "System Memcached",
			},
		},
		TypeMeta: metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name: "latest",
					Annotations: map[string]string{
						"openshift.io/display-name": "System Memcached (latest)",
					},
					From: &v1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: ampImages.Options.ampRelease,
					},
				},
				imagev1.TagReference{
					Name: ampImages.Options.ampRelease,
					Annotations: map[string]string{
						"openshift.io/display-name": "System " + ampImages.Options.ampRelease + " Memcached",
					},
					From: &v1.ObjectReference{
						Kind: "DockerImage",
						Name: ampImages.Options.systemMemcachedImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Insecure: insecureImportPolicy,
					},
				},
			},
		},
	}
}

func (ampImages *AmpImages) buildDeploymentsServiceAccount() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "amp",
		},
		ImagePullSecrets: []v1.LocalObjectReference{
			v1.LocalObjectReference{
				Name: "threescale-registry-auth"}}}
}

func (ampImages *AmpImages) buildParameters(template *templatev1.Template) {
	parameters := []templatev1.Parameter{
		templatev1.Parameter{
			Name:     "AMP_BACKEND_IMAGE",
			Required: true,
			Value:    "quay.io/3scale/apisonator:nightly",
		},
		templatev1.Parameter{
			Name:     "AMP_ZYNC_IMAGE",
			Value:    "quay.io/3scale/zync:nightly",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "AMP_APICAST_IMAGE",
			Value:    "quay.io/3scale/apicast:nightly",
			Required: true,
		},
		templatev1.Parameter{
			Name:     "AMP_SYSTEM_IMAGE",
			Value:    "quay.io/3scale/porta:nightly",
			Required: true,
		},
		templatev1.Parameter{
			Name:        "ZYNC_DATABASE_IMAGE",
			Description: "Zync's PostgreSQL image to use",
			Value:       "registry.access.redhat.com/rhscl/postgresql-10-rhel7",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "MEMCACHED_IMAGE",
			Description: "Memcached image to use",
			Value:       "registry.access.redhat.com/3scale-amp20/memcached",
			Required:    true,
		},
		templatev1.Parameter{
			Name:        "IMAGESTREAM_TAG_IMPORT_INSECURE",
			Description: "Set to true if the server may bypass certificate verification or connect directly over HTTP during image import.",
			Value:       "false",
			Required:    true,
		},
	}
	template.Parameters = append(template.Parameters, parameters...)
}
