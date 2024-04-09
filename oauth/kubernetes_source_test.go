package oauth_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/openmfp/golang-commons/oauth"
	"github.com/stretchr/testify/assert"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func TestNoValidRestConfig(t *testing.T) {
	_, err := oauth.NewKubernetesTokenSource(nil, &oauth.KubernetesTokenSourceConfig{})
	assert.Error(t, err)
}

func TestExpirationOutOfRange(t *testing.T) {
	_, err := oauth.NewKubernetesTokenSource(&rest.Config{}, &oauth.KubernetesTokenSourceConfig{
		ServiceAccount:  types.NamespacedName{Name: "default", Namespace: "default"},
		TokenExpiration: 5 * time.Minute,
	})
	assert.EqualError(t, oauth.ErrTokenExpirationOutOfRange, err.Error())
}

func TestKubernetesTokenSource(t *testing.T) {
	expiration := time.Now().Add(1 * time.Hour).Round(time.Minute)
	serviceAccountName := types.NamespacedName{Name: "default", Namespace: "default"}

	mux := http.NewServeMux()
	mux.HandleFunc(
		fmt.Sprintf("/api/v1/namespaces/%s/serviceaccounts/%s/token", serviceAccountName.Namespace, serviceAccountName.Name),
		func(w http.ResponseWriter, r *http.Request) {

			err := json.NewEncoder(w).Encode(&authenticationv1.TokenRequest{
				Status: authenticationv1.TokenRequestStatus{
					Token:               "the-test-token",
					ExpirationTimestamp: metav1.NewTime(expiration),
				},
			})
			assert.NoError(t, err)
		},
	)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	source, err := oauth.NewKubernetesTokenSource(&rest.Config{Host: srv.URL}, &oauth.KubernetesTokenSourceConfig{
		ServiceAccount: serviceAccountName,
	})
	assert.NoError(t, err)

	token, err := source.Token()
	assert.NoError(t, err)
	assert.Equal(t, "the-test-token", token.AccessToken)
	assert.Equal(t, expiration, token.Expiry)
}

func TestKubernetesTokenSourceInvalid(t *testing.T) {
	serviceAccountName := types.NamespacedName{Name: "default", Namespace: "default"}

	mux := http.NewServeMux()
	mux.HandleFunc(
		fmt.Sprintf("/api/v1/namespaces/%s/serviceaccounts/%s/token", serviceAccountName.Namespace, serviceAccountName.Name),
		func(w http.ResponseWriter, r *http.Request) {

			_, err := w.Write([]byte("some invalid json"))
			assert.NoError(t, err)
		},
	)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	source, err := oauth.NewKubernetesTokenSource(&rest.Config{Host: srv.URL}, &oauth.KubernetesTokenSourceConfig{
		ServiceAccount: serviceAccountName,
	})
	assert.NoError(t, err)

	_, err = source.Token()
	assert.Error(t, err)
}

func TestKubernetesTokenSourceInvalidHttpConnection(t *testing.T) {
	serviceAccountName := types.NamespacedName{Name: "default", Namespace: "default"}

	source, err := oauth.NewKubernetesTokenSource(&rest.Config{Host: "test"}, &oauth.KubernetesTokenSourceConfig{
		ServiceAccount: serviceAccountName,
	})
	assert.NoError(t, err)

	_, err = source.Token()
	assert.Error(t, err)
}
