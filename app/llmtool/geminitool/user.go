package geminitool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gatsu420/kisu-be/app/repository/bqrepo"
	"golang.org/x/oauth2"
	"google.golang.org/genai"
)

type Aggregation struct {
	Func   string `json:"func"`
	Column string `json:"column"`
}
type GetUserArgs struct {
	Columns      []string      `json:"columns"`
	Emails       []string      `json:"emails"`
	Aggregations []Aggregation `json:"aggregations"`
}

var getUserDeclaration = &genai.FunctionDeclaration{
	Name: "getUser",
	Description: `Get user information based on their email.
	Available columns: email, name, age.
	When the prompt requires aggregation (sum(), count(), avg(), etc), use aggregations parameter
	for aggregate queries and columns parameter for group-by fields.`,
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"columns": {
				Type:        genai.TypeArray,
				Items:       &genai.Schema{Type: genai.TypeString},
				Description: "informations that need to be checked",
			},
			"emails": {
				Type:        genai.TypeArray,
				Items:       &genai.Schema{Type: genai.TypeString},
				Description: "user email",
			},
			"aggregations": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"func": {
							Type:        genai.TypeString,
							Description: "aggregation function to apply",
						},
						"column": {
							Type:        genai.TypeString,
							Description: "column that is being aggregated",
						},
					},
				},
				Description: "what aggregations the prompt requires",
			},
		},
		Required: []string{"columns", "emails"},
	},
	Response: &genai.Schema{
		Type:        genai.TypeObject,
		Description: "key-value pairs where keys are column names or aggregation column like max_age, avg_age, etc. aggregated column are named as {func}_{column}",
	},
}

func (t *toolImpl) getUser(ctx context.Context, token *oauth2.Token, rawArgs json.RawMessage) (json.RawMessage, error) {
	var args GetUserArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal raw args: %w", err)
	}

	aggregations := []bqrepo.Aggregation{}
	for _, a := range args.Aggregations {
		aggregations = append(aggregations, bqrepo.Aggregation{
			Func:   a.Func,
			Column: a.Column,
		})
	}

	result, err := t.bqRepo.GetUser(ctx, bqrepo.GetUserArgs{
		Token:        token,
		Columns:      args.Columns,
		Emails:       args.Emails,
		Aggregations: aggregations,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get result: %w", err)
	}

	marshaledResult, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal result: %w", err)
	}

	return marshaledResult, nil
}

func (t *toolImpl) GetUser() WiringItem {
	return WiringItem{
		Declaration: getUserDeclaration,
		Func:        t.getUser,
	}
}
