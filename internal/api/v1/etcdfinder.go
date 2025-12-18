package v1

import (
	"fmt"
	"net/http"

	"github.com/etcdfinder/etcdfinder/internal/api/dto"
	"github.com/etcdfinder/etcdfinder/internal/service"
	"github.com/gin-gonic/gin"
)

type EtcdfinderHandler struct {
	etcdSvcClt service.Etcdfinder
}

func NewEtcdfinderHandler(etcdSvcClt service.Etcdfinder) *EtcdfinderHandler {
	return &EtcdfinderHandler{
		etcdSvcClt: etcdSvcClt,
	}
}

func (e *EtcdfinderHandler) GetKey(c *gin.Context) {
	var req dto.GetKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("invalid request: %w", err)) //nolint
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(err) //nolint
		return
	}

	resp, err := e.etcdSvcClt.GetKey(c.Request.Context(), req.Key)
	if err != nil {
		c.Error(err) //nolint
		return
	}

	c.JSON(http.StatusOK, dto.GetKeyResponse{
		Key:   req.Key,
		Value: resp,
	})
}

func (e *EtcdfinderHandler) SearchKeys(c *gin.Context) {
	var req dto.SearchKeysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("invalid request: %w", err)) //nolint
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(err) //nolint
		return
	}

	resp, err := e.etcdSvcClt.SearchKeys(c.Request.Context(), req.SearchStr)
	if err != nil {
		c.Error(err) //nolint
		return
	}

	c.JSON(http.StatusOK, dto.SearchKeysResponse{
		Keys: resp,
	})
}

func (e *EtcdfinderHandler) PutKey(c *gin.Context) {
	var req dto.PutKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("invalid request: %w", err)) //nolint
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(err) //nolint
		return
	}

	if err := e.etcdSvcClt.PutKey(c.Request.Context(), req.Key, req.Value); err != nil {
		c.Error(err) //nolint
		return
	}

	c.JSON(http.StatusOK, dto.PutKeyResponse(req))
}

func (e *EtcdfinderHandler) DeleteKey(c *gin.Context) {
	var req dto.DeleteKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(fmt.Errorf("invalid request: %w", err)) //nolint
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(err) //nolint
		return
	}

	if err := e.etcdSvcClt.DeleteKey(c.Request.Context(), req.Key); err != nil {
		c.Error(err) //nolint
		return
	}

	c.JSON(http.StatusOK, dto.DeleteKeyResponse(req))
}

func (e *EtcdfinderHandler) GetIngestionDelay(c *gin.Context) {
	c.JSON(http.StatusOK, dto.GetIngestionDelayResponse{
		IngestionDelay: e.etcdSvcClt.GetIngestionDelay(c.Request.Context()),
	})
}
