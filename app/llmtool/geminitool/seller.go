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
	TableName string `json:"table_name"`
	ViewName  string `json:"view_name"`
	Query     string `json:"query"`
}

var getSellerDeclaration = &genai.FunctionDeclaration{
	Name: "rumah-aya.some_event.merchants",
	Description: `
	Table name is rumah-aya.some_event.merchants, while view name is
	that and added with _hashed. Query selects from view, not from table.

	The view has these columns:
	- name (string): seller name
	- shop_id (integer): ID of shop each seller has
	- age (integer): seller age
	- hashed_email (string): hashed seller email

	Column hashed_email doesn't need to be selected. Although it's based
	on random hash per request anyway, giving it out will kinda mislead
	users into perceiving we expose param to result.

	Sample query using the view:
	- get seller information
	  select
		name,
		shop_id,
		age,
	  from rumah-aya.some_event.merchants_hashed
	- get average age per seller
	  select
	  	name,
		avg(age)
	  from rumah-aya.some_event.merchants_hashed
	  group by 1
	`,
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"hashed_email": {
				Type:        genai.TypeString,
				Description: "Hashed seller emails delimited by comma",
			},
			"table_name": {
				Type:        genai.TypeString,
				Description: "Data source which is used to create view",
			},
			"view_name": {
				Type:        genai.TypeString,
				Description: "Data source which is used for query to select from",
			},
			"query": {
				Type:        genai.TypeString,
				Description: "Query to get wanted information",
			},
		},
		Required: []string{"hashed_email", "query", "table_name", "view_name"},
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
		Token:     token,
		TableName: args.TableName,
		ViewName:  args.ViewName,
		Query:     args.Query,
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
