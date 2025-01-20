// Copyright 2024 New Vector Ltd.
// Copyright 2017 Vector Creations Ltd
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-Element-Commercial
// Please see LICENSE files in the repository root for full details.

package routing

import (
	"context"
	"fmt"
	"net/http"

	appserviceAPI "github.com/element-hq/dendrite/appservice/api"
	"github.com/element-hq/dendrite/clientapi/auth"
	"github.com/element-hq/dendrite/clientapi/auth/authtypes"
	"github.com/element-hq/dendrite/clientapi/userutil"
	roomserverAPI "github.com/element-hq/dendrite/roomserver/api"
	"github.com/element-hq/dendrite/setup/config"
	userapi "github.com/element-hq/dendrite/userapi/api"
	"github.com/matrix-org/gomatrixserverlib/spec"
	"github.com/matrix-org/util"
)

type loginResponse struct {
	UserID        string              `json:"user_id"`
	AccessToken   string              `json:"access_token"`
	DeviceID      string              `json:"device_id"`
	AccountType   userapi.AccountType `json:"account_type"`
	ParentAccount string              `json:"parent_account"`
	JoinedRooms   []string            `json:"joined_rooms"`
}

type flows struct {
	Flows []flow `json:"flows"`
}

type flow struct {
	Type string `json:"type"`
}

