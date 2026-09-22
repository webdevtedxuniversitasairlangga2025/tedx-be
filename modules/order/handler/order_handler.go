package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/order/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/order/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
)

type OrderHandler interface {
	Create(ctx *gin.Context)
	GetByID(ctx *gin.Context)
	GetMyOrders(ctx *gin.Context)
	GetAll(ctx *gin.Context)
	Approve(ctx *gin.Context)
	Reject(ctx *gin.Context)
	UploadProof(ctx *gin.Context)
}

type orderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(injector *do.Injector, s service.OrderService) OrderHandler {
	return &orderHandler{orderService: s}
}

func (h *orderHandler) Create(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	var req dto.OrderCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.Create(ctx.Request.Context(), userID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_ORDER, result)
	ctx.JSON(http.StatusCreated, res)
}

func (h *orderHandler) GetByID(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")
	result, err := h.orderService.GetByID(ctx.Request.Context(), userID, id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *orderHandler) GetMyOrders(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	var pagination dto.PaginationRequest
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.GetMyOrders(ctx.Request.Context(), userID, pagination)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *orderHandler) GetAll(ctx *gin.Context) {
	var pagination dto.PaginationRequest
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	var filter dto.OrderFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.GetAll(ctx.Request.Context(), filter, pagination)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *orderHandler) Approve(ctx *gin.Context) {
	adminID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")
	result, err := h.orderService.Approve(ctx.Request.Context(), adminID, id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_APPROVE_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_APPROVE_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *orderHandler) Reject(ctx *gin.Context) {
	adminID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")
	var req dto.OrderRejectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.Reject(ctx.Request.Context(), adminID, id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_REJECT_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_REJECT_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func uploadToImageKit(fileHeader *multipart.FileHeader) (string, error) {
	privateKey := os.Getenv("IMAGEKIT_PRIVATE_KEY")
	endpoint := os.Getenv("IMAGEKIT_URL_ENDPOINT")
	if privateKey == "" {
		return "", fmt.Errorf("IMAGEKIT_PRIVATE_KEY not set")
	}
	if endpoint == "" {
		endpoint = "https://ik.imagekit.io/tedxunair"
	}
	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("file", "data:"+fileHeader.Header.Get("Content-Type")+";base64,"+b64)
	_ = writer.WriteField("fileName", fmt.Sprintf("%s%s", uuid.New().String(), filepath.Ext(fileHeader.Filename)))
	_ = writer.WriteField("folder", "/orders")
	_ = writer.WriteField("useUniqueFileName", "true")
	writer.Close()
	req, err := http.NewRequest("POST", "https://upload.imagekit.io/api/v1/files/upload", body)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(privateKey, "")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respData, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("imagekit upload failed %d: %s", resp.StatusCode, string(respData))
	}
	var out struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(respData, &out); err != nil {
		return "", err
	}
	if out.URL == "" {
		return "", fmt.Errorf("imagekit no url")
	}
	return out.URL, nil
}

func (h *orderHandler) UploadProof(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")
	contentType := ctx.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, err := ctx.FormFile("file")
		if err != nil {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPLOAD_PROOF, "only jpg, jpeg, png, webp allowed", nil)
			ctx.JSON(http.StatusBadRequest, res)
			return
		}
		if file.Size > 5<<20 {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPLOAD_PROOF, "file too large, max 5MB", nil)
			ctx.JSON(http.StatusBadRequest, res)
			return
		}
		url, err := uploadToImageKit(file)
		if err != nil {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPLOAD_PROOF, err.Error(), nil)
			ctx.JSON(http.StatusBadRequest, res)
			return
		}
		req := dto.OrderUploadProofRequest{PaymentProofURL: url}
		result, err := h.orderService.UploadProof(ctx.Request.Context(), userID, id, req)
		if err != nil {
			res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPLOAD_PROOF, err.Error(), nil)
			ctx.JSON(http.StatusBadRequest, res)
			return
		}
		res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPLOAD_PROOF, result)
		ctx.JSON(http.StatusOK, res)
		return
	}
	var req dto.OrderUploadProofRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.UploadProof(ctx.Request.Context(), userID, id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPLOAD_PROOF, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPLOAD_PROOF, result)
	ctx.JSON(http.StatusOK, res)
}
