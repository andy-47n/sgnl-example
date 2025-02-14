package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	framework "github.com/sgnl-ai/adapter-framework"
	api_adapter_v1 "github.com/sgnl-ai/adapter-framework/api/adapter/v1"
)

const (
	Teams string = "teams"
)

type Entity struct {
	uniqueIDAttrExternalID string
}

type Datasource struct {
	Client *http.Client
}

type DatasourceResponse struct {
	Teams  []map[string]any `json:"teams,omitempty"`
	Limit  int64            `json:"limit"`
	Offset int64            `json:"offset"`
	More   bool             `json:"more"`
}

var (
	ValidEntityExternalIDs = map[string]Entity{
		Teams: {
			uniqueIDAttrExternalID: "id",
		},
	}
)

func NewClient(timeout int) Client {
	return &Datasource{
		Client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func (d *Datasource) GetPage(ctx context.Context, request *Request) (*Response, *framework.Error) {
	var req *http.Request

	url := fmt.Sprintf("%s/%s?limit=%d&offset=%s", request.BaseURL, request.EntityExternalID, request.PageSize, request.Cursor)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &framework.Error{
			Message: "Failed to create HTTP request to datasource.",
			Code:    api_adapter_v1.ErrorCode_ERROR_CODE_INTERNAL,
		}
	}

	// Timeout API calls that take longer than 5 seconds
	apiCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req = req.WithContext(apiCtx)

	req.Header.Add("Accept", "application/vnd.pagerduty+json;version=2")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", request.Token)

	res, err := d.Client.Do(req)
	if err != nil {
		return nil, &framework.Error{
			Message: "Failed to send request to datasource.",
			Code:    api_adapter_v1.ErrorCode_ERROR_CODE_INTERNAL,
		}
	}

	response := &Response{
		StatusCode:       res.StatusCode,
		RetryAfterHeader: res.Header.Get("Retry-After"),
	}

	if res.StatusCode != http.StatusOK {
		return response, nil
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &framework.Error{
			Message: "Failed to read response body.",
			Code:    api_adapter_v1.ErrorCode_ERROR_CODE_DATASOURCE_FAILED,
		}
	}

	objects, nextCursor, parseErr := ParseResponse(body)
	if parseErr != nil {
		return nil, parseErr
	}

	response.Objects = objects
	response.NextCursor = nextCursor

	return response, nil
}

func ParseResponse(body []byte) (objects []map[string]any, nextCursor string, err *framework.Error) {
	var data *DatasourceResponse

	unmarshalErr := json.Unmarshal(body, &data)
	if unmarshalErr != nil {
		return nil, "", &framework.Error{
			Message: fmt.Sprintf("Failed to unmarshal the datasource response: %v.", unmarshalErr),
			Code:    api_adapter_v1.ErrorCode_ERROR_CODE_INTERNAL,
		}
	}

	if data.More {
		nextCursor = fmt.Sprintf("%d", data.Limit+data.Offset)
	} else {
		nextCursor = ""
	}

	return data.Teams, nextCursor, nil
}
