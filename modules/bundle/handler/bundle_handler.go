package handler

import (
	"net/http"

	"github.com/webdevtedxuniversitasairlangga/modules/bundle/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/bundle/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type (
	BundleHandler interface {
		Create(ctx *gin.Context)
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
		AddImage(ctx *gin.Context)
		DeleteImage(ctx *gin.Context)
	}

	bundleHandler struct {
		bundleService service.BundleService
	}
)

func NewBundleHandler(injector *do.Injector, bs service.BundleService) BundleHandler {
	return &bundleHandler{
		bundleService: bs,
	}
}

func (c *bundleHandler) Create(ctx *gin.Context) {
	var req dto.BundleCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.bundleService.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_BUNDLE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_BUNDLE, result)
	ctx.JSON(http.StatusCreated, res)
}

// GetAll adalah endpoint publik: tidak ada pengambilan user_id dari context,
// karena bundle bukan data milik-user melainkan katalog global.
func (c *bundleHandler) GetAll(ctx *gin.Context) {
	var query dto.BundleQueryRequest
	if err := ctx.ShouldBindQuery(&query); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.bundleService.GetAll(ctx.Request.Context(), query)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_BUNDLE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_BUNDLE, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *bundleHandler) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	result, err := c.bundleService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_BUNDLE, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_BUNDLE, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *bundleHandler) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dto.BundleUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.bundleService.Update(ctx.Request.Context(), id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_BUNDLE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_BUNDLE, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *bundleHandler) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.bundleService.Delete(ctx.Request.Context(), id); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_BUNDLE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_BUNDLE, nil)
	ctx.JSON(http.StatusOK, res)
}

func (c *bundleHandler) AddImage(ctx *gin.Context) {
	bundleID := ctx.Param("id")

	var req dto.BundleImageCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.bundleService.AddImage(ctx.Request.Context(), bundleID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_ADD_BUNDLE_IMAGE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_ADD_BUNDLE_IMAGE, result)
	ctx.JSON(http.StatusCreated, res)
}

func (c *bundleHandler) DeleteImage(ctx *gin.Context) {
	bundleID := ctx.Param("id")
	imageID := ctx.Param("imageId")

	if err := c.bundleService.DeleteImage(ctx.Request.Context(), bundleID, imageID); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_BUNDLE_IMAGE, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_BUNDLE_IMAGE, nil)
	ctx.JSON(http.StatusOK, res)
}
