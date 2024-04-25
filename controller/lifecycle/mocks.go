package lifecycle

//go:generate go run -mod=mod github.com/vektra/mockery/v2 --srcpkg=sigs.k8s.io/controller-runtime/pkg/client --name=Client --case=underscore --with-expecter
//go:generate go run -mod=mod github.com/vektra/mockery/v2 --srcpkg=sigs.k8s.io/controller-runtime/pkg/client --name=SubResourceClient --case=underscore --with-expecter
