package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	framework "github.com/sgnl-ai/adapter-framework"
	api_adapter_v1 "github.com/sgnl-ai/adapter-framework/api/adapter/v1"
	"github.com/sgnl-ai/adapter-framework/web"
)

type Adapter struct {
	Client Client
}

func NewAdapter(client Client) framework.Adapter[Config] {
	return &Adapter{
		Client: client,
	}
}

func (a *Adapter) GetPage(ctx context.Context, request *framework.Request[Config]) framework.Response {
	if err := a.ValidateGetPageRequest(ctx, request); err != nil {
		return framework.NewGetPageResponseError(err)
	}

	return a.RequestPageFromDatasource(ctx, request)
}

func (a *Adapter) RequestPageFromDatasource(
	ctx context.Context, request *framework.Request[Config],
) framework.Response {

	if !strings.HasPrefix(request.Address, "https://") {
		request.Address = "https://" + request.Address
	}
	req := &Request{
		BaseURL: request.Address,

		Token: request.Auth.HTTPAuthorization,

		PageSize:         request.PageSize,
		EntityExternalID: request.Entity.ExternalId,
		Cursor:           request.Cursor,
	}

	resp, err := a.Client.GetPage(ctx, req)
	if err != nil {
		return framework.NewGetPageResponseError(err)
	}

	if adapterErr := web.HTTPError(resp.StatusCode, resp.RetryAfterHeader); adapterErr != nil {
		return framework.NewGetPageResponseError(adapterErr)
	}

	parsedObjects, parserErr := web.ConvertJSONObjectList(
		&request.Entity,
		resp.Objects,

		web.WithJSONPathAttributeNames(),

		web.WithDateTimeFormats(
			[]web.DateTimeFormatWithTimeZone{
				{Format: time.RFC3339, HasTimeZone: true},
				{Format: time.RFC3339Nano, HasTimeZone: true},
				{Format: "2006-01-02T15:04:05.000Z0700", HasTimeZone: true},
				{Format: "2006-01-02", HasTimeZone: false},
			}...,
		),
	)
	if parserErr != nil {
		return framework.NewGetPageResponseError(
			&framework.Error{
				Message: fmt.Sprintf("Failed to convert datasource response objects: %v.", parserErr),
				Code:    api_adapter_v1.ErrorCode_ERROR_CODE_INTERNAL,
			},
		)
	}

	page := &framework.Page{
		Objects: parsedObjects,
	}

	page.NextCursor = resp.NextCursor

	return framework.NewGetPageResponseSuccess(page)
}
