package testSupport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

func CreateFakeClient(t *testing.T, objects ...client.Object) client.WithWatch {
	builder := fake.NewClientBuilder()
	s := runtime.NewScheme()
	sBuilder := scheme.Builder{GroupVersion: schema.GroupVersion{Group: "test.openmfp.com", Version: "v1alpha1"}}
	for _, obj := range objects {
		sBuilder.Register(obj)
		builder.WithStatusSubresource(obj)
	}
	err := sBuilder.AddToScheme(s)
	assert.NoError(t, err)
	builder.WithScheme(s)
	builder.WithObjects(objects...)
	return builder.Build()
}
