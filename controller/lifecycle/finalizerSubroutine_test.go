package lifecycle

import (
	"context"

	"github.com/openmfp/golang-commons/controller/testSupport"
	"github.com/openmfp/golang-commons/errors"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const finalizer = "finalizer"

type finalizerSubroutine struct {
	client client.Client
}

func (c finalizerSubroutine) Process(_ context.Context, runtimeObj RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	instance := runtimeObj.(*testSupport.TestApiObject)
	instance.Status.Some = "other string"
	return controllerruntime.Result{}, nil
}

func (c finalizerSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	return controllerruntime.Result{}, nil
}

func (c finalizerSubroutine) GetName() string {
	return "changeStatus"
}

func (c finalizerSubroutine) Finalizers() []string {
	return []string{
		finalizer,
	}
}
