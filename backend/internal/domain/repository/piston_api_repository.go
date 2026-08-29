package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"backend/internal/config"
	"backend/internal/domain/entities"
	logger "backend/internal/log"
	"backend/internal/response"
)

type PistonApiRepository struct {
	config *config.PistonConfig
}

func NewPistonApiRepository(config *config.PistonConfig) *PistonApiRepository {
	return &PistonApiRepository{
		config: config,
	}
}

func (repository *PistonApiRepository) RunSubmission(ctx context.Context, request entities.SubmittedSolutionRequest) (res response.ExecuteSubmissionResponse, err error) {
	logger.Info("PistonApiRepository: RunSubmittedSolutions")

	url := repository.config.PistonBaseApi

	jsonData, err := json.Marshal(request)
	if err != nil {
		err = response.NewInternalError(err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url+"/v2/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		err = response.NewInternalError(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = response.NewServiceUnavailable(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = response.NewBadGateway(err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = response.NewBadGateway(fmt.Errorf("piston returned %d: %s", resp.StatusCode, body))
		return
	}

	if err = json.Unmarshal(body, &res); err != nil {
		err = response.NewBadGateway(err)
		return
	}

	return
}
