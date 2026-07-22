package geminitool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gatsu420/kisu-be/app/repository/bqrepo"
	"golang.org/x/oauth2"
	"google.golang.org/genai"
)

type GetSellerArgs struct {
	TableName   string `json:"table_name"`
	ViewName    string `json:"view_name"`
	ParamColumn string `json:"param_column"`
	Query       string `json:"query"`
}

var getSellerDeclaration = &genai.FunctionDeclaration{
	Name: "rumah-aya.some_event.merchants",
	Description: `
	Table name is rumah-aya.some_event.merchants, while view name is
	that and added with _hashed. Query selects from view, not from table.

	The view has these columns:
	- email (string): seller email.
	- name (string): seller name
	- shop_id (integer): ID of shop each seller has
	- age (integer): seller age
	- param (string): param value for filtering row

	Param is created by hashing email, so set param_column as "email".
	Always include "email" in select statement, and include "param" in
	where statement.

	Sample query using the view:
	- get seller information
	  select
		name,
		email,
		shop_id,
		age,
	  from rumah-aya.some_event.merchants_hashed
	- get average age per seller
	  select
	  	name,
		email,
		avg(age)
	  from rumah-aya.some_event.merchants_hashed
	  group by 1, 2
	`,
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"table_name": {
				Type:        genai.TypeString,
				Description: "Data source which is used to create view",
			},
			"view_name": {
				Type:        genai.TypeString,
				Description: "Data source which is used for query to select from",
			},
			"param": {
				Type:        genai.TypeString,
				Description: "param value for filtering view",
			},
			"param_column": {
				Type:        genai.TypeString,
				Description: "What column to use to filter param value",
			},
			"query": {
				Type:        genai.TypeString,
				Description: "Query to get wanted information",
			},
		},
		Required: []string{"param", "query", "table_name", "view_name"},
	},
	Response: &genai.Schema{
		Type:        genai.TypeArray,
		Items:       &genai.Schema{Type: genai.TypeString},
		Description: "List of information returned by query, 1 item represents 1 row",
	},
}

func (t *toolImpl) getSeller(ctx context.Context, token *oauth2.Token, rawArgs json.RawMessage) (json.RawMessage, error) {
	var args GetSellerArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal raw args: %w", err)
	}

	result, err := t.bqRepo.GetInformation(ctx, bqrepo.GetSellerArgs{
		Token:       token,
		TableName:   args.TableName,
		ViewName:    args.ViewName,
		ParamColumn: args.ParamColumn,
		Query:       args.Query,
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

func (t *toolImpl) GetSeller() WiringItem {
	return WiringItem{
		Declaration: getSellerDeclaration,
		Func:        t.getSeller,
	}
}
