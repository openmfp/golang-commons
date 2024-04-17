package lifecycle

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/openmfp/golang-commons/controller/testSupport"
	"github.com/openmfp/golang-commons/logger/testlogger"
	"github.com/openmfp/golang-commons/sentry"
)

func TestLifecycle(t *testing.T) {
	namespace := "bar"
	name := "foo"
	request := controllerruntime.Request{
		NamespacedName: types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
	}
	ctx := context.Background()

	t.Run("Lifecycle with a not found object", func(t *testing.T) {
		// Arrange
		fakeClient := testSupport.CreateFakeClient(t, &testSupport.TestApiObject{})

		mgr, logger := createLifecycleManager([]Subroutine{}, fakeClient)

		// Act
		result, err := mgr.Reconcile(ctx, request, &testSupport.TestApiObject{})

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		logMessages, err := logger.GetLogMessages()
		assert.NoError(t, err)
		assert.Equal(t, len(logMessages), 2)
		assert.Equal(t, logMessages[0].Message, "start reconcile")
		assert.Contains(t, logMessages[1].Message, "instance not found")
	})

	t.Run("Lifecycle with a finalizer - add finalizer", func(t *testing.T) {
		// Arrange
		instance := &testSupport.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{
			finalizerSubroutine{
				client: fakeClient,
			},
		}, fakeClient)

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(instance.ObjectMeta.Finalizers))
	})

	t.Run("Lifecycle with a finalizer - finalization", func(t *testing.T) {
		// Arrange
		now := &metav1.Time{Time: time.Now()}
		finalizers := []string{finalizer}
		instance := &testSupport.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				DeletionTimestamp: now,
				Finalizers:        finalizers,
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{
			finalizerSubroutine{
				client: fakeClient,
			},
		}, fakeClient)

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(instance.ObjectMeta.Finalizers))
	})

	t.Run("Lifecycle with a finalizer - skip finalization if the finalizer is not in there", func(t *testing.T) {
		// Arrange
		now := &metav1.Time{Time: time.Now()}
		finalizers := []string{"other-finalizer"}
		instance := &testSupport.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				DeletionTimestamp: now,
				Finalizers:        finalizers,
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{
			finalizerSubroutine{
				client: fakeClient,
			},
		}, fakeClient)

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(instance.ObjectMeta.Finalizers))
	})
	t.Run("Lifecycle with a finalizer - failing finalization subroutine", func(t *testing.T) {
		// Arrange
		now := &metav1.Time{Time: time.Now()}
		finalizers := []string{finalizer}
		instance := &testSupport.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				DeletionTimestamp: now,
				Finalizers:        finalizers,
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{
			finalizerSubroutine{
				client: fakeClient,
				err:    fmt.Errorf("some error"),
			},
		}, fakeClient)

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.Error(t, err)
		assert.Equal(t, 1, len(instance.ObjectMeta.Finalizers))
	})

	t.Run("Lifecycle without changing status", func(t *testing.T) {
		// Arrange
		instance := &testSupport.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: testSupport.TestStatus{Some: "string"},
		}
		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, logger := createLifecycleManager([]Subroutine{}, fakeClient)

		// Act
		result, err := mgr.Reconcile(ctx, request, instance)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		logMessages, err := logger.GetLogMessages()
		assert.NoError(t, err)
		assert.Equal(t, len(logMessages), 3)
		assert.Equal(t, logMessages[0].Message, "start reconcile")
		assert.Equal(t, logMessages[1].Message, "skipping status update, since they are equal")
		assert.Equal(t, logMessages[2].Message, "end reconcile")
	})

	t.Run("Lifecycle with changing status", func(t *testing.T) {
		// Arrange
		instance := &testSupport.TestApiObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Status: testSupport.TestStatus{Some: "string"},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, logger := createLifecycleManager([]Subroutine{
			changeStatusSubroutine{
				client: fakeClient,
			},
		}, fakeClient)

		// Act
		result, err := mgr.Reconcile(ctx, request, instance)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		logMessages, err := logger.GetLogMessages()
		assert.NoError(t, err)
		assert.Equal(t, len(logMessages), 5)
		assert.Equal(t, logMessages[0].Message, "start reconcile")
		assert.Equal(t, logMessages[1].Message, "start subroutine")
		assert.Equal(t, logMessages[2].Message, "end subroutine")
		assert.Equal(t, logMessages[3].Message, "updating resource status")
		assert.Equal(t, logMessages[4].Message, "end reconcile")

		serverObject := &testSupport.TestApiObject{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, serverObject)
		assert.NoError(t, err)
		assert.Equal(t, serverObject.Status.Some, "other string")
	})

	t.Run("Lifecycle with spread reconciles", func(t *testing.T) {
		// Arrange
		instance := &implementingSpreadReconciles{
			testSupport.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: testSupport.TestStatus{
					Some:               "string",
					ObservedGeneration: 0,
				},
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{
			changeStatusSubroutine{
				client: fakeClient,
			},
		}, fakeClient)
		mgr.WithSpreadingReconciles()

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.NoError(t, err)
		assert.Equal(t, instance.Generation, instance.Status.ObservedGeneration)
	})

	t.Run("Lifecycle with spread reconciles on deleted object", func(t *testing.T) {
		// Arrange
		instance := &implementingSpreadReconciles{
			testSupport.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:              name,
					Namespace:         namespace,
					Generation:        2,
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
					Finalizers:        []string{finalizer},
				},
				Status: testSupport.TestStatus{
					Some:               "string",
					ObservedGeneration: 2,
					NextReconcileTime:  metav1.Time{Time: time.Now().Add(2 * time.Hour)},
				},
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{
			changeStatusSubroutine{
				client: fakeClient,
			},
		}, fakeClient)
		mgr.WithSpreadingReconciles()

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)
		assert.NoError(t, err)
		assert.Len(t, instance.Finalizers, 0)

	})

	t.Run("Lifecycle with spread reconciles skips if the generation is the same", func(t *testing.T) {
		// Arrange
		instance := &implementingSpreadReconciles{
			testSupport.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: testSupport.TestStatus{
					Some:               "string",
					ObservedGeneration: 1,
				},
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{failureScenarioSubroutine{Retry: false, RequeAfter: false}}, fakeClient)
		mgr.WithSpreadingReconciles()

		// Act
		result, err := mgr.Reconcile(ctx, request, instance)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), instance.Status.ObservedGeneration)
		assert.GreaterOrEqual(t, 12*time.Hour, result.RequeueAfter)
	})

	t.Run("Lifecycle with spread reconciles and processing fails", func(t *testing.T) {
		// Arrange
		instance := &implementingSpreadReconciles{
			testSupport.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: testSupport.TestStatus{
					Some:               "string",
					ObservedGeneration: 0,
				},
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{failureScenarioSubroutine{Retry: false, RequeAfter: false}}, fakeClient)
		mgr.WithSpreadingReconciles()

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.Error(t, err)
		assert.Equal(t, int64(0), instance.Status.ObservedGeneration)
	})

	t.Run("Lifecycle with spread reconciles and processing needs requeue", func(t *testing.T) {
		// Arrange
		instance := &implementingSpreadReconciles{
			testSupport.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: testSupport.TestStatus{
					Some:               "string",
					ObservedGeneration: 0,
				},
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{failureScenarioSubroutine{Retry: true, RequeAfter: false}}, fakeClient)
		mgr.WithSpreadingReconciles()

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), instance.Status.ObservedGeneration)
	})

	t.Run("Lifecycle with spread reconciles and processing needs requeueAfter", func(t *testing.T) {
		// Arrange
		instance := &implementingSpreadReconciles{
			testSupport.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
				},
				Status: testSupport.TestStatus{
					Some:               "string",
					ObservedGeneration: 0,
				},
			},
		}
		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{failureScenarioSubroutine{Retry: false, RequeAfter: true}}, fakeClient)
		mgr.WithSpreadingReconciles()

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), instance.Status.ObservedGeneration)
	})

	t.Run("Lifecycle with spread not implementing the interface", func(t *testing.T) {
		// Arrange
		instance := &notImplementingSpreadReconciles{
			ObjectMeta: metav1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Generation: 1,
			},
			Status: testSupport.TestStatus{
				Some:               "string",
				ObservedGeneration: 0,
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		mgr, _ := createLifecycleManager([]Subroutine{
			changeStatusSubroutine{
				client: fakeClient,
			},
		}, fakeClient)
		mgr.WithSpreadingReconciles()

		// Act
		_, err := mgr.Reconcile(ctx, request, instance)

		assert.Errorf(t, err, "SpreadReconciles is enabled, but instance does not implement RuntimeObjectSpreadReconcileStatus interface. This is a programming error.")
	})

	t.Run("Should setup with manager", func(t *testing.T) {
		// Arrange
		instance := &testSupport.TestApiObject{}
		fakeClient := testSupport.CreateFakeClient(t, instance)
		m, err := manager.New(&rest.Config{}, manager.Options{
			Scheme: fakeClient.Scheme(),
		})
		assert.NoError(t, err)

		lm, _ := createLifecycleManager([]Subroutine{failureScenarioSubroutine{Retry: false, RequeAfter: true}}, fakeClient)
		tr := &testReconciler{
			lifecycleManager: lm,
		}

		// Act
		err = lm.SetupWithManager(m, 0, "testReconciler", instance, "test", tr)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("Lifecycle with spread reconciles and refresh label", func(t *testing.T) {
		// Arrange
		instance := &implementingSpreadReconciles{
			testSupport.TestApiObject{
				ObjectMeta: metav1.ObjectMeta{
					Name:       name,
					Namespace:  namespace,
					Generation: 1,
					Labels:     map[string]string{SpreadReconcileRefreshLabel: "true"},
				},
				Status: testSupport.TestStatus{
					Some:               "string",
					ObservedGeneration: 1,
				},
			},
		}

		fakeClient := testSupport.CreateFakeClient(t, instance)

		lm, _ := createLifecycleManager([]Subroutine{
			changeStatusSubroutine{
				client: fakeClient,
			},
		}, fakeClient)
		lm.WithSpreadingReconciles()

		// Act
		_, err := lm.Reconcile(ctx, request, instance)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), instance.Status.ObservedGeneration)

		serverObject := &implementingSpreadReconciles{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, serverObject)
		assert.NoError(t, err)
		assert.Equal(t, serverObject.Status.Some, "other string")
		_, ok := serverObject.Labels[SpreadReconcileRefreshLabel]
		assert.False(t, ok)
	})

	t.Run("Should handle a client error", func(t *testing.T) {
		// Arrange
		lm, log := createLifecycleManager([]Subroutine{}, nil)

		testErr := fmt.Errorf("test error")

		// Act
		result, err := lm.handleClientError("test", log.Logger, testErr, sentry.Tags{})

		// Assert
		assert.Error(t, err)
		assert.Equal(t, testErr, err)
		assert.Equal(t, controllerruntime.Result{}, result)
	})
}

type testReconciler struct {
	lifecycleManager *LifecycleManager
}

func (r *testReconciler) Reconcile(ctx context.Context, req controllerruntime.Request) (controllerruntime.Result, error) {
	return r.lifecycleManager.Reconcile(ctx, req, &testSupport.TestApiObject{})
}

func createLifecycleManager(subroutines []Subroutine, c client.Client) (*LifecycleManager, *testlogger.TestLogger) {
	log := testlogger.New()
	mgr := NewLifecycleManager(log.Logger, "test-operator", "test-controller", c, subroutines)
	return mgr, log
}
