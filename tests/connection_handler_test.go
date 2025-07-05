package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"unifriend-api/handlers"
	"unifriend-api/models"
	"unifriend-api/tests/factory"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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

	req.AddCookie(authCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var connection models.ConnectionRequest
	models.DB.Where("requesting_user_id", user1.ID).First(&connection)
	fmt.Println(connection)
	assert.Equal(t, user1.ID, connection.RequestingUserID)
	assert.Equal(t, user2.ID, connection.RequestedUserID)
	assert.Equal(t, models.StatusPending, connection.Status)
}

func TestCreateConnectionWhenThereIsUnacceptedRequest(t *testing.T) {
	SetupTestDB()
	defer models.TearDownTestDB()

	user1 := factory.UserFactory()
	user2 := factory.UserFactory()

	models.DB.Create(&user1)
	models.DB.Create(&user2)

	connectionRequest := factory.ConnectionRequestFactory()
	connectionRequest.RequestingUserID = user1.ID
	connectionRequest.RequestedUserID = user2.ID
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
	connection.UserAID = user1.ID
	connection.UserBID = user2.ID
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


    connReq := models.ConnectionRequest{
        RequestingUserID: otherUser.ID,
        RequestedUserID:  user.ID,
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
	fmt.Println("Response Body:", rec.Body.String())
    assert.Equal(t, http.StatusOK, rec.Code)

    var response map[string][]handlers.ConnectionRequestResponse
    err := json.Unmarshal(rec.Body.Bytes(), &response)
    assert.NoError(t, err)
	data := response["data"]
    assert.Len(t, data, 1)
    assert.Equal(t, connReq.ID, data[0].ID)
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
        RequestingUserID: requestingUser.ID,
        RequestedUserID:  requestedUser.ID,
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
}

func TestRejectConnectionRequest(t *testing.T) {
    SetupTestDB()
    defer models.TearDownTestDB()

    requestingUser := factory.UserFactory()
    models.DB.Create(&requestingUser)

    requestedUser := factory.UserFactory()
    models.DB.Create(&requestedUser)

    connectionRequest := models.ConnectionRequest{
        RequestingUserID: requestingUser.ID,
        RequestedUserID:  requestedUser.ID,
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
    assert.Equal(t, "Connection request accepted", response["message"])

    var updatedRequest models.ConnectionRequest
    models.DB.First(&updatedRequest, connectionRequest.ID)
    assert.Equal(t, 0, updatedRequest.Status)
    assert.NotNil(t, updatedRequest.AnswerAt)
}
