//go:build integration
// +build integration

package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/maksemen2/pvz-service/internal/app"
	"github.com/maksemen2/pvz-service/internal/delivery/http/httpdto"
	"github.com/maksemen2/pvz-service/internal/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testClient struct {
	t     *testing.T
	url   string
	token string
}

func (c *testClient) doRequest(method, path string, body interface{}) *http.Response {
	var buf bytes.Buffer
	if body != nil {
		require.NoError(c.t, json.NewEncoder(&buf).Encode(body))
	}

	req, err := http.NewRequest(method, c.url+path, &buf)
	require.NoError(c.t, err)

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(c.t, err)

	return resp
}

func (c *testClient) createPVZ(body httpdto.PostPvzJSONRequestBody) httpdto.PVZ {
	resp := c.doRequest(http.MethodPost, "/pvz", body)
	defer resp.Body.Close()

	require.Equal(c.t, http.StatusCreated, resp.StatusCode, "unexpected status code when creating PVZ")

	var pvz httpdto.PVZ

	require.NoError(c.t, json.NewDecoder(resp.Body).Decode(&pvz))

	return pvz
}

func (c *testClient) createReception(body httpdto.PostReceptionsJSONRequestBody) httpdto.Reception {
	resp := c.doRequest(http.MethodPost, "/receptions", body)
	defer resp.Body.Close()

	require.Equal(c.t, http.StatusCreated, resp.StatusCode, "unexpected status code when creating reception")

	var reception httpdto.Reception

	require.NoError(c.t, json.NewDecoder(resp.Body).Decode(&reception))

	return reception
}

func (c *testClient) addProduct(body httpdto.PostProductsJSONRequestBody) httpdto.Product {
	resp := c.doRequest(http.MethodPost, "/products", body)
	defer resp.Body.Close()

	require.Equal(c.t, http.StatusCreated, resp.StatusCode, "unexpected status code when adding product")

	var product httpdto.Product

	require.NoError(c.t, json.NewDecoder(resp.Body).Decode(&product))

	return product
}

func (c *testClient) closeReception(pvzID uuid.UUID) httpdto.Reception {
	path := fmt.Sprintf("/pvz/%s/close_last_reception", pvzID)

	resp := c.doRequest(http.MethodPost, path, nil)
	defer resp.Body.Close()

	require.Equal(c.t, http.StatusOK, resp.StatusCode, "unexpected status code when closing reception")

	var reception httpdto.Reception

	require.NoError(c.t, json.NewDecoder(resp.Body).Decode(&reception))

	return reception
}

func TestWorkflow(t *testing.T) {
	cfg, cleanupContainer := testhelpers.SetupTestEnvironment(t)

	a, err := app.Initialize(cfg)
	require.NoError(t, err)

	cleanupDB, err := testhelpers.CreateTestDB(a.Database)
	require.NoError(t, err)

	defer func() {
		cleanupDB()
		cleanupContainer()
	}()

	router := a.BuildRouter()

	ts := httptest.NewServer(router)
	defer ts.Close()

	newClient := func(token string) *testClient {
		return &testClient{
			t:     t,
			url:   ts.URL,
			token: token,
		}
	}

	var (
		moderatorToken string
		employeeToken  string
	)

	t.Run("get auth tokens", func(t *testing.T) {
		resp := newClient("").doRequest(
			http.MethodPost,
			"/dummyLogin",
			httpdto.PostDummyLoginJSONRequestBody{
				Role: httpdto.PostDummyLoginJSONBodyRoleModerator,
			},
		)
		defer resp.Body.Close()
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&moderatorToken))

		resp = newClient("").doRequest(
			http.MethodPost,
			"/dummyLogin",
			httpdto.PostDummyLoginJSONRequestBody{
				Role: httpdto.PostDummyLoginJSONBodyRoleEmployee,
			},
		)
		defer resp.Body.Close()
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&employeeToken))
	})

	moderatorClient := newClient(moderatorToken)
	employeeClient := newClient(employeeToken)

	t.Run("full workflow", func(t *testing.T) {
		var pvz httpdto.PVZ

		t.Run("create PVZ", func(t *testing.T) {
			pvz = moderatorClient.createPVZ(httpdto.PostPvzJSONRequestBody{
				City: "Москва",
			})

			assert.NotEqual(t, uuid.Nil, *pvz.Id, "PVZ ID should not be empty")
			assert.Equal(t, "Москва", string(pvz.City), "PVZ city should match request")
		})

		var reception httpdto.Reception

		t.Run("create reception", func(t *testing.T) {
			reception = employeeClient.createReception(httpdto.PostReceptionsJSONRequestBody{
				PvzId: *pvz.Id,
			})

			assert.Equal(t, *pvz.Id, reception.PvzId, "Reception PVZ ID should match created PVZ")
			assert.Equal(t, httpdto.InProgress, reception.Status, "New reception should be in progress")
			assert.NotEqual(t, uuid.Nil, *reception.Id, "Reception ID should not be empty")
		})

		t.Run("add 50 products", func(t *testing.T) {
			for i := 0; i < 50; i++ {
				product := employeeClient.addProduct(httpdto.PostProductsJSONRequestBody{
					Type:  "электроника",
					PvzId: *pvz.Id,
				})

				assert.Equal(t, *reception.Id, product.ReceptionId,
					"Product should be added to correct reception")
				assert.NotEqual(t, uuid.Nil, *product.Id,
					"Product ID should not be empty")
			}
		})

		t.Run("close reception", func(t *testing.T) {
			updatedReception := employeeClient.closeReception(*pvz.Id)

			assert.Equal(t, *pvz.Id, updatedReception.PvzId,
				"Reception PVZ ID should remain the same")
			assert.Equal(t, httpdto.Close, updatedReception.Status,
				"Reception should be closed")
		})
	})
}
