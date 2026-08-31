package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"toll_plaza/internal/domain"
	"toll_plaza/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTollService struct {
	mock.Mock
}

func (m *MockTollService) GetTollsBetweenPincodes(ctx context.Context, req domain.TollRequest) (*domain.TollResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TollResponse), args.Error(1)
}

func setupTestRouter(mockService *MockTollService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	tollHandler := handler.NewTollHandler(mockService)
	r.POST("/api/v1/toll-plazas", tollHandler.GetTollPlazas)
	return r
}

func TestTollHandler_Success(t *testing.T) {
	mockService := new(MockTollService)
	r := setupTestRouter(mockService)

	expectedResponse := &domain.TollResponse{
		Route: domain.RouteSummary{
			SourcePincode:      "110001",
			DestinationPincode: "560001",
			DistanceInKm:       2100,
		},
		TollPlazas: []domain.TollPlazaResponse{
			{
				Name:               "Toll Plaza 1",
				Latitude:           28.7041,
				Longitude:          77.1025,
				DistanceFromSource: 200,
			},
			{
				Name:               "Toll Plaza 2",
				Latitude:           19.0760,
				Longitude:          72.8777,
				DistanceFromSource: 1400,
			},
		},
	}

	reqPayload := domain.TollRequest{
		SourcePincode:      "110001",
		DestinationPincode: "560001",
	}
	mockService.On("GetTollsBetweenPincodes", mock.Anything, reqPayload).Return(expectedResponse, nil)

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/toll-plazas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp domain.TollResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "110001", resp.Route.SourcePincode)
	assert.Equal(t, "560001", resp.Route.DestinationPincode)
	assert.Equal(t, 2100, resp.Route.DistanceInKm)
	assert.Len(t, resp.TollPlazas, 2)
}

func TestTollHandler_InvalidPincode(t *testing.T) {
	mockService := new(MockTollService)
	r := setupTestRouter(mockService)

	reqPayload := domain.TollRequest{
		SourcePincode:      "010001",
		DestinationPincode: "560001",
	}
	mockService.On("GetTollsBetweenPincodes", mock.Anything, reqPayload).Return(nil, domain.ErrInvalidPincode)

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/toll-plazas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errResp domain.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "Invalid source or destination pincode", errResp.Error)
}

func TestTollHandler_SamePincode(t *testing.T) {
	mockService := new(MockTollService)
	r := setupTestRouter(mockService)

	reqPayload := domain.TollRequest{
		SourcePincode:      "110001",
		DestinationPincode: "110001",
	}
	mockService.On("GetTollsBetweenPincodes", mock.Anything, reqPayload).Return(nil, domain.ErrSamePincode)

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/toll-plazas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errResp domain.ErrorResponse
	_ = json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, "Source and destination pincodes cannot be the same", errResp.Error)
}

func TestTollHandler_NoTollPlazas(t *testing.T) {
	mockService := new(MockTollService)
	r := setupTestRouter(mockService)

	expectedResponse := &domain.TollResponse{
		Route: domain.RouteSummary{
			SourcePincode:      "110001",
			DestinationPincode: "110002",
			DistanceInKm:       5,
		},
		TollPlazas: []domain.TollPlazaResponse{},
	}

	reqPayload := domain.TollRequest{
		SourcePincode:      "110001",
		DestinationPincode: "110002",
	}
	mockService.On("GetTollsBetweenPincodes", mock.Anything, reqPayload).Return(expectedResponse, nil)

	body, _ := json.Marshal(reqPayload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/toll-plazas", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp domain.TollResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 5, resp.Route.DistanceInKm)
	assert.Empty(t, resp.TollPlazas)
}
