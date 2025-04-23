package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	mockdb "github.com/lucasHSantiago/gobank/internal/db/mock"
	db "github.com/lucasHSantiago/gobank/internal/db/sqlc"
	"github.com/lucasHSantiago/gobank/internal/db/util"
	"github.com/lucasHSantiago/gobank/internal/token"
	"github.com/stretchr/testify/require"
)

func TestRenewAccessTokenAPI(t *testing.T) {
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		setupToken    func(maker token.Maker) (gin.H, *token.Payload, db.Session)
		buildStubs    func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload := randomToken(t, maker, user)
				session := randomSession(user, refreshToken)
				body := gin.H{
					"refresh_token": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), refreshPayload.ID).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload := randomToken(t, maker, user)
				session := randomSession(user, refreshToken)
				body := gin.H{
					"invalid_field": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ErrorVerifyToken",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload, err := maker.CreateToken(user.Username, util.DepositorRole, -time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, refreshToken)
				require.NotEmpty(t, refreshPayload)

				session := randomSession(user, refreshToken)
				body := gin.H{
					"refresh_token": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "NoTokenFound",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload := randomToken(t, maker, user)
				session := randomSession(user, refreshToken)
				body := gin.H{
					"refresh_token": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, db.ErrRecordNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload := randomToken(t, maker, user)
				session := randomSession(user, refreshToken)
				body := gin.H{
					"refresh_token": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Session{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "SessionBlocked",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload := randomToken(t, maker, user)

				session := randomSession(user, refreshToken)
				session.IsBlocked = true

				body := gin.H{
					"refresh_token": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), refreshPayload.ID).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "IncorrectSessionUser",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload := randomToken(t, maker, user)

				session := randomSession(user, refreshToken)
				session.Username = "invalid"

				body := gin.H{
					"refresh_token": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), refreshPayload.ID).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "MismatchedSessionToken",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload := randomToken(t, maker, user)

				session := randomSession(user, refreshToken)
				session.RefreshToken = "invalid"

				body := gin.H{
					"refresh_token": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), refreshPayload.ID).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SessionExpired",
			setupToken: func(maker token.Maker) (gin.H, *token.Payload, db.Session) {
				refreshToken, refreshPayload := randomToken(t, maker, user)

				session := randomSession(user, refreshToken)
				session.ExpiresAt = time.Now().Add(-time.Minute)

				body := gin.H{
					"refresh_token": refreshToken,
				}

				return body, refreshPayload, session
			},
			buildStubs: func(store *mockdb.MockStore, refreshPayload token.Payload, session db.Session) {
				store.EXPECT().
					GetSession(gomock.Any(), refreshPayload.ID).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			server := newTestServer(t, store)

			body, refreshPayload, session := tc.setupToken(server.tokenMaker)
			tc.buildStubs(store, *refreshPayload, session)

			recorder := httptest.NewRecorder()

			data, err := json.Marshal(body)
			require.NoError(t, err)

			url := "/token/renew_access"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			request.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomSession(user db.User, refreshToken string) db.Session {
	return db.Session{
		ID:           uuid.New(),
		Username:     user.Username,
		RefreshToken: refreshToken,
		IsBlocked:    false,
		ExpiresAt:    time.Now().Add(time.Minute * 5),
		CreatedAt:    time.Now().Add(-time.Minute),
	}
}

func randomToken(t *testing.T, maker token.Maker, user db.User) (string, *token.Payload) {
	refreshToken, refreshPayload, err := maker.CreateToken(user.Username, util.DepositorRole, time.Minute*5)
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)
	require.NotEmpty(t, refreshPayload)

	return refreshToken, refreshPayload
}