// Login implements GET and POST /login
func Login(
	req *http.Request,
	cfg *config.ClientAPI,
	userAPI userapi.ClientUserAPI,
	rsAPI roomserverAPI.ClientRoomserverAPI,
	asAPI appserviceAPI.AppServiceInternalAPI,
) util.JSONResponse {
	if req.Method == http.MethodGet {
		loginFlows := []flow{{Type: authtypes.LoginTypePassword}}
		if len(cfg.Derived.ApplicationServices) > 0 {
			loginFlows = append(loginFlows, flow{Type: authtypes.LoginTypeApplicationService})
		}
		// TODO: support other forms of login, depending on config options
		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: flows{
				Flows: loginFlows,
			},
		}
	} else if req.Method == http.MethodPost {
		login, cleanup, authErr := auth.LoginFromJSONReader(req, userAPI, userAPI, cfg)
		fmt.Println("[clientapi/routing/login.go][Login] login = ", login)
		if authErr != nil {
			return *authErr
		}
		// make a device/access token
		// authErr2, accessToken := completeAuth(req.Context(), cfg.Matrix, userAPI, login, req.RemoteAddr, req.UserAgent())
		authErr2, _ := completeAuth(req.Context(), cfg.Matrix, userAPI, login, req.RemoteAddr, req.UserAgent(), rsAPI)

		// TODO: neu la temp user dang nhap lan dau tien (chua co chat voi parent_account)
		// TODO: thi tao chat voi parent_account
		// var res api.QueryAccessTokenResponse
		// err := userAPI.QueryAccessToken(req.Context(), &api.QueryAccessTokenRequest{
		// 	AccessToken:      *accessToken,
		// 	AppServiceUserID: req.URL.Query().Get("user_id"),
		// }, &res)
		// if err != nil {
		// 	util.GetLogger(req.Context()).WithError(err).Error("userAPI.QueryAccessToken failed")
		// 	return util.JSONResponse{
		// 		Code: http.StatusInternalServerError,
		// 		JSON: spec.InternalServerError{},
		// 	}
		// }
		// // if res.Err != "" {
		// // 	if strings.HasPrefix(strings.ToLower(res.Err), "forbidden:") { // TODO: use actual error and no string comparison
		// // 		return util.JSONResponse{
		// // 			Code: http.StatusForbidden,
		// // 			JSON: spec.Forbidden(res.Err),
		// // 		}
		// // 	}
		// // }
		// // if res.Device == nil {
		// // 	return util.JSONResponse{
		// // 		Code: http.StatusUnauthorized,
		// // 		JSON: spec.UnknownToken("Unknown token"),
		// // 	}
		// // }
		// if res.Device != nil && res.Device.AccountType == api.AccountTypeTempUser {
		// 	// Create room
		// 	var createRequest createRoomRequest
		// 	// createRoomJsonString := `{"preset":"trusted_private_chat","visibility":"private","invite":["@minh:ct1.echat.tc"],"is_direct":true,"power_level_content_override":{"events":{"m.room.name":50,"m.room.avatar":50,"m.room.power_levels":100,"m.room.history_visibility":100,"m.room.canonical_alias":50,"m.room.tombstone":100,"m.room.server_acl":100,"m.room.encryption":100,"org.matrix.msc3401.call.member":0,"org.matrix.msc3401.call":100}},"initial_state":[{"type":"m.room.guest_access","state_key":"","content":{"guest_access":"can_join"}},{"type":"m.room.encryption","state_key":"","content":{"algorithm":"m.megolm.v1.aes-sha2"}}]}`
		// 	createRoomJsonString := fmt.Sprintf(`{"preset":"trusted_private_chat","visibility":"private","invite":["@%s"],"is_direct":true,"power_level_content_override":{"events":{"m.room.name":50,"m.room.avatar":50,"m.room.power_levels":100,"m.room.history_visibility":100,"m.room.canonical_alias":50,"m.room.tombstone":100,"m.room.server_acl":100,"m.room.encryption":100,"org.matrix.msc3401.call.member":0,"org.matrix.msc3401.call":100}},"initial_state":[{"type":"m.room.guest_access","state_key":"","content":{"guest_access":"can_join"}},{"type":"m.room.encryption","state_key":"","content":{"algorithm":"m.megolm.v1.aes-sha2"}}]}`, res.Device.ParentAccount)
		// 	fmt.Println("[Login] createRoomJsonString = ", createRoomJsonString)

		// 	// Convert JSON string to map
		// 	// var createRoomJsonObject map[string]interface{}
		// 	var createRoomJsonObject *createRoomRequest
		// 	err := json.Unmarshal([]byte(createRoomJsonString), &createRoomJsonObject)
		// 	if err != nil {
		// 		fmt.Println("[Login] Error decoding createRoomJsonString:", err)
		// 	} else {
		// 		// evTime, _ := httputil.ParseTSParam(req)
		// 		// _, roomID := createRoom(req.Context(), createRequest, res.Device, cfg, userAPI, rsAPI, asAPI, evTime)

		// 		// // Send m.direct message
		// 		// err = s.updateMDirect(ctx, oldRoomID, newRoomID, membership.Localpart, membership.Domain, roomSize)

		// 		// // Send invite to parent_account to join
		// 		// inviteJsonString := fmt.Sprintf(`{"user_id":"%s"}`, res.Device.ParentAccount)
		// 		// fmt.Println("[Login] inviteJsonString = ", inviteJsonString)

		// 		// // Convert JSON string to map
		// 		// var inviteJsonObject *threepid.MembershipRequest
		// 		// err := json.Unmarshal([]byte(inviteJsonString), &inviteJsonObject)
		// 		// if err != nil {
		// 		// 	fmt.Println("[Login] Error decoding inviteJsonObject:", err)
		// 		// } else {
		// 		// 	inviteStored, jsonErrResp := checkAndProcessThreepid(
		// 		// 		req, res.Device, inviteJsonObject, cfg, rsAPI, userAPI, *roomID, evTime,
		// 		// 	)
		// 		// 	if jsonErrResp != nil {
		// 		// 		return *jsonErrResp
		// 		// 	}

		// 		// 	// If an invite has been stored on an identity server, it means that a
		// 		// 	// m.room.third_party_invite event has been emitted and that we shouldn't
		// 		// 	// emit a m.room.member one.
		// 		// 	if inviteStored {
		// 		// 		return util.JSONResponse{
		// 		// 			Code: http.StatusOK,
		// 		// 			JSON: struct{}{},
		// 		// 		}
		// 		// 	}

		// 		// 	if inviteJsonObject.UserID == "" {
		// 		// 		return util.JSONResponse{
		// 		// 			Code: http.StatusBadRequest,
		// 		// 			JSON: spec.BadJSON("missing user_id"),
		// 		// 		}
		// 		// 	}

		// 		// 	deviceUserID, err := spec.NewUserID(res.Device.UserID, true)
		// 		// 	if err != nil {
		// 		// 		return util.JSONResponse{
		// 		// 			Code: http.StatusForbidden,
		// 		// 			JSON: spec.Forbidden("You don't have permission to kick this user, bad userID"),
		// 		// 		}
		// 		// 	}

		// 		// 	errRes := checkMemberInRoom(req.Context(), rsAPI, *deviceUserID, *roomID)
		// 		// 	if errRes != nil {
		// 		// 		return *errRes
		// 		// 	}

		// 		// 	// We already received the return value, so no need to check for an error here.
		// 		// 	sendInviteRes, _ := sendInvite(req.Context(), res.Device, *roomID, inviteJsonObject.UserID, inviteJsonObject.Reason, cfg, rsAPI, evTime)
		// 		// 	fmt.Println("[Login] sendInviteRes:", sendInviteRes)

		// 		// 	// send message
		// 		// 	// room.CreateAndInsert(t, bob, "m.room.message", map[string]interface{}{"body": "hello world"})
		// 		// 	// sendEventRes := SendEvent(req, res.Device, *roomID, eventType, nil, &res.Device.UserID, cfg, rsAPI, nil)
		// 		// 	// fmt.Println("[Login] sendEventRes:", sendEventRes)

		// 		// 	// sync

		// 		// }
		// 	}
		// }

		cleanup(req.Context(), &authErr2)
		return authErr2
	}
	return util.JSONResponse{
		Code: http.StatusMethodNotAllowed,
		JSON: spec.NotFound("Bad method"),
	}
}

