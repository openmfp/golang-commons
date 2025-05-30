package lifecycle

import (
	"context"
	"time"

	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmfp/golang-commons/controller/testSupport"
	openMfpErrors "github.com/openmfp/golang-commons/errors"
)

const subroutineFinalizer = "finalizer"

type finalizerSubroutine struct {
	client       client.Client
	err          error
	requeue      bool
	requeueAfter time.Duration
}

func (c finalizerSubroutine) Process(_ context.Context, runtimeObj RuntimeObject) (controllerruntime.Result, openMfpErrors.OperatorError) {
	instance := runtimeObj.(*testSupport.TestApiObject)
	instance.Status.Some = "other string"
	return controllerruntime.Result{}, nil
}

func (c finalizerSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, openMfpErrors.OperatorError) {
	if c.err != nil {
		return controllerruntime.Result{}, openMfpErrors.NewOperatorError(c.err, true, true)
	}
	if c.requeue {
		return controllerruntime.Result{Requeue: true}, nil
	}
	if c.requeueAfter > 0 {
		return controllerruntime.Result{RequeueAfter: c.requeueAfter}, nil
	}

	return controllerruntime.Result{}, nil
}

func (c finalizerSubroutine) GetName() string {
	return "changeStatus"
}

func (c finalizerSubroutine) Finalizers() []string {
	return []string{
		subroutineFinalizer,
	}
}
