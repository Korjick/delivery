package geo

import (
	"context"
	"delivery/internal/core/domain/kernel"
	"delivery/internal/core/ports"
	"delivery/internal/generated/clients/geopb"
	"delivery/internal/pkg/errs"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	client  geopb.GeoClient
	timeout time.Duration
}

func NewClient(host string) (ports.GeoClient, error) {
	if strings.TrimSpace(host) == "" {
		return nil, errs.NewValueIsRequired("geoServiceGrpcHost")
	}

	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("create Geo gRPC client: %w", err)
	}

	return &Client{
		conn:    conn,
		client:  geopb.NewGeoClient(conn),
		timeout: 5 * time.Second,
	}, nil
}

func (c *Client) GetLocation(ctx context.Context, street string) (kernel.Location, error) {
	if strings.TrimSpace(street) == "" {
		return kernel.Location{}, errs.NewValueIsRequired("street")
	}

	requestContext, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	response, err := c.client.GetGeolocation(requestContext, &geopb.GetGeolocationRequest{Street: street})
	if err != nil {
		return kernel.Location{}, fmt.Errorf("get geolocation: %w", err)
	}
	if response == nil || response.Location == nil {
		return kernel.Location{}, errs.NewValueIsInvalid("geo location")
	}

	location, err := kernel.NewLocation(int(response.Location.X), int(response.Location.Y))
	if err != nil {
		return kernel.Location{}, fmt.Errorf("validate Geo location: %w", err)
	}
	return *location, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