func completeAuth(
	ctx context.Context, cfg *config.Global,
	userAPI userapi.ClientUserAPI,
	login *auth.Login,
	ipAddr, userAgent string,
	rsAPI roomserverAPI.ClientRoomserverAPI,
) (util.JSONResponse, *string) {
	token, err := auth.GenerateAccessToken()
	if err != nil {
		util.GetLogger(ctx).WithError(err).Error("auth.GenerateAccessToken failed")
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: spec.InternalServerError{},
		}, nil
	}

	localpart, serverName, err := userutil.ParseUsernameParam(login.Username(), cfg)
	if err != nil {
		util.GetLogger(ctx).WithError(err).Error("auth.ParseUsernameParam failed")
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: spec.InternalServerError{},
		}, nil
	}

	var performRes userapi.PerformDeviceCreationResponse
	err = userAPI.PerformDeviceCreation(ctx, &userapi.PerformDeviceCreationRequest{
		DeviceDisplayName: login.InitialDisplayName,
		DeviceID:          login.DeviceID,
		AccessToken:       token,
		Localpart:         localpart,
		ServerName:        serverName,
		IPAddr:            ipAddr,
		UserAgent:         userAgent,
	}, &performRes)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: spec.Unknown("failed to create device: " + err.Error()),
		}, nil
	}

	// get rooms this user joined
	// newlyJoinedRooms := GetJoinedRooms(res, userID)
	deviceUserID, _ := spec.NewUserID(performRes.Device.UserID, true)
	rooms, err := rsAPI.QueryRoomsForUser(ctx, *deviceUserID, "join")
	var roomIDStrs []string
	if err != nil {
		util.GetLogger(ctx).WithError(err).Error("QueryRoomsForUser failed")
	} else {
		if rooms == nil {
			roomIDStrs = []string{}
		} else {
			roomIDStrs = make([]string, len(rooms))
			for i, roomID := range rooms {
				roomIDStrs[i] = roomID.String()
			}
		}
	}

	fmt.Println("[clientapi/routing/login.go][completeAuth] login.Identifier.AccountType = ", login.Identifier.AccountType)
	fmt.Println("[clientapi/routing/login.go][completeAuth] login.Identifier.ParentAccount = ", login.Identifier.ParentAccount)

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: loginResponse{
			UserID:        performRes.Device.UserID,
			AccessToken:   performRes.Device.AccessToken,
			DeviceID:      performRes.Device.ID,
			AccountType:   login.Identifier.AccountType,
			ParentAccount: login.Identifier.ParentAccount,
			JoinedRooms:   roomIDStrs,
		},
	}, &performRes.Device.AccessToken
}
