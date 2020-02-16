package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/test/builder"
	corev1 "k8s.io/api/core/v1"
)

func GenerateKubectlTask() pipelinev1.Task {
	task := pipelinev1.Task{
		TypeMeta:   createTaskTypeMeta(),
		ObjectMeta: createTaskObjectMeta("deploy-using-kubectl-task"),
		Spec: pipelinev1.TaskSpec{
			Inputs: createInputsFromKubectl(),
			Steps:  createStepsFromKubectl(),
		},
	}
	return task

}

func createInputsFromKubectl() pipelinev1.Inputs {
	params := []pipelinev1.ParamSpec{
		pipelinev1.ParamSpec{
			Name:        "PATHTODEPLOYMENT",
			Type:        pipelinev1.ParamTypeString,
			Description: "Path to the manifest to apply",
			Default:     builder.ArrayOrString("deploy"),
		},
		pipelinev1.ParamSpec{
			Name:        "NAMESPACE",
			Type:        pipelinev1.ParamTypeString,
			Description: "Namespace to deploy into",
		},
		pipelinev1.ParamSpec{
			Name:        "DRYRUN",
			Type:        pipelinev1.ParamTypeString,
			Description: "If true run a server-side dryrun.",
			Default:     builder.ArrayOrString("false"),
		},
		pipelinev1.ParamSpec{
			Name:        "YAMLPATHTOIMAGE",
			Type:        pipelinev1.ParamTypeString,
			Description: "The path to the image to replace in the yaml manifest (arg to yq)",
		},
	}
	validResource := createTaskResource("source", "git")
	validimageResource := createTaskResource("image", "image")

	inputs := pipelinev1.Inputs{
		Resources: []pipelinev1.TaskResource{validResource, validimageResource},
		Params:    params,
	}

	return inputs

}

func createStepsFromKubectl() []pipelinev1.Step {
	Steps := []pipelinev1.Step{
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
			},
		},
	}

	return Steps

}
