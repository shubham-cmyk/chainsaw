package kubectl

import (
	"testing"
	"time"

	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWaitForResource(t *testing.T) {
	tests := []struct {
		name    string
		waiter  *v1alpha1.Wait
		want    *v1alpha1.Command
		wantErr bool
	}{
		{
			name:    "nil waiter",
			waiter:  nil,
			wantErr: false,
		},
		{
			name:    "empty waiter",
			waiter:  &v1alpha1.Wait{},
			wantErr: true,
		},
		{
			name: "valid resource and condition",
			waiter: &v1alpha1.Wait{
				Resource:  "pods",
				Condition: "Ready",
			},
			want: &v1alpha1.Command{
				Entrypoint: "kubectl",
				Args:       []string{"wait", "pods", "--for=condition=Ready", "-n", "$NAMESPACE"},
			},
			wantErr: false,
		},
		{
			name: "with resource name",
			waiter: &v1alpha1.Wait{
				Resource: "pods",
				ObjectLabelsSelector: v1alpha1.ObjectLabelsSelector{
					Name: "my-pod",
				},
				Condition: "Ready",
			},
			want: &v1alpha1.Command{
				Entrypoint: "kubectl",
				Args:       []string{"wait", "pods/my-pod", "--for=condition=Ready", "-n", "$NAMESPACE"},
			},
			wantErr: false,
		},
		{
			name: "with selector",
			waiter: &v1alpha1.Wait{
				Resource:  "pods",
				Condition: "Ready",
				ObjectLabelsSelector: v1alpha1.ObjectLabelsSelector{
					Selector: "app=my-app",
				},
			},
			want: &v1alpha1.Command{
				Entrypoint: "kubectl",
				Args:       []string{"wait", "pods", "--for=condition=Ready", "-l", "app=my-app", "-n", "$NAMESPACE"},
			},
			wantErr: false,
		},
		{
			name: "with timeout",
			waiter: &v1alpha1.Wait{
				Resource:  "pods",
				Condition: "Ready",
				Timeout:   &metav1.Duration{Duration: 120 * time.Second},
			},
			want: &v1alpha1.Command{
				Timeout:    &metav1.Duration{Duration: 120 * time.Second},
				Entrypoint: "kubectl",
				Args:       []string{"wait", "pods", "--for=condition=Ready", "-n", "$NAMESPACE", "--timeout=2m0s"},
			},
			wantErr: false,
		},
		{
			name: "all namespaces",
			waiter: &v1alpha1.Wait{
				Resource:      "pods",
				Condition:     "Ready",
				AllNamespaces: true,
			},
			want: &v1alpha1.Command{
				Entrypoint: "kubectl",
				Args:       []string{"wait", "pods", "--for=condition=Ready", "--all-namespaces"},
			},
			wantErr: false,
		},
		{
			name: "name and selector error",
			waiter: &v1alpha1.Wait{
				Resource: "pods",
				ObjectLabelsSelector: v1alpha1.ObjectLabelsSelector{
					Selector: "app=my-app",
					Name:     "my-pod",
				},
				Condition: "Ready",
			},
			wantErr: true,
		},
		{
			name: "missing condition",
			waiter: &v1alpha1.Wait{
				Resource: "pods",
				ObjectLabelsSelector: v1alpha1.ObjectLabelsSelector{
					Name: "my-pod",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WaitForResource(tt.waiter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
