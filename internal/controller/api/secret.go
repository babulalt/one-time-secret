package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/berrybytes/simplesecrets/internal/controller/token"
	db "github.com/berrybytes/simplesecrets/internal/model/sqlc"
	"github.com/berrybytes/simplesecrets/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type createSecretRequest struct {
	Content      string `json:"content" binding:"required"`
	Hashpassword string `json:"hashpassword"`
}

var Cont string

func (server *Server) createOneTimeSecret(ctx *gin.Context) {
	var req createSecretRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hash, _ := util.HashPassword(req.Hashpassword)
	token := util.GetToken()
	args := db.CreateSecretParams{
		Content:      req.Content,
		Token:        token,
		Hashpassword: req.Hashpassword,
	}
	length := len(args.Hashpassword)
	if length == 0 {
		args.Hashpassword = ""
	} else {
		args.Hashpassword = hash
	}

	secret, err := server.store.CreateSecret(ctx, args)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	Cont = util.RandomString(60)
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		ctx.JSON(http.StatusUnprocessableEntity, errorResponse(errors.New("required baseURL")))
		return
	}
	url := fmt.Sprintf("%s/one-time-secret/%d/%s", baseURL, secret.ID, Cont)
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"url": url,
	})
}

func (server *Server) createSecret(ctx *gin.Context) {
	var req createSecretRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hash, _ := util.HashPassword(req.Hashpassword)
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	args := db.CreateSecretParams{
		Content:      req.Content,
		Creator:      authPayload.Username,
		Hashpassword: req.Hashpassword,
	}
	length := len(args.Hashpassword)
	if length == 0 {
		args.Hashpassword = ""
	} else {
		args.Hashpassword = hash
	}

	secret, err := server.store.CreateSecret(ctx, args)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	Cont = util.RandomString(60)
	url := fmt.Sprintf("localhost:3000/secrets/%d/%s", secret.ID, Cont)
	ctx.JSON(http.StatusOK, url)
}

// ShowAccount godoc
// @Summary      Show an account
// @Description  get string by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  db.Secret
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /secret/{id} [get]
func (server *Server) getOneTimeSecret(ctx *gin.Context) {
	var req getSecretRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	secret, err := server.store.GetSecret(ctx, req.ID)
	logrus.Info("Request ID", req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if !secret.Isview {
		ctx.JSON(http.StatusOK, secret.Content)
		args := db.UpdateSecretParams{
			ID:     req.ID,
			Isview: true,
		}
		secret, err := server.store.UpdateSecret(ctx, args)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		logrus.Info("view secret :: ", secret)
	} else {
		ctx.JSON(http.StatusOK, "You have already view")
	}

}

type getSecretRequest struct {
	ID     int64 `uri:"id" binding:"required,min=1"`
	Isview bool
}

// ShowAccount godoc
// @Summary      Show an account
// @Description  get string by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  db.Secret
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /secret/{id} [get]
func (server *Server) getSecret(ctx *gin.Context) {
	var req getSecretRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	secret, err := server.store.GetSecret(ctx, req.ID)
	fmt.Println(req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if secret.Creator != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	if !secret.Isview {
		ctx.JSON(http.StatusOK, secret.Content)
		args := db.UpdateSecretParams{
			ID:     req.ID,
			Isview: true,
		}
		secret, err := server.store.UpdateSecret(ctx, args)
		if err != nil {
			panic(err)
		}
		fmt.Println(secret)
	} else {
		ctx.JSON(http.StatusOK, "You have already view")
	}

}

type listSecretRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listSecrets(ctx *gin.Context) {
	var req listSecretRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.ListSecretsParams{
		Creator: authPayload.Username,
		Limit:   req.PageSize,
		Offset:  (req.PageID - 1) * req.PageSize,
	}
	secrets, err := server.store.ListSecrets(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, secrets)
}
