package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/test/builder"
	corev1 "k8s.io/api/core/v1"
)

func AssembleBuildah() pipelinev1.Task {

	task := pipelinev1.Task{
		TypeMeta:   createTaskTypeMeta(),
		ObjectMeta: createTaskObjectMeta("buildah-task"),
		Spec: pipelinev1.TaskSpec{
			Inputs:  createInputsFromBuildah(),
			Outputs: createOutputResource(),
			Steps:   createStepsFromBuildah(),
			Volumes: createVolumeSpec(),
		},
	}
	return task
}

func createInputsFromBuildah() *pipelinev1.Inputs {
	params := []pipelinev1.ParamSpec{
		pipelinev1.ParamSpec{
			Name:        "BUILDER_IMAGE",
			Type:        pipelinev1.ParamTypeString,
			Description: "The location of the buildah builder image.",
			Default:     builder.ArrayOrString("quay.io/buildah/stable:v1.11.3"),
		},
		pipelinev1.ParamSpec{
			Name:        "DOCKERFILE",
			Type:        pipelinev1.ParamTypeString,
			Description: "Path to the Dockerfile to build.",
			Default:     builder.ArrayOrString("./Dockerfile"),
		},
		pipelinev1.ParamSpec{
			Name:        "TLSVERIFY",
			Type:        pipelinev1.ParamTypeString,
			Description: "Verify the TLS on the registry endpoint (for push/pull to a non-TLS registry)",
			Default:     builder.ArrayOrString("true"),
		},
	}
	validInputResource := createTaskResource("source", "git")

	inputs := &pipelinev1.Inputs{
		Resources: []pipelinev1.TaskResource{validInputResource},
		Params:    params,
	}
	return inputs
}

func createOutputResource() *pipelinev1.Outputs {
	validOutputResource := createTaskResource("image", "image")
	outputs := &pipelinev1.Outputs{
		Resources: []pipelinev1.TaskResource{validOutputResource},
	}
	return outputs
}

func createVolumeSpec() []corev1.Volume {
	volumes := []corev1.Volume{{
		Name: "varlibcontainers",
		// EmptyDir: ,
	}}
	return volumes
}

func createStepsFromBuildah() []pipelinev1.Step {
	steps := []pipelinev1.Step{
		pipelinev1.Step{
			Container: corev1.Container{
				Name:       "replace-image",
				Image:      "mikefarah/yq",
				WorkingDir: "/workspace/source",
				Command:    []string{"yq"},
				Args: []string{
					"w",
					"-i",
					"$(inputs.params.PATHTODEPLOYMENT)/deployment.yaml",
					"$(inputs.params.YAMLPATHTOIMAGE)",
					"$(inputs.resources.image.url)",
				},
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "varlibcontainers",
					MountPath: "//tekton/home",
				}},
				SecurityContext: &corev1.SecurityContext{},
			},
		},
		pipelinev1.Step{
			Container: corev1.Container{
				Name:       "run-kubectl",
				Image:      "quay.io/kmcdermo/k8s-kubectl:latest",
				WorkingDir: "/workspace/source",
				Command:    []string{"kubectl"},
				Args: []string{
					"apply",
					"--dry-run=$(inputs.params.DRYRUN)",
					"-n",
					"$(inputs.params.NAMESPACE)",
					"-k",
					"$(inputs.params.PATHTODEPLOYMENT)",
				},
				VolumeMounts: []corev1.VolumeMount{{
					Name:      "varlibcontainers",
					MountPath: "//tekton/home",
				}},
				SecurityContext: &corev1.SecurityContext{},
			},
		},
	}
	return steps
}
