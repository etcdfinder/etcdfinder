package dto

import "github.com/etcdfinder/etcdfinder/internal/customerrors"

type GetKeyRequest struct {
	Key string `json:"key"`
}

func (g *GetKeyRequest) Validate() error {
	if g.Key == "" {
		return customerrors.ErrKeyRequired
	}
	return nil
}

type GetKeyResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SearchKeysRequest struct {
	SearchStr string `json:"search_str"`
}

func (s *SearchKeysRequest) Validate() error {
	return nil
}

type SearchKeysResponse struct {
	Keys []string `json:"keys"`
}

type PutKeyRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (p *PutKeyRequest) Validate() error {
	if p.Key == "" {
		return customerrors.ErrKeyRequired
	}
	if p.Value == "" {
		return customerrors.ErrValueRequired
	}
	return nil
}

type PutKeyResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DeleteKeyRequest struct {
	Key string `json:"key"`
}

func (d *DeleteKeyRequest) Validate() error {
	if d.Key == "" {
		return customerrors.ErrKeyRequired
	}
	return nil
}

type DeleteKeyResponse struct {
	Key string `json:"key"`
}

type GetIngestionDelayResponse struct {
	IngestionDelay int `json:"ingestion_delay"`
}
