package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"unifriend-api/handlers"
	"unifriend-api/models"
	"unifriend-api/tests/factory"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func TestCreateConnectionRequestSuccess(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user1 := factory.UserFactory()
	user2 := factory.UserFactory()

	models.DB.Create(&user1)
	models.DB.Create(&user2)
	req, _ := http.NewRequest("POST", "/api/connections/request/user/"+fmt.Sprintf("%d", user2.ID), nil)
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user1.ID),
		Path:  "/",
	}

    absPath, err := filepath.Abs("json-schemas/test_accept_connection_request.json")
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	schemaLoader := gojsonschema.NewReferenceLoader("file://" + absPath)

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

    loader := gojsonschema.NewStringLoader(rec.Body.String())
	result, err := gojsonschema.Validate(schemaLoader, loader)
	if err != nil {
		log.Fatalf("Error validating schema: %v", err)
	}

	assert.Equal(t, http.StatusCreated, rec.Code)
	var connection models.ConnectionRequest
	models.DB.Where("requesting_user_id", user1.ID).First(&connection)

	assert.Equal(t, user1.ID, connection.RequestingUserID)
	assert.Equal(t, user2.ID, connection.RequestedUserID)
	assert.Equal(t, models.StatusPending, connection.Status)
    assert.Equal(t, true, result.Valid())
}

func TestCreateConnectionWhenThereIsUnacceptedRequest(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user1 := factory.UserFactory()
	user2 := factory.UserFactory()

	models.DB.Create(&user1)
	models.DB.Create(&user2)

	connectionRequest := factory.ConnectionRequestFactory()
	connectionRequest.RequestingUser = user1
	connectionRequest.RequestedUser = user2
	models.DB.Create(&connectionRequest)

	req, _ := http.NewRequest("POST", "/api/connections/request/user/"+fmt.Sprintf("%d", user2.ID), nil)
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user1.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateConnectionWhenUsersAlreadyConnected(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user1 := factory.UserFactory()
	user2 := factory.UserFactory()

	models.DB.Create(&user1)
	models.DB.Create(&user2)

	connection := factory.ConnectionFactory()
	connection.UserA = user1
	connection.UserB = user2
	models.DB.Create(&connection)

	req, _ := http.NewRequest("POST", "/api/connections/request/user/"+fmt.Sprintf("%d", user2.ID), nil)
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{
		Name:  "auth_token",
		Value: factory.GetUserFactoryToken(user1.ID),
		Path:  "/",
	}

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetConnectionRequestsSuccess(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    user := factory.UserFactory()
    otherUser := factory.UserFactory()
    models.DB.Create(&user)
    models.DB.Create(&otherUser)

    userResponse := factory.UserResponseFactory()
    userResponse.User = user
    models.DB.Create(&userResponse)

    otherUserResponse := factory.UserResponseFactory()
    otherUserResponse.User = otherUser
    models.DB.Create(&otherUserResponse)

    connReq := models.ConnectionRequest{
        RequestingUser: otherUser,
        RequestedUser:  user,
        Status:           models.StatusPending,
    }
    models.DB.Create(&connReq)

    req, _ := http.NewRequest("GET", "/api/connections/requests", nil)
    req.Header.Set("Content-Type", "application/json")
    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(user.ID),
        Path:  "/",
    }
    req.AddCookie(authCookie)

    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)
    assert.Equal(t, http.StatusOK, rec.Code)

    var response map[string][]handlers.ConnectionRequestResponse
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)
	data := response["data"]
    assert.Len(t, data, 1)
    assert.Equal(t, connReq.ID, data[0].ID)
    assert.Zero(t, data[0].RequestingUser.Score)
    assert.Equal(t, otherUser.ID, data[0].RequestingUserID)
}

func TestAcceptConnectionRequest(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    requestingUser := factory.UserFactory()
    models.DB.Create(&requestingUser)

    requestedUser := factory.UserFactory()
    models.DB.Create(&requestedUser)

    connectionRequest := models.ConnectionRequest{
        RequestingUser: requestingUser,
        RequestedUser:  requestedUser,
        Status:           models.StatusPending,
    }

    models.DB.Create(&connectionRequest)

    url := "/api/connections/requests/" + strconv.FormatUint(uint64(connectionRequest.ID), 10) + "/accept"
    req, _ := http.NewRequest("PUT", url, nil)

    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(requestedUser.ID),
    }
    req.AddCookie(authCookie)

    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response gin.H
    json.Unmarshal(rec.Body.Bytes(), &response)
    assert.Equal(t, "Connection request accepted", response["message"])

    var updatedRequest models.ConnectionRequest
    models.DB.First(&updatedRequest, connectionRequest.ID)
    assert.Equal(t, 1, updatedRequest.Status)
    assert.NotNil(t, updatedRequest.AnswerAt)

    var newConnection models.Connection
    err := models.DB.Where("user_a = ? AND user_b = ?", requestingUser.ID, requestedUser.ID).First(&newConnection).Error
    assert.NoError(t, err)
    assert.NotZero(t, newConnection.ID)
    assert.Equal(t, connectionRequest.ID, newConnection.ConnectionRequestID)
}

func TestRejectConnectionRequest(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    requestingUser := factory.UserFactory()
    models.DB.Create(&requestingUser)

    requestedUser := factory.UserFactory()
    models.DB.Create(&requestedUser)

    connectionRequest := models.ConnectionRequest{
        RequestingUser: requestingUser,
        RequestedUser:  requestedUser,
        Status:           models.StatusPending,
    }

    models.DB.Create(&connectionRequest)

    url := "/api/connections/requests/" + strconv.FormatUint(uint64(connectionRequest.ID), 10) + "/reject"
    req, _ := http.NewRequest("PUT", url, nil)

    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(requestedUser.ID),
    }
    req.AddCookie(authCookie)

    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response gin.H
    json.Unmarshal(rec.Body.Bytes(), &response)
    assert.Equal(t, "Connection request rejected", response["message"])

    var updatedRequest models.ConnectionRequest
    models.DB.First(&updatedRequest, connectionRequest.ID)
    assert.Equal(t, 0, updatedRequest.Status)
    assert.NotNil(t, updatedRequest.AnswerAt)
}

func TestDeleteConnection(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    requestingUser := factory.UserFactory()
    models.DB.Create(&requestingUser)

    requestedUser := factory.UserFactory()
    models.DB.Create(&requestedUser)

    connectionRequest := models.Connection{
        UserA: requestingUser,
        UserB:  requestedUser,
    }

    models.DB.Create(&connectionRequest)

    connection := factory.ConnectionFactory()
    connection.UserA = requestingUser
    connection.UserB = requestedUser
    models.DB.Create(&connection)

    url := "/api/connections/" + strconv.FormatUint(uint64(connection.ID), 10)
    req, _ := http.NewRequest("DELETE", url, nil)

    authCookie := &http.Cookie{
        Name:  "auth_token",
        Value: factory.GetUserFactoryToken(requestedUser.ID),
    }
    req.AddCookie(authCookie)

    rec := httptest.NewRecorder()
    router.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var newConnection models.Connection
    err := models.DB.Where("id = ?", connection.ID).First(&newConnection).Error
    assert.Error(t, err)
    assert.Zero(t, newConnection.ID)
}
