package geminitool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gatsu420/kisu-be/app/repository/bqrepo"
	"golang.org/x/oauth2"
	"google.golang.org/genai"
)

type GetUserArgs struct {
	Columns []string `json:"columns"`
	Emails  []string `json:"emails"`
}

type GetUserItem struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	Age   int    `json:"age,omitempty"`
}

var getUserDeclaration = &genai.FunctionDeclaration{
	Name: "getUser",
	Description: `Get user information based on their email. Information returned are
	email, name, and age.`,
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
		},
		Required: []string{"columns", "emails"},
	},
	Response: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"email": {Type: genai.TypeString, Description: "user email"},
			"name":  {Type: genai.TypeString, Description: "user name"},
			"age":   {Type: genai.TypeInteger, Description: "user age"},
		},
	},
}

func (t *toolImpl) getUser(ctx context.Context, token *oauth2.Token, rawArgs json.RawMessage) (json.RawMessage, error) {
	var args GetUserArgs
	err := json.Unmarshal(rawArgs, &args)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal raw args: %w", err)
	}

	rows, err := t.bqRepo.GetUser(ctx, bqrepo.GetUserArgs{
		Token:   token,
		Columns: args.Columns,
		Emails:  args.Emails,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get rows: %w", err)
	}

	items := []GetUserItem{}
	for _, r := range rows {
		items = append(items, GetUserItem{
			Email: r.Email,
			Name:  r.Name,
			Age:   r.Age,
		})
	}

	marshaledItems, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal items: %w", err)
	}

	return marshaledItems, nil
}

func (t *toolImpl) GetUser() WiringItem {
	return WiringItem{
		Declaration: getUserDeclaration,
		Func:        t.getUser,
	}
}
