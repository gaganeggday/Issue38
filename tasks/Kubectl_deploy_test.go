package tasks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/test/builder"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGithubStatusTask(t *testing.T) {
	wantedTask := v1alpha1.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "deploy-using-kubectl-task",
		},
		Spec: v1alpha1.TaskSpec{
			Inputs: &v1alpha1.Inputs{
				Resources: []v1alpha1.TaskResource{
					v1alpha1.TaskResource{
						ResourceDeclaration: v1alpha1.ResourceDeclaration{
							Name: "source",
							Type: "git",
						},
					},
					v1alpha1.TaskResource{
						ResourceDeclaration: v1alpha1.ResourceDeclaration{
							Name: "image",
							Type: "image",
						},
					},
				},
				Params: []v1alpha1.ParamSpec{
					v1alpha1.ParamSpec{
						Name:        "PATHTODEPLOYMENT",
						Type:        v1alpha1.ParamTypeString,
						Description: "Path to the manifest to apply",
						Default:     builder.ArrayOrString("deploy"),
					},
					v1alpha1.ParamSpec{
						Name:        "NAMESPACE",
						Type:        v1alpha1.ParamTypeString,
						Description: "Namespace to deploy into",
					},
					v1alpha1.ParamSpec{
						Name:        "DRYRUN",
						Type:        v1alpha1.ParamTypeString,
						Description: "If true run a server-side dryrun.",
						Default:     builder.ArrayOrString("false"),
					},
					v1alpha1.ParamSpec{
						Name:        "YAMLPATHTOIMAGE",
						Type:        v1alpha1.ParamTypeString,
						Description: "The path to the image to replace in the yaml manifest (arg to yq)",
					},
				},
			},
			Steps: []v1alpha1.Step{
				v1alpha1.Step{
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
				v1alpha1.Step{
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
			},
		},
	}

	githubStatusTask := GenerateKubectlTask()
	if diff := cmp.Diff(wantedTask, githubStatusTask); diff != "" {
		t.Errorf("GenerateGithubStatusTask() failed:\n%s", diff)
	}
}
