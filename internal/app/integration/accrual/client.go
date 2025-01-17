package accrual

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/dto"
)

const (
	defaultClientTimeout   = time.Second
	defaultRetryTimeout    = 3 * time.Second
	orderStatusURLTemplate = "%s/api/orders/%d"
)

type IClient interface {
	IsAlive() bool
	GetOrderStatus(ctx context.Context, orderID int64, userID uuid.UUID) (*model.OrderStatus, error)
}

type Client struct {
	address string
	client  *http.Client
}

func NewClient(address string) *Client {
	return &Client{
		address: address,
		client: &http.Client{
			Timeout: defaultClientTimeout,
		},
	}
}

func (c *Client) IsAlive() bool {
	runAddress := strings.TrimPrefix(c.address, "http://")

	conn, err := net.DialTimeout("tcp", runAddress, c.client.Timeout)
	if err != nil {
		return false
	}

	defer conn.Close()

	return true
}

func parseRetryAfter(header string, defaultTimeout time.Duration) time.Duration {
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return defaultTimeout
}

func (c *Client) GetOrderStatus(ctx context.Context, orderID int64, userID uuid.UUID) (*model.OrderStatus, error) {
	url := fmt.Sprintf(orderStatusURLTemplate, c.address, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, e.Wrap("failed to create HTTP request", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, e.Wrap("accrual client", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		status := &dto.OrderStatusAccrual{}
		if err = easyjson.UnmarshalFromReader(resp.Body, status); err != nil {
			return nil, e.ErrJSONUnmarshal
		}

		orderID, err := strconv.ParseInt(status.ID, 10, 64)
		if err != nil {
			return nil, e.ErrJSONUnmarshal
		}

		orderStatus := &model.OrderStatus{
			ID:      orderID,
			UserID:  userID,
			Status:  status.Status,
			Accrual: status.Accrual,
		}

		return orderStatus, nil
	case http.StatusNoContent:
		return nil, e.ErrAccrualOrderNotRegistered
	case http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), defaultRetryTimeout)

		return nil, &e.AccrualTooManyRequestsError{RetryAfter: retryAfter}
	case http.StatusInternalServerError:
		return nil, e.ErrAccrualInternalServer
	default:
		return nil, e.ErrAccrualUnknownResponse
	}
}
