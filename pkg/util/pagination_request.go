package util

import (
	"net/url"
	"strconv"
)

type PaginationReq struct {
	Size uint64 `json:"size"`
	Page uint64 `json:"page"`
}

func NewPaginationReq(size uint64, page uint64) *PaginationReq {
	return &PaginationReq{Size: size, Page: page}
}

func (r *PaginationReq) ProcessQueryParams(q url.Values) error {
	limit, err := r.parseValue(q.Get("size"), "10", "size param error")
	if err != nil {
		return err
	}
	page, err := r.parseValue(q.Get("page"), "1", "page param error")
	if err != nil {
		return err
	}

	r.Size = limit
	r.Page = page
	return nil
}

func (r *PaginationReq) ProcessQueryParamsOptional(q url.Values) error {
	size, err := r.parseValue(q.Get("size"), "0", "size param error")
	if err != nil {
		return err
	}

	page, err := r.parseValue(q.Get("page"), "0", "page param error")
	if err != nil {
		return err
	}

	r.Size = size
	r.Page = page
	return nil
}

func (r *PaginationReq) parseValue(v, defaultValue, errMsg string) (uint64, error) {
	if v == "" || v == "0" {
		v = defaultValue
	}

	intVal, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}

	return intVal, nil
}

func (r *PaginationReq) GetDBOffset() uint64 {
	return (r.Page - 1) * r.Size
}
