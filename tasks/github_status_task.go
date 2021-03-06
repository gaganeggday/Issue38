package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

// GenerateGithubStatusTask will return a github-status-task
func GenerateGithubStatusTask() pipelinev1.Task {
	task := pipelinev1.Task{
		TypeMeta:   createTaskTypeMeta(),
		ObjectMeta: createTaskObjectMeta("create-github-status-task"),
		Spec: pipelinev1.TaskSpec{
			Inputs: createInputsForGithubStatusTask(),
			TaskSpec: v1alpha2.TaskSpec{
				Steps: createStepsForGithubStatusTask(),
			},
		},
	}
	return task
}

func argsForStartStatusStep() []string {
	return []string{
		"create-status",
		"--repo",
		"$(inputs.params.REPO)",
		"--sha",
		"$(inputs.params.COMMIT_SHA)",
		"--state",
		"$(inputs.params.STATE)",
		"--target-url",
		"$(inputs.params.TARGET_URL)",
		"--description",
		"$(inputs.params.DESCRIPTION)",
		"--context",
		"$(inputs.params.CONTEXT)",
	}
}

func createStepsForGithubStatusTask() []pipelinev1.Step {
	return []pipelinev1.Step{
		pipelinev1.Step{
			Container: corev1.Container{
				Name:       "start-status",
				Image:      "quay.io/kmcdermo/github-tool:latest",
				WorkingDir: "/workspace/source",
				Env: []corev1.EnvVar{
					createEnvFromSecret("GITHUB_TOKEN", "github-auth", "token"),
				},
				Command: []string{"github-tools"},
				Args:    argsForStartStatusStep(),
			},
		},
	}
}

func createInputsForGithubStatusTask() *pipelinev1.Inputs {
	inputs := pipelinev1.Inputs{
		Params: []pipelinev1.ParamSpec{
			pipelinev1.ParamSpec{
				Name:        "REPO",
				Type:        pipelinev1.ParamTypeString,
				Description: "The repo to publish the status update for e.g. tektoncd/triggers",
			},
			pipelinev1.ParamSpec{
				Name:        "COMMIT_SHA",
				Type:        pipelinev1.ParamTypeString,
				Description: "The specific commit to report a status for.",
			},
			pipelinev1.ParamSpec{
				Name:        "STATE",
				Type:        pipelinev1.ParamTypeString,
				Description: "The state to report error, failure, pending, or success.",
			},
			pipelinev1.ParamSpec{
				Name:        "TARGET_URL",
				Type:        pipelinev1.ParamTypeString,
				Description: "The target URL to associate with this status.",
				Default: &pipelinev1.ArrayOrString{
					StringVal: "",
				},
			},
			pipelinev1.ParamSpec{
				Name:        "DESCRIPTION",
				Type:        pipelinev1.ParamTypeString,
				Description: "A short description of the status.",
			},
			pipelinev1.ParamSpec{
				Name:        "CONTEXT",
				Type:        pipelinev1.ParamTypeString,
				Description: "A string label to differentiate this status from the status of other systems.",
			},
		},
	}
	return &inputs
}
