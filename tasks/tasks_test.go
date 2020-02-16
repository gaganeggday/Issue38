package tasks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	"github.com/tektoncd/pipeline/test/builder"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildah(t *testing.T) {
	// func TestBuildah() {

	var task pipelinev1.Task
	task.ObjectMeta = v1.ObjectMeta{
		Name: "buildah-task",
	}
	task.TypeMeta = v1.TypeMeta{
		Kind:       "Task",
		APIVersion: "tekton.dev/v1alpha1",
	}
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
	validInputResource := pipelinev1.TaskResource{
		ResourceDeclaration: pipelinev1.ResourceDeclaration{
			Name: "source",
			Type: "git",
		},
	}

	validOutputResource := pipelinev1.TaskResource{
		ResourceDeclaration: pipelinev1.ResourceDeclaration{
			Name: "image",
			Type: "image",
		},
	}
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

	inputs := &pipelinev1.Inputs{
		Resources: []pipelinev1.TaskResource{validInputResource},
		Params:    params,
	}
	outputs := &pipelinev1.Outputs{
		Resources: []pipelinev1.TaskResource{validOutputResource},
	}

	volumes := []corev1.Volume{{
		Name: "varlibcontainers",
		// EmptyDir: ,
	}}

	task.Spec = pipelinev1.TaskSpec{
		Inputs:  inputs,
		Outputs: outputs,
		Steps:   steps,
		Volumes: volumes,
	}
	deployFromSourceTask := AssembleBuildah()
	if diff := cmp.Diff(task, deployFromSourceTask); diff != "" {
		t.Fatalf("GenerateDeployFromSourceTask() failed \n%s", diff)
	}

}
func TestGithubStatusTask(t *testing.T) {
	wantedTask := pipelinev1.Task{
		TypeMeta: v1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "create-github-status-task",
		},
		Spec: pipelinev1.TaskSpec{
			Inputs: &pipelinev1.Inputs{
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
			},
			TaskSpec: v1alpha2.TaskSpec{
				Steps: []pipelinev1.Step{
					pipelinev1.Step{
						Container: corev1.Container{
							Name:       "start-status",
							Image:      "quay.io/kmcdermo/github-tool:latest",
							WorkingDir: "/workspace/source",
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name: "GITHUB_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "github-auth",
											},
											Key: "token",
										},
									},
								},
							},
							Command: []string{"github-tools"},
							Args: []string{
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
							},
						},
					},
				},
			},
		},
	}

	githubStatusTask := GenerateGithubStatusTask()
	if diff := cmp.Diff(wantedTask, githubStatusTask); diff != "" {
		t.Fatalf("GenerateGithubStatusTask() failed:\n%s", diff)
	}
}

func TestDeployFromSourceTask(t *testing.T) {
	wantedTask := pipelinev1.Task{
		TypeMeta: v1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "deploy-from-source-task",
		},
		Spec: pipelinev1.TaskSpec{
			Inputs: &pipelinev1.Inputs{
				Resources: []pipelinev1.TaskResource{
					pipelinev1.TaskResource{
						ResourceDeclaration: pipelinev1.ResourceDeclaration{
							Name: "source",
							Type: "git",
						},
					},
				},
				Params: []pipelinev1.ParamSpec{
					pipelinev1.ParamSpec{
						Name:        "PATHTODEPLOYMENT",
						Description: "Path to the manifest to apply",
						Type:        pipelinev1.ParamTypeString,
						Default: &pipelinev1.ArrayOrString{
							StringVal: "deploy",
						},
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
						Default: &pipelinev1.ArrayOrString{
							StringVal: "false",
						},
					},
				},
			},
			TaskSpec: v1alpha2.TaskSpec{
				Steps: []pipelinev1.Step{
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
				},
			},
		},
	}
	deployFromSourceTask := GenerateDeployFromSourceTask()
	if diff := cmp.Diff(wantedTask, deployFromSourceTask); diff != "" {
		t.Fatalf("GenerateDeployFromSourceTask() failed \n%s", diff)
	}
}

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
