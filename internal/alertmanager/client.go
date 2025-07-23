/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alertmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/silence-operator/silence-operator/api/v1alpha1"
)

type AlertManagerInterface interface {
	GetSilence(id string) (*silence.GetSilenceOK, error)
	GetSilences(filter []string) (*silence.GetSilencesOK, error)
	UpsertSilence(s *v1alpha1.Silence, startsAt *strfmt.DateTime) (string, error)
	DeleteSilence(id string) error
}

type AlertManager struct {
	Host            string
	Author          string
	InstanceName    string
	SilenceDuration time.Duration

	am *client.AlertmanagerAPI
}

func (c *AlertManager) GetSilences(filter []string) (*silence.GetSilencesOK, error) {
	return c.am.Silence.GetSilences(&silence.GetSilencesParams{
		Filter: filter,
	})
}

func (c *AlertManager) GetSilence(id string) (*silence.GetSilenceOK, error) {
	return c.am.Silence.GetSilence(&silence.GetSilenceParams{
		SilenceID: strfmt.UUID(id),
	})
}

// UpsertSilence will check if there is a silence with the same matchers.
// It will update it if it exists and create a new one if it doesn't.
func (c *AlertManager) UpsertSilence(ctx context.Context, s *v1alpha1.Silence, startsAt *strfmt.DateTime) (string, error) {
	log := ctrl.LoggerFrom(ctx)

	if s.Status.AlertManagerID == "" {
		found := false

		filter := s.Spec.Matchers.String()

		result, err := c.GetSilences(filter)
		if err != nil {
			return "", err
		}

		existingSilences := result.GetPayload()

		for _, existingSilence := range existingSilences {
			if *existingSilence.Status.State == models.SilenceStatusStateExpired {
				continue
			}

			if len(existingSilence.Matchers) == len(s.Spec.Matchers) {
				log.Info("found an existing silence, updating existing silence", "silence", existingSilence.ID)

				s.Status.AlertManagerID = *existingSilence.ID
				found = true

				break
			}
		}

		if !found {
			log.Info("no existing silence found, new one will be created")
		}
	}

	matchers := models.Matchers{}

	for _, m := range s.Spec.Matchers {
		matchers = append(matchers, &models.Matcher{
			IsEqual: &m.IsEqual,
			IsRegex: &m.IsRegex,
			Name:    &m.Name,
			Value:   &m.Value,
		})
	}

	now := time.Now()

	if startsAt == nil {
		nowFmt := strfmt.DateTime(now)
		startsAt = &nowFmt
	}

	endsAt := strfmt.DateTime(now.Add(c.SilenceDuration))
	comment := fmt.Sprintf("%s\nInstance: %s", s.Spec.Comment, c.InstanceName)

	result, err := c.am.Silence.PostSilences(&silence.PostSilencesParams{
		Silence: &models.PostableSilence{
			ID: s.Status.AlertManagerID,
			Silence: models.Silence{
				Comment:   &comment,
				CreatedBy: &c.Author,
				EndsAt:    &endsAt,
				StartsAt:  startsAt,
				Matchers:  matchers,
			},
		},
	})
	if err != nil {
		return "", err
	}

	newId := result.GetPayload().SilenceID
	log.Info("silence created", "id", newId)

	return newId, nil
}

func (c *AlertManager) DeleteSilence(id string) error {
	_, err := c.am.Silence.DeleteSilence(&silence.DeleteSilenceParams{
		SilenceID: strfmt.UUID(id),
	})

	return err
}

func New(host string, author string, instanceName string, silenceDuration time.Duration) *AlertManager {
	transportConfig := client.DefaultTransportConfig().WithHost(host)

	return &AlertManager{
		Host:            host,
		Author:          author,
		InstanceName:    instanceName,
		SilenceDuration: silenceDuration,

		am: client.NewHTTPClientWithConfig(strfmt.Default, transportConfig),
	}
}
