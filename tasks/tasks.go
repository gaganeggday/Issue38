package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateTasks will return a slice of tasks
func GenerateTasks() []pipelinev1.Task {
	return []pipelinev1.Task{
		GenerateGithubStatusTask(),
		GenerateDeployFromSourceTask(),
	}
}

func createTaskTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "Task",
		APIVersion: "tekton.dev/v1alpha1",
	}
}

func createTaskObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
	}
}

func createTaskResource(name string, resourceType string) pipelinev1.TaskResource {
	return pipelinev1.TaskResource{
		ResourceDeclaration: pipelinev1.ResourceDeclaration{
			Name: name,
			Type: resourceType,
		},
	}
}

func createEnvFromSecret(name string, localObjectReference string, key string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: localObjectReference,
				},
				Key: key,
			},
		},
	}
}
